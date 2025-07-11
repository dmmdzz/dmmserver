// internal/services/banning/banning.go
package banning

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
	"github.com/gin-gonic/gin"
	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/utils"
)

var (
	ipBans                    sync.Map    // key: IP地址 (string), value: true (bool)
	deviceIDBans              sync.Map    // key: DeviceID (string), value: true (bool)
	realDeviceIDBans          sync.Map    // key: RealDeviceID (string), value: true (bool)
	deviceInfoBans            sync.Map    // key: DeviceInfo (string), value: true (bool)

	activityInWindow          atomic.Bool // 标记在当前窗口内是否有客户端活动
	lastRefreshTime           time.Time   // 记录上次成功刷新封禁列表的时间
	refreshMutex              sync.Mutex  // 互斥锁，确保同一时间只有一个协程在刷新列表
	noActivityStreakStartTime time.Time   // 记录无活动周期的开始时间，用于15分钟暂停逻辑
)

const (
	refreshInterval       = 55 * time.Second // 封禁列表刷新周期
	maxNoActivityDuration = 15 * time.Minute // 客户端连续无活动15分钟后，后台定时刷新将暂停
)

// Init 模块初始化函数，由bootstrap调用
func Init() {
	log.Println("Banning service is starting...")
	loadBansFromDB() // 首次启动时，强制加载一次封禁数据
	go startIntelligentTicker() // 启动后台智能刷新协程
	log.Println("Banning service started successfully.")
}

// loadBansFromDB 从数据库加载所有封禁列表到内存
func loadBansFromDB() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	// 重新加载IP封禁
	var ips []model.BanIP
	db.DB.Find(&ips) // 从数据库中查找所有IP封禁记录
	var newIpBans sync.Map // 创建一个新的map来存储数据
	for _, ban := range ips {
		newIpBans.Store(ban.IP, true) // 将IP作为key存入map
	}
	ipBans = newIpBans // 原子替换旧的map，确保数据一致性
	log.Printf("Loaded %d IP bans.", len(ips))

	// 重新加载DeviceID封禁
	var devices []model.BanDeviceID
	db.DB.Find(&devices)
	var newDeviceIDBans sync.Map
	for _, ban := range devices {
		newDeviceIDBans.Store(ban.DeviceID, true)
	}
	deviceIDBans = newDeviceIDBans
	log.Printf("Loaded %d DeviceID bans.", len(devices))

	// 重新加载RealDeviceID封禁 (假设RealDeviceID在model中也有对应表)
	var realDevices []model.BanRealDeviceID
	db.DB.Find(&realDevices)
	var newRealDeviceIDBans sync.Map
	for _, ban := range realDevices {
		newRealDeviceIDBans.Store(ban.RealDeviceID, true)
	}
	realDeviceIDBans = newRealDeviceIDBans
	log.Printf("Loaded %d RealDeviceID bans.", len(realDevices))
	
	// 重新加载DeviceInfo封禁
	var deviceInfos []model.BanDeviceInfo
	db.DB.Find(&deviceInfos)
	var newDeviceInfoBans sync.Map
	for _, ban := range deviceInfos {
		stableDeviceInfo := model.GetStableDeviceInfo(ban.DeviceInfo)
		newDeviceInfoBans.Store(stableDeviceInfo, true)
	}
	deviceInfoBans = newDeviceInfoBans
	log.Printf("Loaded %d DeviceInfo bans.", len(deviceInfos))

	lastRefreshTime = time.Now() // 更新上次刷新时间
}

// startIntelligentTicker 启动一个定时器，根据客户端活动情况智能刷新列表
func startIntelligentTicker() {
	ticker := time.NewTicker(refreshInterval) // 定时器周期调整为55秒
	defer ticker.Stop()

	for range ticker.C {
		if activityInWindow.Load() {
			activityInWindow.Store(false)       // 重置活动标记，为下一个窗口期做准备
			noActivityStreakStartTime = time.Time{} // 客户端有活动，重置无活动周期的开始时间
			log.Printf("Activity detected in last %s window. Refreshing ban lists.", refreshInterval)
			loadBansFromDB()
		} else {
			// 在过去的55秒窗口内没有客户端活动
			if noActivityStreakStartTime.IsZero() {
				// 这是第一个没有活动的55秒窗口，开始记录无活动周期
				// 无活动周期从这个刚结束的55秒窗口的开始算起
				noActivityStreakStartTime = time.Now().Add(-refreshInterval)
				log.Printf("No activity in last %s window. Starting %.0f-minute grace period for refreshes. Refreshing ban lists.", refreshInterval, maxNoActivityDuration.Minutes())
				loadBansFromDB()
			} else {
				// 已经处于无活动周期中
				if time.Since(noActivityStreakStartTime) < maxNoActivityDuration {
					// 仍在15分钟的宽限期内，继续刷新
					log.Printf("Continuing refresh during no-activity grace period (%.1f min into streak of %.1f min total). Refreshing ban lists.",
						time.Since(noActivityStreakStartTime).Minutes(), maxNoActivityDuration.Minutes())
					loadBansFromDB()
				} else {
					// 超过15分钟无活动，暂停定时刷新
					//log.Printf("No activity for over %.0f minutes (streak started at %s, current time %s, duration %.1f min). Pausing periodic ban list refresh.",
					//	maxNoActivityDuration.Minutes(), noActivityStreakStartTime.Format(time.RFC3339), time.Now().Format(time.RFC3339), time.Since(noActivityStreakStartTime).Minutes())
					// 此处不执行loadBansFromDB()，即暂停刷新
				}
			}
		}
	}
}

// NotifyActivity 由API中间件调用，标记发生了客户端活动
func NotifyActivity() {
	activityInWindow.Store(true) // 设置原子布尔值为true
	// 如果后台刷新因为长时间无活动而暂停，新的活动应该能使其在下一个ticker周期恢复。
	// CheckAndRefreshIfStale 也会在处理请求前确保数据不是太旧，并在必要时触发刷新和活动通知。
}

// CheckAndRefreshIfStale 是一个关键函数，在处理请求前调用
// 确保数据不是因为上一个周期无活动而变得“过时”
func CheckAndRefreshIfStale() {
	// 如果距离上次刷新已超过55秒（意味着上个ticker周期可能没刷新，或者后台刷新已因长时间无活动而暂停）
	if time.Since(lastRefreshTime) > refreshInterval {
		log.Printf("Ban list is stale (older than %s). Forcing an immediate refresh. Signalling activity to reset any inactivity streak.", refreshInterval)
		loadBansFromDB() // 执行刷新，这会更新lastRefreshTime
		// 标记发生了客户端活动。如果后台刷新因为15分钟无活动而暂停，
		// 这个标记会确保下一个ticker周期检测到活动，从而重置无活动计时器并恢复正常刷新。
		NotifyActivity() // 使用 NotifyActivity 来标记活动
	}
}

// CheckRequestIPBanned 检查请求IP是否被封禁，使用utils.GetClientIP获取IP
func CheckRequestIPBanned(c *gin.Context) bool {
	// 使用工具函数获取客户端IP
	ip := utils.GetClientIP(c)
	if ip == "" {
		return false // 无法获取IP，暂时放行
	}
	_, found := ipBans.Load(ip) // 在内存map中查找
	return found
}

// CheckDeviceIDBanned 检查DeviceID是否被封禁
func CheckDeviceIDBanned(deviceID string) bool {
	if deviceID == "" {
		return false
	}
	_, found := deviceIDBans.Load(deviceID)
	return found
}

// CheckRealDeviceIDBanned 检查RealDeviceID是否被封禁
func CheckRealDeviceIDBanned(realDeviceID string) bool {
	if realDeviceID == "" {
		return false
	}
	_, found := realDeviceIDBans.Load(realDeviceID)
	return found
}

// CheckDeviceInfoBanned 检查DeviceInfo是否被封禁
func CheckDeviceInfoBanned(deviceInfo string) bool {
	if deviceInfo == "" {
		return false
	}
	// 使用工具函数获取稳定的设备信息（移除变化的内存值部分）
	stableDeviceInfo := model.GetStableDeviceInfo(deviceInfo)
	var found bool
	deviceInfoBans.Range(func(key, value interface{}) bool {
		// 由于封禁列表中的deviceInfo也已经处理过，这里可以直接比较
		if key.(string) == stableDeviceInfo {
			found = true
			return false // 找到匹配项，停止遍历
		}
		return true // 继续遍历
	})
	return found
	return found
}
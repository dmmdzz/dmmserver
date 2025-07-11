// internal/services/playtime/playtime.go
package playtime

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
	"dmmserver/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 玩家游玩时长数据结构
type PlayerPlaytimeDB struct {
	PlayedTime      int64     // 已游玩时间（需要存入数据库）
	IsVIP           bool      // 是否为VIP玩家（需要存入数据库）
	DailyPlayTime   int64     // 每日可游玩时间，固定值，不允许重置（需要存入数据库）
	TodayExtraTime  int64     // 今日额外游玩时间，只限今天（需要存入数据库）
	LastUpdateTime  time.Time // 上次更新时间（需要存入数据库）
}

// 玩家游玩时长数据结构
type PlayerPlaytimeData struct {
	RemainingTime   int64     // 剩余游玩时间（计算得出，不存入数据库）
	LastLoginTime   time.Time // 上次登录时间（只存在内存中）
	DeviceID        string    // 设备ID（不需要存入数据库,只存在dmm_playerinfo数据表中）
	RealDeviceID    string    // 真实设备ID（不需要存入数据库,只存在dmm_playerinfo数据表中）
	IP              string    // IP地址（不需要存入数据库,只存在dmm_playerinfo数据表中）
}





// 游玩时长设置结构
type PlaytimeSettings struct {
	FreePlaytimeSeconds int64 `json:"freePlaytimeSeconds"` // 普通玩家每日免费游玩时长（秒）
	VIPPlaytimeSeconds  int64 `json:"vipPlaytimeSeconds"`  // VIP玩家每日游玩时长（秒）
	ResetHour           int   `json:"resetHour"`           // 每日重置时间（小时，UTC+8）
}

var (
	// 内存缓存
	playerPlaytimeCache sync.Map // key: deviceID (string), value: PlayerPlaytimeData
	playtimeSettings    atomic.Value // 存储当前的游玩时长设置
	allPlayerInfos      []model.PlayerInfo // 存储所有玩家的设备信息，用于多维度验证

	// 同步控制
	refreshMutex              sync.Mutex  // 互斥锁，确保同一时间只有一个协程在刷新数据
	activityInWindow          atomic.Bool // 标记在当前窗口内是否有客户端活动
	lastRefreshTime           time.Time   // 记录上次成功刷新数据的时间
	noActivityStreakStartTime time.Time   // 记录无活动周期的开始时间
	lastResetTime             time.Time   // 记录上次重置时间
)

const (
	refreshInterval       = 6 * time.Hour       // 数据刷新周期（6小时）
	maxNoActivityDuration = 15 * time.Minute     // 客户端连续无活动15分钟后，后台定时刷新将暂停
	defaultFreePlaytime   = 90 * 60             // 默认免费游玩时长（90分钟，即1.5小时）
	defaultVIPPlaytime    = 10 * 60 * 60        // 默认VIP游玩时长（10小时）
	defaultResetHour      = 1                   // 默认重置时间（凌晨1点，UTC+8）
)

// Init 模块初始化函数，由bootstrap调用
func Init() {
	log.Println("Playtime service is starting...")
	loadPlaytimeSettingsFromDB() // 首次启动时，强制加载一次设置数据
	loadPlaytimeDataFromDB()     // 首次启动时，强制加载一次玩家游玩时长数据
	loadAllPlayerInfos()         // 首次启动时，加载所有玩家的设备信息
	go startIntelligentTicker()  // 启动后台智能刷新协程
	go startDailyResetTicker()   // 启动每日重置协程
	log.Println("Playtime service started successfully.")
}

// loadPlaytimeSettingsFromDB 从数据库加载游玩时长设置到内存
func loadPlaytimeSettingsFromDB() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	var settings model.ServerSettings
	result := db.DB.First(&settings)

	if result.Error != nil {
		// 如果未找到记录或发生其他错误，使用默认设置
		log.Printf("加载游玩时长设置失败: %v，使用默认设置", result.Error)
		defaultSettings := PlaytimeSettings{
			FreePlaytimeSeconds: defaultFreePlaytime,
			VIPPlaytimeSeconds:  defaultVIPPlaytime,
			ResetHour:           defaultResetHour,
		}
		playtimeSettings.Store(defaultSettings)
		return
	}

	// 解析PlaytimeSettings字段
	var ptSettings PlaytimeSettings
	if settings.PlaytimeSettings == "" {
		// 如果字段为空，使用默认设置并保存到数据库
		ptSettings = PlaytimeSettings{
			FreePlaytimeSeconds: defaultFreePlaytime,
			VIPPlaytimeSeconds:  defaultVIPPlaytime,
			ResetHour:           defaultResetHour,
		}
		
		// 序列化设置
		settingsJSON, err := json.Marshal(ptSettings)
		if err != nil {
			log.Printf("序列化游玩时长设置失败: %v", err)
		} else {
			// 更新数据库
			settings.PlaytimeSettings = string(settingsJSON)
			db.DB.Save(&settings)
		}
	} else {
		// 解析现有设置
		err := json.Unmarshal([]byte(settings.PlaytimeSettings), &ptSettings)
		if err != nil {
			log.Printf("解析游玩时长设置失败: %v，使用默认设置", err)
			ptSettings = PlaytimeSettings{
				FreePlaytimeSeconds: defaultFreePlaytime,
				VIPPlaytimeSeconds:  defaultVIPPlaytime,
				ResetHour:           defaultResetHour,
			}
		}
	}

	// 存储到内存
	playtimeSettings.Store(ptSettings)
	log.Printf("游玩时长设置已加载: 免费时长=%d秒, VIP时长=%d秒, 重置时间=%d点", 
		ptSettings.FreePlaytimeSeconds, ptSettings.VIPPlaytimeSeconds, ptSettings.ResetHour)
}

// loadPlaytimeDataFromDB 从数据库加载所有玩家的游玩时长数据到内存
func loadPlaytimeDataFromDB() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	// 查询所有玩家数据
	var players []model.PlayerData
	result := db.DB.Find(&players)
	if result.Error != nil {
		log.Printf("加载玩家数据失败: %v", result.Error)
		return
	}

	// 获取当前设置
	ptSettings := playtimeSettings.Load().(PlaytimeSettings)

	// 处理每个玩家的数据
	for _, player := range players {
		// 解析玩家的游玩时长数据
		var playtimeData PlayerPlaytimeData
		
		// 检查是否有现有数据
		if player.PlaytimeData != "" {
			err := json.Unmarshal([]byte(player.PlaytimeData), &playtimeData)
			if err != nil {
				log.Printf("解析玩家 %s 的游玩时长数据失败: %v，使用默认值", player.DeviceID, err)
				// 使用默认值
				playtimeData = PlayerPlaytimeData{
					PlayedTime:     0,
					IsVIP:         false,
					DailyPlayTime: ptSettings.FreePlaytimeSeconds,
					TodayExtraTime: 0,
					LastUpdateTime: time.Now(),
					LastLoginTime:  time.Time{},
				}
			} else {
				// 确保新增字段有默认值
				if playtimeData.DailyPlayTime == 0 {
					if playtimeData.IsVIP {
						playtimeData.DailyPlayTime = ptSettings.VIPPlaytimeSeconds
					} else {
						playtimeData.DailyPlayTime = ptSettings.FreePlaytimeSeconds
					}
				}
			}
		} else {
			// 没有现有数据，使用默认值
			playtimeData = PlayerPlaytimeData{
				PlayedTime:     0,
				IsVIP:         false,
				DailyPlayTime: ptSettings.FreePlaytimeSeconds,
				TodayExtraTime: 0,
				LastUpdateTime: time.Now(),
				LastLoginTime:  time.Time{},
			}
		}

		// 存储到内存缓存
		playerPlaytimeCache.Store(player.DeviceID, playtimeData)
	}

	log.Printf("已加载 %d 个玩家的游玩时长数据到内存", len(players))
	lastRefreshTime = time.Now() // 更新上次刷新时间
}

// savePlaytimeDataToDB 将内存中的游玩时长数据保存到数据库
func savePlaytimeDataToDB() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	// 检查是否需要导入strings包
	var _ = strings.Split

	log.Println("开始将游玩时长数据保存到数据库...")
	
	// 统计计数器
	updateCount := 0
	errorCount := 0

	// 遍历内存缓存中的所有玩家数据
	playerPlaytimeCache.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		playtimeData := value.(PlayerPlaytimeData)

		// 创建只包含需要存储到数据库的字段的结构体
		dbPlaytimeData := struct {
			PlayedTime     int64     `json:"playedTime"`
			IsVIP          bool      `json:"isVIP"`
			DailyPlayTime  int64     `json:"dailyPlayTime"`
			TodayExtraTime int64     `json:"todayExtraTime"`
			LastUpdateTime time.Time `json:"lastUpdateTime"`
		}{
			PlayedTime:     playtimeData.PlayedTime,
			IsVIP:          playtimeData.IsVIP,
			DailyPlayTime:  playtimeData.DailyPlayTime,
			TodayExtraTime: playtimeData.TodayExtraTime,
			LastUpdateTime: playtimeData.LastUpdateTime,
		}
		
		// 序列化数据
		dataJSON, err := json.Marshal(dbPlaytimeData)
		if err != nil {
			log.Printf("序列化玩家 %s 的游玩时长数据失败: %v", deviceID, err)
			errorCount++
			return true // 继续遍历
		}

		// 更新数据库
		result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("playtime_data", string(dataJSON))
		if result.Error != nil {
			log.Printf("更新玩家 %s 的游玩时长数据失败: %v", deviceID, result.Error)
			errorCount++
		} else if result.RowsAffected > 0 {
			updateCount++
		}

		return true // 继续遍历
	})

	log.Printf("游玩时长数据保存完成: 更新 %d 条记录，失败 %d 条", updateCount, errorCount)
}

// startIntelligentTicker 启动一个定时器，根据客户端活动情况智能刷新数据
func startIntelligentTicker() {
	ticker := time.NewTicker(refreshInterval) // 定时器周期设置为6小时
	defer ticker.Stop()

	for range ticker.C {
		if activityInWindow.Load() {
			activityInWindow.Store(false)       // 重置活动标记，为下一个窗口期做准备
			noActivityStreakStartTime = time.Time{} // 客户端有活动，重置无活动周期的开始时间
			log.Printf("检测到过去 %s 窗口期内有客户端活动。刷新游玩时长数据。", refreshInterval)
			
			// 先保存数据到数据库，再重新加载
			savePlaytimeDataToDB()
			loadPlaytimeDataFromDB()
			loadPlaytimeSettingsFromDB()
			loadAllPlayerInfos() // 刷新玩家设备信息
		} else {
			// 如果当前窗口期内没有活动
			if noActivityStreakStartTime.IsZero() {
				// 这是第一个无活动的窗口期，记录开始时间
				noActivityStreakStartTime = time.Now()
				log.Printf("未检测到客户端活动。开始记录无活动周期。")
			} else if time.Since(noActivityStreakStartTime) >= maxNoActivityDuration {
				// 如果无活动时间超过阈值，暂停刷新
				log.Printf("连续 %s 无客户端活动。暂停游玩时长数据刷新。", time.Since(noActivityStreakStartTime))
				continue
			}

			// 即使没有活动，也执行一次刷新，但降低日志级别
			log.Printf("虽然未检测到客户端活动，但仍执行定期刷新。")
			savePlaytimeDataToDB()
			loadPlaytimeDataFromDB()
			loadPlaytimeSettingsFromDB()
			loadAllPlayerInfos() // 刷新玩家设备信息
		}
	}
}

// startDailyResetTicker 启动每日重置定时器
func startDailyResetTicker() {
	// 创建一个固定的UTC+8时区
	cst := time.FixedZone("CST", 8*60*60)
	
	// 计算下一次重置时间
	nextReset := calculateNextResetTime()
	log.Printf("下一次游玩时长重置时间: %v", nextReset.In(cst))

	for {
		// 计算等待时间
		waitDuration := time.Until(nextReset)
		log.Printf("距离下一次游玩时长重置还有: %v", waitDuration)

		// 等待到重置时间
		time.Sleep(waitDuration)

		// 执行重置
		resetAllPlaytimeData()

		// 计算下一次重置时间
		nextReset = calculateNextResetTime()
		log.Printf("游玩时长已重置，下一次重置时间: %v", nextReset.In(cst))
	}
}

// calculateNextResetTime 计算下一次重置时间（UTC+8时区的指定小时）
func calculateNextResetTime() time.Time {
	// 获取当前设置中的重置小时
	ptSettings := playtimeSettings.Load().(PlaytimeSettings)
	resetHour := ptSettings.ResetHour

	// 创建一个固定的UTC+8时区
	cst := time.FixedZone("CST", 8*60*60)

	// 获取当前时间（UTC）
	now := time.Now().UTC()

	// 计算今天的重置时间点（在UTC+8时区中的resetHour点）
	// 首先在UTC+8时区创建日期时间
	resetTodayInCST := time.Date(now.Year(), now.Month(), now.Day(), resetHour, 0, 0, 0, cst)
	// 然后转换回UTC时间
	resetToday := resetTodayInCST.UTC()

	// 如果当前时间已经过了今天的重置时间，则计算明天的重置时间
	if now.After(resetToday) {
		resetToday = resetToday.Add(24 * time.Hour)
	}

	return resetToday
}

// resetAllPlaytimeData 重置所有玩家的游玩时长数据
func resetAllPlaytimeData() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	log.Println("开始重置所有玩家的游玩时长数据...")

	// 获取当前设置
	ptSettings := playtimeSettings.Load().(PlaytimeSettings)

	// 遍历内存缓存中的所有玩家数据
	playerPlaytimeCache.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		playtimeData := value.(PlayerPlaytimeData)

		// 根据玩家VIP状态重置剩余时长
		if playtimeData.IsVIP {
			playtimeData.RemainingTime = ptSettings.VIPPlaytimeSeconds
			// DailyPlayTime是固定值，不允许重置
			if playtimeData.DailyPlayTime < ptSettings.VIPPlaytimeSeconds {
				playtimeData.DailyPlayTime = ptSettings.VIPPlaytimeSeconds
			}
		} else {
			playtimeData.RemainingTime = ptSettings.FreePlaytimeSeconds
			// DailyPlayTime是固定值，不允许重置
			if playtimeData.DailyPlayTime < ptSettings.FreePlaytimeSeconds {
				playtimeData.DailyPlayTime = ptSettings.FreePlaytimeSeconds
			}
		}
		
		// 重置今日额外时间（只限今天使用）
		playtimeData.TodayExtraTime = 0

		// 更新最后更新时间
		playtimeData.LastUpdateTime = time.Now()

		// 更新缓存
		playerPlaytimeCache.Store(deviceID, playtimeData)

		return true // 继续遍历
	})

	// 保存到数据库
	savePlaytimeDataToDB()

	// 更新上次重置时间
	lastResetTime = time.Now()
	log.Println("所有玩家的游玩时长数据已重置")
}

// CheckAndRefreshIfStale 检查设置是否过期，如果过期则刷新
func CheckAndRefreshIfStale() {
	// 如果上次刷新时间超过刷新间隔的两倍，则认为数据已过期
	if time.Since(lastRefreshTime) > 2*refreshInterval {
		log.Println("游玩时长数据已过期，正在刷新...")
		loadPlaytimeSettingsFromDB()
		loadPlaytimeDataFromDB()
	}
}

// NotifyActivity 通知服务有客户端活动，用于智能刷新机制
func NotifyActivity() {
	activityInWindow.Store(true)
}

// GetPlayerPlaytimeData 获取玩家的游玩时长数据，如果不存在则创建默认数据
func GetPlayerPlaytimeData(deviceID string) (PlayerPlaytimeData, error) {
	// 检查内存缓存
	if data, ok := playerPlaytimeCache.Load(deviceID); ok {
		// 动态计算剩余游玩时长
		data := data.(PlayerPlaytimeData)
		// 计算剩余时长 = 每日固定时长 + 今日额外时长
		data.RemainingTime = data.DailyPlayTime + data.TodayExtraTime
		return data, nil
	}

	// 如果内存中没有，尝试从数据库加载
	var player model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&player)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 玩家不存在，返回错误
			return PlayerPlaytimeData{}, result.Error
		}
		log.Printf("获取玩家 %s 数据失败: %v", deviceID, result.Error)
		return PlayerPlaytimeData{}, result.Error
	}

	// 解析玩家的游玩时长数据
	var playtimeData PlayerPlaytimeData
	if player.PlaytimeData != "" {
		err := json.Unmarshal([]byte(player.PlaytimeData), &playtimeData)
		if err != nil {
			log.Printf("解析玩家 %s 的游玩时长数据失败: %v，使用默认值", deviceID, err)
			// 使用默认值
			ptSettings := playtimeSettings.Load().(PlaytimeSettings)
			playtimeData = PlayerPlaytimeData{
				PlayedTime:     0,
				IsVIP:         false,
				DailyPlayTime: ptSettings.FreePlaytimeSeconds,
				TodayExtraTime: 0,
				LastUpdateTime: time.Now(),
				LastLoginTime:  time.Time{},
				DeviceID:      deviceID,
			}
		}
	} else {
		// 没有现有数据，使用默认值
		ptSettings := playtimeSettings.Load().(PlaytimeSettings)
		playtimeData = PlayerPlaytimeData{
			PlayedTime:     0,
			IsVIP:         false,
			DailyPlayTime: ptSettings.FreePlaytimeSeconds,
			TodayExtraTime: 0,
			LastUpdateTime: time.Now(),
			LastLoginTime:  time.Time{},
			DeviceID:      deviceID,
		}
	}

	// 查找PlayerInfo记录，获取设备信息
	var playerInfo model.PlayerInfo
	piResult := db.DB.Where("device_id = ?", deviceID).First(&playerInfo)
	if piResult.Error == nil {
		// 如果找到记录，获取最新的设备信息
		ipList := splitTextByDoubleNewline(playerInfo.IP)
		if len(ipList) > 0 {
			playtimeData.IP = ipList[len(ipList)-1] // 使用最新的IP
		}
		
		realDeviceIDList := splitTextByDoubleNewline(playerInfo.RealDeviceID)
		if len(realDeviceIDList) > 0 {
			playtimeData.RealDeviceID = realDeviceIDList[len(realDeviceIDList)-1] // 使用最新的realDeviceID
		}
	}

	// 动态计算剩余游玩时长
	playtimeData.RemainingTime = playtimeData.DailyPlayTime + playtimeData.TodayExtraTime

	// 存储到内存缓存
	playerPlaytimeCache.Store(deviceID, playtimeData)

	return playtimeData, nil
}

// UpdatePlayerPlaytimeData 更新玩家的游玩时长数据
func UpdatePlayerPlaytimeData(deviceID string, data PlayerPlaytimeData) error {
	// 更新内存缓存
	playerPlaytimeCache.Store(deviceID, data)

	// 如果距离上次数据库刷新时间较短，则不立即写入数据库
	// 依赖定期批量写入机制来减少数据库压力
	return nil
}

// SetPlayerVIPStatus 设置玩家的VIP状态
func SetPlayerVIPStatus(deviceID string, isVIP bool) error {
	// 获取玩家数据
	data, err := GetPlayerPlaytimeData(deviceID)
	if err != nil {
		return err
	}

	// 更新VIP状态
	data.IsVIP = isVIP

	// 如果设置为VIP，且剩余时间小于VIP时长，则更新剩余时间
	if isVIP {
		ptSettings := playtimeSettings.Load().(PlaytimeSettings)
		if data.RemainingTime < ptSettings.VIPPlaytimeSeconds {
			data.RemainingTime = ptSettings.VIPPlaytimeSeconds
		}
	}

	// 更新数据
	return UpdatePlayerPlaytimeData(deviceID, data)
}

// AddPlayerPlaytime 为玩家增加游玩时长（用于幸运玩家奖励）
// 此函数用于Telegram频道抽取的幸运玩家，增加的时长仅影响当日剩余时长，不影响次日重置
func AddPlayerPlaytime(deviceID string, additionalSeconds int64) error {
	// 获取玩家数据
	data, err := GetPlayerPlaytimeData(deviceID)
	if err != nil {
		return err
	}

	// 增加游玩时长
	data.RemainingTime += additionalSeconds

	// 记录日志
	log.Printf("为幸运玩家 %s 增加游玩时长 %d 秒，当前剩余时长: %d 秒", deviceID, additionalSeconds, data.RemainingTime)

	// 创建只包含需要存储到数据库的字段的结构体
	dbPlaytimeData := PlayerPlaytimeDB{
		PlayedTime:     data.PlayedTime,
		IsVIP:          data.IsVIP,
		DailyPlayTime:  data.DailyPlayTime,
		TodayExtraTime: data.TodayExtraTime,
		LastUpdateTime: data.LastUpdateTime,
	}

	// 立即更新数据库，确保数据不会丢失
	// 序列化数据
	dataJSON, err := json.Marshal(dbPlaytimeData)
	if err != nil {
		log.Printf("序列化幸运玩家 %s 的游玩时长数据失败: %v", deviceID, err)
		return err
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("playtime_data", string(dataJSON))
	if result.Error != nil {
		log.Printf("更新幸运玩家 %s 的游玩时长数据失败: %v", deviceID, result.Error)
		return result.Error
	}

	// 更新内存缓存
	playerPlaytimeCache.Store(deviceID, data)

	return nil
}

// ConsumePlaytime 消耗玩家游玩时长（登录时调用）
// 返回值：剩余时长（秒），是否允许登录，错误
func ConsumePlaytime(deviceID string) (int64, bool, error) {
	// 获取玩家数据
	data, err := GetPlayerPlaytimeData(deviceID)
	if err != nil {
		return 0, false, err
	}

	// 如果玩家有剩余时长，允许登录
	if data.RemainingTime > 0 {
		// 更新最后登录时间
		data.LastLoginTime = time.Now()
		
		// 更新数据
		err = UpdatePlayerPlaytimeData(deviceID, data)
		if err != nil {
			return 0, false, err
		}
		
		return data.RemainingTime, true, nil
	}

	// 没有剩余时长，不允许登录
	return 0, false, nil
}

// UpdatePlaytimeOnRequest 在请求处理期间更新玩家游玩时长
// 此函数应在请求中间件中调用，用于计算玩家游玩时长并更新剩余时间
func UpdatePlaytimeOnRequest(c *gin.Context, deviceID string) error {
	// 通知服务有客户端活动
	NotifyActivity()

	// 获取客户端IP和设备信息
	ip := utils.GetClientIP(c)
	realDeviceID, _ := c.GetPostForm("realDeviceID")

	// 获取玩家数据
	data, err := GetPlayerPlaytimeData(deviceID)
	if err != nil {
		return err
	}

	// 更新设备信息
	if ip != "" {
		data.IP = ip
	}
	if realDeviceID != "" {
		data.RealDeviceID = realDeviceID
	}


	// 如果玩家已登录，计算消耗的时长
	if !data.LastLoginTime.IsZero() {
		// 计算从上次登录到现在消耗的时间（秒）
		elapsed := time.Since(data.LastLoginTime).Seconds()
		
		// 优先消耗今日额外时长
		if data.TodayExtraTime > 0 {
			if data.TodayExtraTime >= int64(elapsed) {
				// 额外时长足够消耗
				data.TodayExtraTime -= int64(elapsed)
			} else {
				// 额外时长不够，消耗部分额外时长，剩余从固定时长中扣除
				remaining := int64(elapsed) - data.TodayExtraTime
				data.TodayExtraTime = 0
				data.DailyPlayTime -= remaining
				if data.DailyPlayTime < 0 {
					data.DailyPlayTime = 0
				}
			}
		} else {
			// 没有额外时长，直接从固定时长中扣除
			data.DailyPlayTime -= int64(elapsed)
			if data.DailyPlayTime < 0 {
				data.DailyPlayTime = 0
			}
		}
		
		// 更新最后登录时间
		data.LastLoginTime = time.Now()
		
		// 更新数据
		err = UpdatePlayerPlaytimeData(deviceID, data)
		if err != nil {
			return err
		}
		
		// 如果总剩余时长为0，返回错误
		if data.DailyPlayTime <= 0 && data.TodayExtraTime <= 0 {
			return game_error.New(-3001, "您的游玩时长已用完，请明天再来")
		}
	}

	// 更新PlayerInfo表中的设备信息
	var playerInfo model.PlayerInfo
	result := db.DB.Where("device_id = ?", deviceID).First(&playerInfo)
	if result.Error == nil {
		// 如果找到记录，更新设备信息
		if ip != "" {
			playerInfo.AddIP(ip)
		}
		if realDeviceID != "" {
			playerInfo.AddRealDeviceID(realDeviceID)
		}
		
		// 保存更新后的玩家信息
		db.DB.Save(&playerInfo)
	} else if result.Error == gorm.ErrRecordNotFound {
		// 如果记录不存在，创建新记录
		// 首先查询dmm_playerdata获取roleID
		var playerData model.PlayerData
		pdResult := db.DB.Where("device_id = ?", deviceID).First(&playerData)
		if pdResult.Error == nil {
			// 创建新记录
			playerInfo = model.PlayerInfo{
				DeviceID: deviceID,
				RoleID:   playerData.RoleID,
			}
			
			// 添加初始数据
			if ip != "" {
				playerInfo.IP = ip
			}
			if realDeviceID != "" {
				playerInfo.RealDeviceID = realDeviceID
			}
			
			// 保存新记录
			db.DB.Create(&playerInfo)
		}
	}

	return nil
}

// AddLuckyPlayersPlaytime 为多个幸运玩家批量增加游玩时长
// 参数：luckyPlayers - 幸运玩家的deviceID列表，additionalSeconds - 要增加的时长（秒）
// 返回：成功增加时长的玩家数量，错误信息
func AddLuckyPlayersPlaytime(luckyPlayers []string, additionalSeconds int64) (int, error) {
	successCount := 0
	var lastError error

	// 遍历所有幸运玩家
	for _, deviceID := range luckyPlayers {
		// 为每个玩家增加游玩时长
		err := AddPlayerPlaytime(deviceID, additionalSeconds)
		if err != nil {
			// 记录错误，但继续处理其他玩家
			log.Printf("为幸运玩家 %s 增加游玩时长失败: %v", deviceID, err)
			lastError = err
			continue
		}

		// 增加成功计数
		successCount++
	}

	// 记录总体结果
	log.Printf("批量增加幸运玩家游玩时长完成: 成功 %d/%d 个玩家", successCount, len(luckyPlayers))

	// 如果有任何玩家处理失败，返回最后一个错误
	if successCount < len(luckyPlayers) {
		return successCount, lastError
	}

	return successCount, nil
}

// splitTextByDoubleNewline 将双换行符分隔的文本分割为字符串数组
// 此函数是model包中同名函数的本地实现，避免循环依赖
func splitTextByDoubleNewline(text string) []string {
	if text == "" {
		return []string{}
	}
	
	// 使用双换行符分割文本
	var result []string
	parts := strings.Split(text, "\n\n")
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	
	return result
}

// loadAllPlayerInfos 从数据库加载所有玩家的设备信息到内存
func loadAllPlayerInfos() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	// 查询所有玩家信息
	var playerInfos []model.PlayerInfo
	result := db.DB.Find(&playerInfos)
	if result.Error != nil {
		log.Printf("加载玩家设备信息失败: %v", result.Error)
		return
	}

	// 更新全局变量
	allPlayerInfos = playerInfos

	log.Printf("已加载 %d 个玩家的设备信息到内存", len(playerInfos))
}

// VerifyPlaytimeMultiDimension 多维度验证玩家身份并检查游玩时长
// 此函数应在登录处理中调用，用于防止玩家通过切换设备规避时长限制
func VerifyPlaytimeMultiDimension(c *gin.Context, deviceID string) error {
	// 通知服务有客户端活动
	NotifyActivity()

	// 获取客户端IP
	ip := utils.GetClientIP(c)
	
	// 获取realDeviceID
	realDeviceID, _ := c.GetPostForm("realDeviceID")

	// 查找PlayerInfo记录
	var playerInfo model.PlayerInfo
	result := db.DB.Where("device_id = ?", deviceID).First(&playerInfo)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Printf("查询玩家信息失败: %v", result.Error)
		return result.Error
	}

	// 多维度验证评分系统
	// 1. 如果使用相同的deviceID，直接通过验证
	// 2. 如果使用不同的deviceID，则根据IP、realDeviceID的匹配程度进行评分
	// 3. 评分达到阈值，则认为是同一玩家，共享游玩时长限制
	
	// 定义权重
	const (
		ipWeight = 35         // IP匹配权重
		realDeviceIDWeight = 40 // realDeviceID匹配权重
		deviceIDWeight = 25   // deviceID匹配权重
		scoreThreshold = 60     // 评分阈值，超过此值则认为是同一玩家
	)
	
	// 如果找到记录，检查IP、realDeviceID和deviceID是否匹配
	if result.Error == nil {
		// 初始化评分
		score := 0
		
		// 检查IP
		ipList := splitTextByDoubleNewline(playerInfo.IP)
		ipFound := false
		for _, knownIP := range ipList {
			if knownIP == ip {
				ipFound = true
				score += ipWeight
				break
			}
		}
		
		// 检查realDeviceID
		realDeviceIDList := splitTextByDoubleNewline(playerInfo.RealDeviceID)
		realDeviceIDFound := false
		for _, knownRealDeviceID := range realDeviceIDList {
			if knownRealDeviceID == realDeviceID {
				realDeviceIDFound = true
				score += realDeviceIDWeight
				break
			}
		}
		
		// 检查deviceID（当前设备与记录中的设备ID相同，直接加上权重分）
		deviceIDFound := true // 当前函数是按deviceID查询的，所以一定匹配
		score += deviceIDWeight
		
		// 记录验证结果
		log.Printf("多维度验证结果 - deviceID: %s, IP匹配: %v, realDeviceID匹配: %v, deviceID匹配: %v, 总评分: %d", 
			deviceID, ipFound, realDeviceIDFound, deviceIDFound, score)
		
		// 如果评分达到阈值，则认为是同一玩家，共享游玩时长限制
		if score >= scoreThreshold {
			// 获取玩家游玩时长数据
			data, err := GetPlayerPlaytimeData(deviceID)
			if err != nil {
				return err
			}
			
			// 如果没有剩余时长，返回错误
			if data.RemainingTime <= 0 {
				return game_error.New(-3001, "您的游玩时长已用完，请明天再来")
			}
			
			// 更新最后登录时间
			data.LastLoginTime = time.Now()
			
			// 更新数据
			err = UpdatePlayerPlaytimeData(deviceID, data)
			if err != nil {
				return err
			}
		} else {
			// 评分未达到阈值，尝试查找可能的关联账号
			// 遍历所有玩家信息，查找IP或realDeviceID匹配的记录
			relatedDeviceIDs := make([]string, 0) // 存储所有可能关联的设备ID
			relationScores := make(map[string]int) // 存储每个设备ID的关联评分
			
			// 第一轮：查找所有可能关联的账号
			for _, info := range allPlayerInfos {
				// 跳过当前设备ID
				if info.DeviceID == deviceID {
					continue
				}
				
				// 初始化关联评分
				relationScore := 0
				
				// 检查IP匹配
				ipList := splitTextByDoubleNewline(info.IP)
				for _, knownIP := range ipList {
					if knownIP == ip && ip != "" {
						relationScore += ipWeight
						break
					}
				}
				
				// 检查realDeviceID匹配
				realDeviceIDList := splitTextByDoubleNewline(info.RealDeviceID)
				for _, knownRealDeviceID := range realDeviceIDList {
					if knownRealDeviceID == realDeviceID && realDeviceID != "" {
						relationScore += realDeviceIDWeight
						break
					}
				}
				
				// 如果关联评分达到阈值，记录该账号
				if relationScore >= scoreThreshold - deviceIDWeight {
					log.Printf("发现可能的关联账号 - 当前deviceID: %s, 关联deviceID: %s, 关联评分: %d", 
						deviceID, info.DeviceID, relationScore)
					
					relatedDeviceIDs = append(relatedDeviceIDs, info.DeviceID)
					relationScores[info.DeviceID] = relationScore
				}
			}
			
			// 第二轮：检查所有关联账号的游玩时长
			for _, relatedDeviceID := range relatedDeviceIDs {
				// 获取关联账号的游玩时长数据
				relatedData, err := GetPlayerPlaytimeData(relatedDeviceID)
				if err != nil {
					// 如果获取失败，继续检查下一个账号
					continue
				}
				
				// 如果关联账号没有剩余时长，返回错误
				if relatedData.RemainingTime <= 0 {
					log.Printf("关联账号 %s 游玩时长已用完，阻止当前账号 %s 登录", relatedDeviceID, deviceID)
					return game_error.New(-3001, "您的游玩时长已用完，请明天再来")
				}
				
				// 如果关联账号有剩余时长但很少（小于10分钟），记录警告日志
				if relatedData.RemainingTime < 600 { // 10分钟 = 600秒
					log.Printf("警告：关联账号 %s 游玩时长即将用完（剩余 %d 秒），当前账号 %s 可能尝试规避时长限制", 
						relatedDeviceID, relatedData.RemainingTime, deviceID)
				}
			}
			
			// 如果没有找到关联账号或所有关联账号都有足够的游玩时长，则允许登录
			// 但需要更新当前账号的最后登录时间
			data, err := GetPlayerPlaytimeData(deviceID)
			if err != nil {
				return err
			}
			
			// 如果当前账号没有剩余时长，返回错误
			if data.RemainingTime <= 0 {
				return game_error.New(-3001, "您的游玩时长已用完，请明天再来")
			}
			
			// 更新最后登录时间
			data.LastLoginTime = time.Now()
			
			// 更新数据
			err = UpdatePlayerPlaytimeData(deviceID, data)
			if err != nil {
				return err
			}
		}
	} else {
		// 如果没有找到PlayerInfo记录，说明是新玩家
		// 创建新的PlayerInfo记录
		if ip != "" || realDeviceID != "" {
			// 首先查询dmm_playerdata获取roleID
			var playerData model.PlayerData
			pdResult := db.DB.Where("device_id = ?", deviceID).First(&playerData)
			if pdResult.Error == nil {
				// 创建新记录
				playerInfo = model.PlayerInfo{
					DeviceID: deviceID,
					RoleID:   playerData.RoleID,
				}
				
				// 添加初始数据
				if ip != "" {
					playerInfo.IP = ip
				}
				if realDeviceID != "" {
					playerInfo.RealDeviceID = realDeviceID
				}
				
				// 保存新记录
				db.DB.Create(&playerInfo)
			}
		}
		
		// 获取玩家游玩时长数据
		data, err := GetPlayerPlaytimeData(deviceID)
		if err != nil {
			return err
		}
		
		// 如果没有剩余时长，返回错误
		if data.RemainingTime <= 0 {
			return game_error.New(-3001, "您的游玩时长已用完，请明天再来")
		}
		
		// 更新最后登录时间
		data.LastLoginTime = time.Now()
		
		// 更新数据
		err = UpdatePlayerPlaytimeData(deviceID, data)
		if err != nil {
			return err
		}
	}
	
	return nil
}
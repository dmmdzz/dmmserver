// internal/services/serversettings/serversettings.go
package serversettings

import (
	"encoding/json"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"dmmserver/db"
	"dmmserver/model"

    "gorm.io/gorm"
)

var (
	currentSettings		atomic.Value	// 存储当前的ServerSettings
	refreshMutex		sync.Mutex	// 互斥锁，确保同一时间只有一个协程在刷新设置
	activityInWindow	atomic.Bool	// 标记在当前窗口内是否有客户端活动
	lastRefreshTime		time.Time	// 记录上次成功刷新设置的时间
	noActivityStreakStartTime time.Time	// 记录无活动周期的开始时间
)

const (
	refreshInterval       = 55 * time.Second // 设置刷新周期
	maxNoActivityDuration = 15 * time.Minute // 客户端连续无活动15分钟后，后台定时刷新将暂停
)

// Init 模块初始化函数，由bootstrap调用
func Init() {
	log.Println("ServerSettings service is starting...")
	loadSettingsFromDB() // 首次启动时，强制加载一次设置数据
	log.Printf("Initial settings loaded: GraphicsOptions length=%d, MiscOptions length=%d", len(currentSettings.Load().(model.ServerSettings).GraphicsOptions), len(currentSettings.Load().(model.ServerSettings).MiscOptions))
	go startIntelligentTicker() // 启动后台智能刷新协程
	log.Println("ServerSettings service started successfully.")
}

// loadSettingsFromDB 从数据库加载ServerSettings到内存
func loadSettingsFromDB() {
	refreshMutex.Lock() // 获取锁，防止并发刷新
	defer refreshMutex.Unlock()

	var settings model.ServerSettings
	result := db.DB.First(&settings)

	if result.Error != nil {
		// 如果未找到记录，创建默认设置
		if result.Error == gorm.ErrRecordNotFound {
			log.Println("未找到服务器设置。正在创建默认设置。")

			// 使用model中定义的默认值
			defaultGraphicsOptions := model.DefaultGraphicsOptions()
			defaultMiscOptions := model.DefaultMiscOptions()

			settings = model.ServerSettings{
				GraphicsOptions: defaultGraphicsOptions,
				MiscOptions:     defaultMiscOptions,
			}

			createResult := db.DB.Create(&settings)
			if createResult.Error != nil {
				log.Printf("创建默认服务器设置失败: %v", createResult.Error)
				// 即使创建失败，也尝试使用默认值填充缓存，避免服务中断
				settings.GraphicsOptions = defaultGraphicsOptions
				settings.MiscOptions = defaultMiscOptions
			}
		} else {
			// 其他数据库错误，记录错误但不中断服务，使用旧缓存或空值
			log.Printf("获取服务器设置失败: %v", result.Error)
			// 不更新缓存，继续使用旧数据或保持为空
			return
		}
	}

	// 成功获取或创建设置，更新缓存
	currentSettings.Store(settings)
	lastRefreshTime = time.Now() // 更新上次刷新时间
	t := reflect.TypeOf(settings)
	fieldCount := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name != "CreatedAt" && field.Name != "UpdatedAt" {
			fieldCount++
		}
	}
	log.Printf("ServerSettings loaded and cached successfully. Fields: GraphicsOptions(%d chars), MiscOptions(%d chars)", len(settings.GraphicsOptions), len(settings.MiscOptions))
}

// startIntelligentTicker 启动一个定时器，根据客户端活动情况智能刷新设置
func startIntelligentTicker() {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		if activityInWindow.Load() {
			activityInWindow.Store(false)
			noActivityStreakStartTime = time.Time{}
			activityInWindow.Store(false)
			log.Printf("活动检测窗口 %s 内有客户端活动，立即刷新服务器设置", refreshInterval)
			loadSettingsFromDB()
		} else {
			currentTime := time.Now()
			if noActivityStreakStartTime.IsZero() {
				noActivityStreakStartTime = time.Now().Add(-refreshInterval)
				log.Printf("开始 %.0f 分钟无活动宽限期，当前时间: %s", maxNoActivityDuration.Minutes(), currentTime.Format(time.RFC3339))
				loadSettingsFromDB()
			} else {
				if currentTime.Sub(noActivityStreakStartTime) < maxNoActivityDuration {
					log.Printf("宽限期剩余时间 %.1f 分钟，继续刷新服务器设置", (maxNoActivityDuration - currentTime.Sub(noActivityStreakStartTime)).Minutes())
					loadSettingsFromDB()
				} else {
					//log.Printf("已超过 %.0f 分钟无活动 (开始于 %s)，暂停定时刷新", maxNoActivityDuration.Minutes(), noActivityStreakStartTime.Format(time.RFC3339))
				}
			}
		}
	}
}

// GetSettings 获取当前缓存的ServerSettings
func GetSettings() model.ServerSettings {
	settings, ok := currentSettings.Load().(model.ServerSettings)
	if !ok {
		// 如果缓存为空或类型错误，尝试加载一次
		log.Println("ServerSettings缓存为空或类型错误，尝试立即加载...")
		loadSettingsFromDB()
		settings, ok = currentSettings.Load().(model.ServerSettings)
		if !ok {
			// 如果加载后仍然失败，返回一个带有默认JSON字符串的空结构体
			log.Println("立即加载ServerSettings失败，返回默认空设置。")
			return model.ServerSettings{
				GraphicsOptions: model.DefaultGraphicsOptions(),
				MiscOptions: model.DefaultMiscOptions(),
			}
		}
	}
	return settings
}

// NotifyActivity 由API中间件或相关逻辑调用，标记发生了客户端活动
func NotifyActivity() {
	activityInWindow.Store(true) // 设置原子布尔值为true
	// 如果后台刷新因为长时间无活动而暂停，新的活动应该能使其在下一个ticker周期恢复。
	// CheckAndRefreshIfStale 也会在处理请求前确保数据不是太旧，并在必要时触发刷新和活动通知。
}

// CheckAndRefreshIfStale 是一个关键函数，在处理请求前调用
// 确保数据不是因为上一个周期无活动而变得“过时”
func CheckAndRefreshIfStale() {
	// 如果距离上次刷新已超过55秒 并且 无活动时间已超过15分钟，则强制刷新
	if time.Since(lastRefreshTime) > refreshInterval && time.Since(noActivityStreakStartTime) > maxNoActivityDuration {
		log.Printf("服务器设置已过期(%s)且无活动时间超过%.0f分钟，强制刷新并重置计时器", refreshInterval, maxNoActivityDuration.Minutes())
		loadSettingsFromDB() // 执行刷新，这会更新lastRefreshTime
		// 重置无活动周期计时
		noActivityStreakStartTime = time.Time{}
		activityInWindow.Store(false)
		// NotifyActivity() // 在这里不需要调用NotifyActivity，因为强制刷新本身就是响应了新的活动请求
	}
}

// ParseGraphicsOptions 解析GraphicsOptions JSON字符串
func ParseGraphicsOptions(jsonStr string) []map[string]interface{} {
	var options []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &options); err != nil {
		log.Printf("解析GraphicsOptions失败: %v，使用默认值", err)
		// 返回model中定义的默认值
		var defaultOptions []map[string]interface{}
		json.Unmarshal([]byte(model.DefaultGraphicsOptions()), &defaultOptions)
		return defaultOptions
	}
	return options
}

// ParseMiscOptions 解析MiscOptions JSON字符串
func ParseMiscOptions(jsonStr string) map[string]interface{} {
	var options map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &options); err != nil {
		log.Printf("解析MiscOptions失败: %v，使用默认值", err)
		// 返回model中定义的默认值
		var defaultOptions map[string]interface{}
		json.Unmarshal([]byte(model.DefaultMiscOptions()), &defaultOptions)
		return defaultOptions
	}
	return options
}
// internal/server/server.go
package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv" // 用于将字符串msg_id转换为整数
	"dmmserver/conf"
	"dmmserver/game_error"
	"dmmserver/handler"
	"dmmserver/services/banning"
//	"dmmserver/services/playtime"
	"dmmserver/services/serversettings"
	"dmmserver/utils"
	"github.com/gin-gonic/gin"
)

// activityMiddleware 负责处理与刷新机制相关的逻辑
func activityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		banning.NotifyActivity()
		banning.CheckAndRefreshIfStale()
		serversettings.NotifyActivity()
		serversettings.CheckAndRefreshIfStale()
		c.Next()
	}
}

// banningEnforcementMiddleware 封禁强制执行中间件
// 它会在任何业务Handler执行之前运行，并根据配置返回特定格式的响应。
func banningEnforcementMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("banningEnforcementMiddleware: Request Content-Type: %s", c.Request.Header.Get("Content-Type")) // 添加日志打印Content-Type
		msgIDStr := c.PostForm("msg_id")
		msgStr := c.PostForm("msg")
		
		var msgData map[string]interface{}
		if msgStr != "" {
			// 这里对msg的解析错误暂时忽略，后续handler如果需要会再次校验
			// 主要是为了从msg中提取 deviceID 等信息进行封禁检查
			_ = json.Unmarshal([]byte(msgStr), &msgData) 
		} else {
			msgData = make(map[string]interface{})
		}
		
		// 将解析后的数据存入上下文，供后续的 dispatchHandler 使用
		log.Printf("banningEnforcementMiddleware: Received msg_id from form: '%s'", msgIDStr) // 添加日志
		c.Set("msg_id", msgIDStr)
		c.Set("msg_data", msgData)
		
		// 1. 执行封禁检查 - 先检查封禁，如果被封禁则立即返回错误，不执行后续逻辑
		var isBanned bool
		
		// IP 检查 
		if banning.CheckRequestIPBanned(c) {
			isBanned = true
		}
		
		// DeviceID 检查 (假设DeviceID在msgData中)
		if !isBanned && msgData != nil {
			if deviceID, ok := msgData["deviceID"].(string); ok { 
				if banning.CheckDeviceIDBanned(deviceID) {
					isBanned = true
				}
			}
		}
		
		// RealDeviceID 检查 (假设RealDeviceID在msgData中)
		if !isBanned && msgData != nil {
			if realDeviceID, ok := msgData["realDeviceID"].(string); ok { 
				if banning.CheckRealDeviceIDBanned(realDeviceID) {
					isBanned = true
				}
			}
		}
		
		// DeviceInfo 检查 (假设DeviceInfo在msgData中)
		if !isBanned && msgData != nil {
			if deviceInfo, ok := msgData["deviceInfo"].(string); ok { 
				if banning.CheckDeviceInfoBanned(deviceInfo) {
					isBanned = true
				}
			}
		}
		
		// 2. 如果命中封禁，立即返回错误，不执行后续逻辑
		if isBanned {
			log.Printf("Request with msg_id [%s] blocked by ban policy.", msgIDStr);

			msgIDInt, err := strconv.Atoi(msgIDStr)
			if err != nil {
				// 如果 msg_id 不是有效的整数，无法查找其响应格式，则返回默认JSON错误
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"errorCode": -65, // 服务器逻辑错误
					"errorMsg":  "Invalid msg_id format for ban response lookup",
				})
				return
			}

			// 根据 response_format.go 中的配置来判断响应格式
			if conf.IsTextResponse(msgIDInt) {
				// 发送纯文本响应
				c.AbortWithStatus(http.StatusOK)
				// 这里的文本内容也可以配置化，但为了简洁，先硬编码一个通用文本
				c.String(http.StatusOK, "获取版本信息失败，请重试") //虚假的错误提示信息误导黑客
			} else {
				// 发送JSON响应 (默认行为)
				// 封禁的JSON错误码直接使用 -3001
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"errorCode": -3001,
					"errorMsg":  "获取版本信息失败，请重试",//虚假的错误提示信息误导黑客
				})
			}
			return // 确保请求中止，不再传递给后续处理器
		}
		
		// 3. 如果未被封禁，再记录客户端IP地址
		// 如果msgData中包含deviceID，则记录IP
		if deviceID, ok := msgData["deviceID"].(string); ok {
			err := utils.RecordClientIP(c, deviceID)
			if err != nil {
				log.Printf("记录客户端IP地址失败: %v", err)
				
				// 检查是否是GameError类型的错误，如果是则直接返回对应的错误码
				if gameErr, ok := err.(*game_error.GameError); ok {
					// 根据 response_format.go 中的配置来判断响应格式
					msgIDInt, convErr := strconv.Atoi(msgIDStr)
					if convErr != nil {
						// 如果 msg_id 不是有效的整数，则返回默认JSON错误
						c.AbortWithStatusJSON(http.StatusOK, gin.H{
							"errorCode": gameErr.Code,
							"errorMsg":  gameErr.Message,
						})
					} else if conf.IsTextResponse(msgIDInt) {
						// 发送纯文本响应
						c.AbortWithStatus(http.StatusOK)
						c.String(http.StatusOK, gameErr.Message)
					} else {
						// 发送JSON响应
						c.AbortWithStatusJSON(http.StatusOK, gin.H{
							"errorCode": gameErr.Code,
							"errorMsg":  gameErr.Message,
						})
					}
					return // 确保请求中止，不再传递给后续处理器
				}
			}
		}



		// 3. 如果所有检查都通过，放行请求到下一个中间件或dispatchHandler
		c.Next()
	}
}

// dispatchHandler 现在是唯一的“发送网关”，它统一处理业务逻辑的返回和发送
func dispatchHandler(c *gin.Context) {
	// 直接从上下文中获取由中间件解析好的数据
	msgIDVal, exists := c.Get("msg_id")
	if !exists {
		// 理论上不会发生，因为 banningEnforcementMiddleware 已经设置了
		c.JSON(http.StatusOK, gin.H{"errorCode": -5}) // 参数丢失
		return
	}
	msgID := msgIDVal.(string)

	msgDataVal, exists := c.Get("msg_data")
	if !exists {
		// 理论上不会发生
		msgDataVal = make(map[string]interface{})
	}
	msgData := msgDataVal.(map[string]interface{})

	log.Printf("dispatchHandler: Retrieved msg_id from context: '%s'", msgID) // 添加日志

	// 从注册表中查找对应的业务处理器
	h, found := handler.GetHandler(msgID)
	log.Printf("dispatchHandler: Handler lookup for msg_id '%s'. Found: %t", msgID, found) // 添加日志
	if !found {
		// 如果 msg_id 对应的处理器不存在
		c.JSON(http.StatusOK, gin.H{"errorCode": -65, "errorMsg": "handler not found"}) // 服务器逻辑错误
		return
	}

	// 调用业务 handler，获取返回的数据和错误
	data, err := h(c, msgData)

	// 处理业务 handler 返回的错误
	if err != nil {
		if gameErr, ok := err.(*game_error.GameError); ok {
			// 如果是业务逻辑返回的 GameError
			c.JSON(http.StatusOK, gin.H{"errorCode": gameErr.Code})
		} else {
			// 如果是其他未知错误 (例如数据库连接错误等)，记录日志并返回一个通用的服务器逻辑错误
			log.Printf("Unhandled internal error for msg_id %s: %v", msgID, err)
			c.JSON(http.StatusOK, gin.H{"errorCode": -65})
		}
		return
	}

	// 处理业务成功的情况
	response := gin.H{}
	if data != nil {
		// 将 handler 返回的业务数据合并到最终响应中
		for key, value := range data {
			response[key] = value
		}
	}

	// 无论 handler 是否返回额外数据，都统一添加成功的 errorCode
	response["errorCode"] = 0 
	c.JSON(http.StatusOK, response)
}

func Run() {
	r := gin.Default()

	// 创建一个 API 组，并应用中间件
	apiGroup := r.Group("/")
	// 顺序很重要：
	// 1. activityMiddleware 负责更新活动时间戳和强制刷新
	// 2. banningEnforcementMiddleware 负责拦截封禁请求
	// 3. PlaytimeMiddleware 负责检查玩家游戏时长限制
	// 4. dispatchHandler 负责业务分发和统一响应
	//apiGroup.Use(activityMiddleware(), banningEnforcementMiddleware(), PlaytimeMiddleware())
	apiGroup.Use(activityMiddleware(), banningEnforcementMiddleware())
	{
		// 注册唯一的请求处理路由
		apiGroup.POST("/", dispatchHandler)
	}

	port := ":" + conf.Conf.Server.Port
	log.Printf("Server is starting, listening on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
// internal/handler/30065.go
package handler

import (
	"dmmserver/utils"
	"log"
	"time"
	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() {
	Register("30065", handle30065)
}

// handle30065 处理设备信息更新请求
func handle30065(c *gin.Context, msgData map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing handler for msg_id=30065. Received msgData: %+v", msgData)

	// 1. 参数解析和验证
	// ----------------------
	// 验证必须的参数字段
	requiredFields := []string{"bundleIdentifier", "deviceInfo", "realDeviceID", "authKey", "accountName", "roleID", "pfID", "deviceID", "version", "baseVerCode", "compVerCode", "sv", "sequenceID"}
	for _, field := range requiredFields {
		if _, exists := msgData[field]; !exists {
			log.Printf("错误：msg_id=30065 缺少 '%s' 参数", field)
			return nil, game_error.New(-5, "参数丢失，请重新登录")
		}
	}

	deviceID, ok := msgData["deviceID"].(string)
	if !ok || deviceID == "" {
		log.Println("错误：msg_id=30065 缺少 'deviceID' 参数")
		return nil, game_error.New(-5, "缺少 'deviceID' 参数")
	}

	deviceInfo, ok := msgData["deviceInfo"].(string)
	if !ok || deviceInfo == "" {
		log.Println("错误：msg_id=30065 缺少 'deviceInfo' 参数")
		return nil, game_error.New(-5, "缺少 'deviceInfo' 参数")
	}

	authKey, ok := msgData["authKey"].(string)
	if !ok || authKey == "" {
		log.Println("错误：msg_id=30065 缺少 'authKey' 参数")
		return nil, game_error.New(-5, "缺少 'authKey' 参数")
	}
	
	// 获取accountName和roleID
	accountName, ok := msgData["accountName"].(string)
	if !ok || accountName == "" {
		log.Println("错误：msg_id=30065 缺少 'accountName' 参数")
		return nil, game_error.New(-5, "缺少 'accountName' 参数")
	}
	
	// 2. 业务逻辑
	// -------------------------------------------------------------
	// 根据 deviceID 查询 dmm_playerdata
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		log.Printf("未找到 deviceID 为 '%s' 的玩家", deviceID)
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 3. 验证 authKey 是否过期且匹配
	currentTime := time.Now().Unix()
	if playerData.AuthKeyExpire < currentTime {
		log.Printf("authKey 已过期，过期时间: %d, 当前时间: %d", playerData.AuthKeyExpire, currentTime)
		return nil, game_error.New(-12, "登录秘钥失效，请重新登录")
	}

	if playerData.AuthKey != authKey {
		log.Printf("authKey 不匹配，请求的 authKey: %s, 数据库中的 authKey: %s", authKey, playerData.AuthKey)
		return nil, game_error.New(-11, "登录验证错误，账号或已在别处登录")
	}
	
	// 验证accountName和roleID是否与数据库中的匹配
	// 使用PublicInfoManager获取玩家名字
	pm := utils.NewPublicInfoManager()
	publicInfo, err := pm.GetPublicInfo(deviceID)
	if err != nil {
		log.Printf("获取玩家公开信息失败: %v", err)
		return nil, game_error.New(-3, "未找到玩家数据")
	}
	
	if publicInfo.Name != accountName {
		log.Printf("accountName 不匹配，请求的 accountName: %s, 数据库中的 accountName: %s", accountName, publicInfo.Name)
		return nil, game_error.New(-13, "非法参数")
	}

	// 4. 处理设备信息更新
	// 从请求中获取realDeviceID
	realDeviceID, ok := msgData["realDeviceID"].(string)
	if !ok || realDeviceID == "" {
		log.Println("错误：msg_id=30065 缺少 'realDeviceID' 参数")
		return nil, game_error.New(-5, "缺少 'realDeviceID' 参数")
	}

	// 在dmm_playerinfo表中查找对应的设备信息
	var playerInfo model.PlayerInfo
	result = db.DB.Where("device_id = ?", deviceID).First(&playerInfo)

	// 如果在dmm_playerinfo表中找不到记录，则创建新记录
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 首先检查roleID是否已存在于dmm_playerinfo表中
			var existingPlayerInfo model.PlayerInfo
			checkResult := db.DB.Where("role_id = ?", playerData.RoleID).First(&existingPlayerInfo)
			if checkResult.Error == nil {
				// 如果找到了记录，说明roleID已存在，返回错误
				log.Printf("roleID %d 已存在于dmm_playerinfo表中，deviceID: %s", playerData.RoleID, existingPlayerInfo.DeviceID)
				return nil, game_error.New(-15, "账号数据异常，请联系客服")
			} else if checkResult.Error != gorm.ErrRecordNotFound {
				// 如果是其他数据库错误
				log.Printf("检查roleID唯一性失败: %v", checkResult.Error)
				return nil, game_error.New(-2, "数据库查询错误")
			}
			
			// 创建新的PlayerInfo记录
			playerInfo = model.PlayerInfo{
				RoleID:   playerData.RoleID, // 使用从playerData查询到的roleID
				DeviceID: deviceID,
			}

			// 初始化JSON数组并添加第一条设备信息
			if err := playerInfo.AddRealDeviceID(realDeviceID); err != nil {
				log.Printf("添加realDeviceID失败: %v", err)
				return nil, game_error.New(-2, "数据库更新错误")
			}

			if err := playerInfo.AddDeviceInfo(deviceInfo); err != nil {
				log.Printf("添加deviceInfo失败: %v", err)
				return nil, game_error.New(-2, "数据库更新错误")
			}

			// 保存新记录到数据库
			createResult := db.DB.Create(&playerInfo)
			if createResult.Error != nil {
				log.Printf("创建玩家设备信息记录失败: %v", createResult.Error)
				return nil, game_error.New(-2, "数据库更新错误")
			}

			log.Printf("成功创建 deviceID 为 '%s' 的玩家设备信息", deviceID)
		} else {
			log.Printf("查询玩家设备信息失败: %v", result.Error)
			return nil, game_error.New(-2, "数据库查询错误")
		}
	} else {
		// 检查是否需要更新设备信息
		// 获取当前realDeviceID和deviceInfo的历史记录数量
		realDeviceIDCount, err := playerInfo.GetRealDeviceIDCount()
		if err != nil {
			log.Printf("获取realDeviceID历史记录数量失败: %v", err)
			return nil, game_error.New(-2, "数据库查询错误")
		}

		deviceInfoCount, err := playerInfo.GetDeviceInfoCount()
		if err != nil {
			log.Printf("获取deviceInfo历史记录数量失败: %v", err)
			return nil, game_error.New(-2, "数据库查询错误")
		}

		// 如果任一历史记录数量超过10条，返回错误
		if realDeviceIDCount >= 10 || deviceInfoCount >= 10 {
			log.Printf("设备信息历史记录过多，deviceID: %s, realDeviceID记录数: %d, deviceInfo记录数: %d", 
				deviceID, realDeviceIDCount, deviceInfoCount)
			return nil, game_error.New(-14, "设备信息变更过于频繁，请联系客服")
		}

		// 添加新的设备信息到历史记录
		if err := playerInfo.AddRealDeviceID(realDeviceID); err != nil {
			log.Printf("添加realDeviceID失败: %v", err)
			return nil, game_error.New(-2, "数据库更新错误")
		}

		if err := playerInfo.AddDeviceInfo(deviceInfo); err != nil {
			log.Printf("添加deviceInfo失败: %v", err)
			return nil, game_error.New(-2, "数据库更新错误")
		}

		// 更新IP地址


		// 保存更新后的设备信息
		updateResult := db.DB.Save(&playerInfo)
		if updateResult.Error != nil {
			log.Printf("更新玩家设备信息失败: %v", updateResult.Error)
			return nil, game_error.New(-2, "数据库更新错误")
		}

		log.Printf("成功更新 deviceID 为 '%s' 的玩家设备信息", deviceID)
	}

	// 5. 成功响应构建
	// --------------------------------
	responseData := map[string]interface{}{
	//	"result": "success",
	//	"message": "设备信息更新成功",
	}

	return responseData, nil
}
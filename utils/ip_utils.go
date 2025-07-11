// internal/utils/ip_utils.go
package utils

import (
	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
//	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetClientIP 从HTTP请求中获取客户端IP地址
// 优先检查'CF-Connecting-IP'请求头，如果不存在则使用Go原生方法获取IP
func GetClientIP(c *gin.Context) string {
	// 优先从Cloudflare的请求头中获取真实IP
	ip := c.GetHeader("CF-Connecting-IP")
	if ip == "" {
		// 如果没有CF-Connecting-IP头，则使用Gin的标准方法获取客户端IP
		ip = c.ClientIP()
	}
	return ip
}

// RecordClientIP 记录客户端IP地址到PlayerInfo表中
// 此函数应在请求处理流程的早期调用，在转发给参数处理程序之前
// 同时验证dmm_playerinfo表中的roleID与dmm_playerdata表中的roleID是否匹配
func RecordClientIP(c *gin.Context, deviceID string) error {
	// 获取客户端IP
	ip := GetClientIP(c)
	if ip == "" {
		log.Println("无法获取客户端IP地址")
		return nil
	}
	
	// 首先查询dmm_playerdata获取roleID，作为基准数据
	var playerData model.PlayerData
	pdResult := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if pdResult.Error != nil {
		log.Printf("未找到 deviceID 为 '%s' 的玩家数据，无法记录IP", deviceID)
		return pdResult.Error
	}
	
	// 查找PlayerInfo记录
	var playerInfo model.PlayerInfo
	result := db.DB.Where("device_id = ?", deviceID).First(&playerInfo)
	if result.Error != nil {
		// 如果记录不存在，需要创建新记录
		
		// 检查roleID是否已存在于dmm_playerinfo表中
		var existingPlayerInfo model.PlayerInfo
		checkResult := db.DB.Where("role_id = ?", playerData.RoleID).First(&existingPlayerInfo)
		if checkResult.Error == nil {
			// 如果找到了记录，说明roleID已存在，返回错误
			log.Printf("roleID %d 已存在于dmm_playerinfo表中，deviceID: %s", playerData.RoleID, existingPlayerInfo.DeviceID)
			return game_error.New(-3001, "账号数据异常，请联系客服")
		} else if checkResult.Error != gorm.ErrRecordNotFound {
			// 如果是其他数据库错误
			log.Printf("检查roleID唯一性失败: %v", checkResult.Error)
			return checkResult.Error
		}
		
		// 创建新记录
		playerInfo = model.PlayerInfo{
			DeviceID: deviceID,
			RoleID:   playerData.RoleID, // 使用从playerData查询到的roleID
			IP:       ip,
		}
		result = db.DB.Create(&playerInfo)
		if result.Error != nil {
			log.Printf("创建PlayerInfo记录失败: %v", result.Error)
			return result.Error
		}
	} else {
		// 如果记录存在，需要验证roleID是否与dmm_playerdata表中的一致
		var playerData model.PlayerData
		pdResult := db.DB.Where("device_id = ?", deviceID).First(&playerData)
		if pdResult.Error != nil {
			log.Printf("未找到 deviceID 为 '%s' 的玩家数据，无法验证roleID", deviceID)
			return pdResult.Error
		}
		
		// 验证roleID是否匹配
		if playerInfo.RoleID != playerData.RoleID {
			log.Printf("roleID不匹配: playerInfo.RoleID=%d, playerData.RoleID=%d", playerInfo.RoleID, playerData.RoleID)
			return game_error.New(-3001, "账号数据异常，请联系客服")
		}
		
		// 如果roleID匹配，添加新的IP到历史记录中
		err := playerInfo.AddIP(ip)
		if err != nil {
			log.Printf("添加IP到历史记录失败: %v", err)
			return err
		}
		
		// 更新记录
		result = db.DB.Save(&playerInfo)
		if result.Error != nil {
			log.Printf("更新PlayerInfo记录失败: %v", result.Error)
			return result.Error
		}
	}
	
	return nil
}
// utils/card_skin_manager_example.go
package utils

import (
	"log"
	"dmmserver/game_error"
)

// 这个示例展示了如何使用CardSkinManager来更新玩家的卡牌皮肤数据

// 示例1：将两个独立的数组合并为一个对象数组并保存到数据库
func ExampleUpdateCardSkinData() {
	// 创建卡牌皮肤管理器
	cm := NewCardSkinManager()

	// 假设这些是从客户端接收到的两个独立数组
	cardOwnSkins := []int{600985, 600011, 600016}
	cardSkinExpiredTimes := []int{0, 0, 1547274976}

	// 更新玩家的卡牌皮肤数据
	deviceID := "example_device_id"
	err := cm.UpdateCardSkinData(deviceID, cardOwnSkins, cardSkinExpiredTimes)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新卡牌皮肤数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新卡牌皮肤数据失败: %v", err)
		}
		return
	}

	log.Printf("成功更新玩家 %s 的卡牌皮肤数据", deviceID)
}

// 示例2：从数据库读取卡牌皮肤数据并转换为独立的数组
func ExampleGetCardSkinData() {
	// 创建卡牌皮肤管理器
	cm := NewCardSkinManager()

	// 从数据库获取卡牌皮肤数据
	deviceID := "example_device_id"
	cardSkins, err := cm.GetCardSkins(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取卡牌皮肤数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取卡牌皮肤数据失败: %v", err)
		}
		return
	}

	// 从卡牌皮肤数据中提取独立的数组
	cardOwnSkins, cardSkinExpiredTimes, _ := cm.ExtractCardSkinArrays(cardSkins)

	// 现在可以使用这些数组
	log.Printf("玩家 %s 拥有 %d 个卡牌皮肤", deviceID, len(cardOwnSkins))
	log.Printf("第一个卡牌皮肤ID: %d, 过期时间: %d", 
		cardOwnSkins[0], cardSkinExpiredTimes[0])
	// 注意：CurStyle字段已完全由Card结构体管理，不再属于CardSkin结构体
}
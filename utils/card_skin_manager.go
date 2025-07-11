// utils/card_skin_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// CardSkin 表示一个卡牌皮肤的结构
type CardSkin struct {
	CardOwnSkin         int `json:"cardOwnSkin"`
	CardSkinExpiredTime int `json:"cardSkinExpiredTime"`
}

// CardSkinManager 提供卡牌皮肤数据的管理功能
type CardSkinManager struct {}

// NewCardSkinManager 创建一个新的卡牌皮肤管理器
func NewCardSkinManager() *CardSkinManager {
	return &CardSkinManager{}
}

// GetCardSkins 从数据库获取指定设备ID的所有卡牌皮肤数据
func (cm *CardSkinManager) GetCardSkins(deviceID string) ([]CardSkin, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析卡牌皮肤数据
	var cardSkins []CardSkin
	if playerData.CardSkins == "" || playerData.CardSkins == "null" {
		// 如果没有卡牌皮肤数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的卡牌皮肤数据为空，使用默认数据", deviceID)
		defaultCardSkins := cm.GetDefaultCardSkins()
		
		// 保存默认数据到数据库
		err := cm.SaveCardSkins(deviceID, defaultCardSkins)
		if err != nil {
			log.Printf("保存默认卡牌皮肤数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultCardSkins, nil
	}

	err := json.Unmarshal([]byte(playerData.CardSkins), &cardSkins)
	if err != nil {
		log.Printf("解析卡牌皮肤数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认卡牌皮肤数据")
		return cm.GetDefaultCardSkins(), nil
		// 注意：这里不返回错误，而是使用默认数据
	}

	return cardSkins, nil
}

// SaveCardSkins 保存卡牌皮肤数据到数据库
func (cm *CardSkinManager) SaveCardSkins(deviceID string, cardSkins []CardSkin) error {
	// 将卡牌皮肤数据序列化为JSON
	cardSkinsJSON, err := json.Marshal(cardSkins)
	if err != nil {
		log.Printf("序列化卡牌皮肤数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("card_skins", string(cardSkinsJSON))
	if result.Error != nil {
		log.Printf("更新卡牌皮肤数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// ConvertArraysToCardSkins 将两个独立的数组转换为卡牌皮肤对象数组
func (cm *CardSkinManager) ConvertArraysToCardSkins(cardOwnSkins []int, cardSkinExpiredTimes []int) []CardSkin {
	// 确定数组长度，取两个数组中最小的长度
	length := len(cardOwnSkins)
	if len(cardSkinExpiredTimes) < length {
		length = len(cardSkinExpiredTimes)
	}

	// 创建卡牌皮肤对象数组
	cardSkins := make([]CardSkin, length)
	for i := 0; i < length; i++ {
		cardSkins[i] = CardSkin{
			CardOwnSkin:         cardOwnSkins[i],
			CardSkinExpiredTime: cardSkinExpiredTimes[i],
		}
	}

	return cardSkins
}

// UpdateCardSkinData 更新玩家的卡牌皮肤数据
func (cm *CardSkinManager) UpdateCardSkinData(deviceID string, cardOwnSkins []int, cardSkinExpiredTimes []int) error {
	// 将两个数组转换为卡牌皮肤对象数组
	cardSkins := cm.ConvertArraysToCardSkins(cardOwnSkins, cardSkinExpiredTimes)

	// 保存到数据库
	return cm.SaveCardSkins(deviceID, cardSkins)
}

// ExtractCardSkinArrays 从卡牌皮肤对象数组中提取两个独立的数组
func (cm *CardSkinManager) ExtractCardSkinArrays(cardSkins []CardSkin) ([]int, []int, error) {
	cardOwnSkins := make([]int, len(cardSkins))
	cardSkinExpiredTimes := make([]int, len(cardSkins))

	for i, skin := range cardSkins {
		cardOwnSkins[i] = skin.CardOwnSkin
		cardSkinExpiredTimes[i] = skin.CardSkinExpiredTime
	}

	return cardOwnSkins, cardSkinExpiredTimes, nil
}

// ParseCardSkinsFromJSON 从JSON字符串解析卡牌皮肤数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (cm *CardSkinManager) ParseCardSkinsFromJSON(cardSkinsJSON string) ([]int, []int, error) {
	// 解析卡牌皮肤数据
	var cardSkins []CardSkin
	if cardSkinsJSON == "" || cardSkinsJSON == "null" {
		// 如果没有卡牌皮肤数据，使用默认数据
		log.Printf("卡牌皮肤数据为空，使用默认数据")
		defaultCardSkins := cm.GetDefaultCardSkins()
		return cm.ExtractCardSkinArrays(defaultCardSkins)
	}

	err := json.Unmarshal([]byte(cardSkinsJSON), &cardSkins)
	if err != nil {
		log.Printf("解析卡牌皮肤数据失败: %v", err)
		// 解析失败时，使用默认数据
		log.Printf("使用默认卡牌皮肤数据")
		defaultCardSkins := cm.GetDefaultCardSkins()
		return cm.ExtractCardSkinArrays(defaultCardSkins)
	}

	// 从卡牌皮肤数据中提取两个独立的数组
	return cm.ExtractCardSkinArrays(cardSkins)
}

// GetDefaultCardSkins 创建默认卡牌皮肤数据并返回
func (cm *CardSkinManager) GetDefaultCardSkins() []CardSkin {
	// 创建默认卡牌皮肤数据 - 包含cardOwnSkin和cardSkinExpiredTime属性
	defaultCardSkins := []CardSkin{
		{CardOwnSkin: 600985, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600011, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600016, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600021, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600036, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600041, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600046, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600026, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600031, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600051, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600056, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600071, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600076, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600081, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600086, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600091, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600121, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600206, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600211, CardSkinExpiredTime: 0},
		{CardOwnSkin: 600126, CardSkinExpiredTime: 0},
		// 添加更多默认卡牌皮肤...
	}
	
	return defaultCardSkins
}
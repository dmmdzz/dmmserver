// utils/card_style_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
)

// CardStyle 表示一个卡牌样式的结构
type CardStyle struct {
	CardOwnStyle         int `json:"cardOwnStyle"`
	CardStyleExpiredTime int `json:"cardStyleExpiredTime"`
}

// CardStyleManager 提供卡牌样式数据的管理功能
type CardStyleManager struct {}

// NewCardStyleManager 创建一个新的卡牌样式管理器
func NewCardStyleManager() *CardStyleManager {
	return &CardStyleManager{}
}

// GetCardStyles 从数据库获取指定设备ID的所有卡牌样式数据
func (cm *CardStyleManager) GetCardStyles(deviceID string) ([]CardStyle, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析卡牌样式数据
	var cardStyles []CardStyle
	if playerData.CardStyles == "" || playerData.CardStyles == "null" {
		// 如果没有卡牌样式数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的卡牌样式数据为空，使用默认数据", deviceID)
		defaultCardStyles := cm.GetDefaultCardStyles()
		
		// 保存默认数据到数据库
		err := cm.SaveCardStyles(deviceID, defaultCardStyles)
		if err != nil {
			log.Printf("保存默认卡牌样式数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultCardStyles, nil
	}

	err := json.Unmarshal([]byte(playerData.CardStyles), &cardStyles)
	if err != nil {
		log.Printf("解析卡牌样式数据失败: %v", err)
		// 解析失败时，直接返回错误
		return nil, game_error.New(-2, "数据处理错误")
	}

	return cardStyles, nil
}

// SaveCardStyles 保存卡牌样式数据到数据库
func (cm *CardStyleManager) SaveCardStyles(deviceID string, cardStyles []CardStyle) error {
	// 检查卡牌样式数据是否为空
	if len(cardStyles) == 0 {
		// 如果为空，使用空数组而不是空字符串
		log.Printf("卡牌样式数据为空，使用空数组 '[]' 代替")
		// result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("card_styles", "[]")
		// if result.Error != nil {
		// 	log.Printf("更新卡牌样式数据失败: %v", result.Error)
			return game_error.New(-2, "数据库更新错误")
		// }
		// return nil
	}

	// 将卡牌样式数据序列化为JSON
	cardStylesJSON, err := json.Marshal(cardStyles)
	if err != nil {
		log.Printf("序列化卡牌样式数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("card_styles", string(cardStylesJSON))
	if result.Error != nil {
		log.Printf("更新卡牌样式数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// ConvertArraysToCardStyles 将两个独立的数组转换为卡牌样式对象数组
func (cm *CardStyleManager) ConvertArraysToCardStyles(cardOwnStyles []int, cardStyleExpiredTimes []int) []CardStyle {
	// 确定数组长度，取两个数组中最小的长度
	length := len(cardOwnStyles)
	if len(cardStyleExpiredTimes) < length {
		length = len(cardStyleExpiredTimes)
	}

	// 创建卡牌样式对象数组
	cardStyles := make([]CardStyle, length)
	for i := 0; i < length; i++ {
		cardStyles[i] = CardStyle{
			CardOwnStyle:         cardOwnStyles[i],
			CardStyleExpiredTime: cardStyleExpiredTimes[i],
		}
	}

	return cardStyles
}

// UpdateCardStyleData 更新玩家的卡牌样式数据
func (cm *CardStyleManager) UpdateCardStyleData(deviceID string, cardOwnStyles []int, cardStyleExpiredTimes []int) error {
	// 将两个数组转换为卡牌样式对象数组
	cardStyles := cm.ConvertArraysToCardStyles(cardOwnStyles, cardStyleExpiredTimes)

	// 保存到数据库
	err := cm.SaveCardStyles(deviceID, cardStyles)
	if err != nil {
		log.Printf("更新卡牌样式数据失败: %v", err)
		return err
	}
	
	return nil
}

// ExtractCardStyleArrays 从卡牌样式对象数组中提取两个独立的数组
func (cm *CardStyleManager) ExtractCardStyleArrays(cardStyles []CardStyle) ([]int, []int, error) {
	cardOwnStyles := make([]int, len(cardStyles))
	cardStyleExpiredTimes := make([]int, len(cardStyles))

	for i, style := range cardStyles {
		cardOwnStyles[i] = style.CardOwnStyle
		cardStyleExpiredTimes[i] = style.CardStyleExpiredTime
	}

	return cardOwnStyles, cardStyleExpiredTimes, nil
}

// GetDefaultCardStyles 创建默认卡牌样式数据并返回
func (cm *CardStyleManager) GetDefaultCardStyles() []CardStyle {
	// 创建默认卡牌样式数据 - 包含cardOwnStyle和cardStyleExpiredTime属性
	defaultCardStyles := []CardStyle{
		{CardOwnStyle: 650041, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650051, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650031, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650061, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650011, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650021, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650101, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650081, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650071, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650091, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650111, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650121, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650181, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650171, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650151, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650161, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650191, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650201, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650241, CardStyleExpiredTime: 0},
		{CardOwnStyle: 650211, CardStyleExpiredTime: 0},
		// 添加更多默认卡牌样式...
	}
	
	return defaultCardStyles
}

// ParseCardStylesFromJSON 从JSON字符串解析卡牌样式数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (cm *CardStyleManager) ParseCardStylesFromJSON(cardStylesJSON string) ([]int, []int, error) {
	// 解析卡牌样式数据
	var cardStyles []CardStyle
	if cardStylesJSON == "" || cardStylesJSON == "null" {
		// 如果没有卡牌样式数据，使用默认数据
		log.Printf("卡牌样式数据为空，使用默认数据")
		defaultCardStyles := cm.GetDefaultCardStyles()
		return cm.ExtractCardStyleArrays(defaultCardStyles)
	}

	err := json.Unmarshal([]byte(cardStylesJSON), &cardStyles)
	if err != nil {
		log.Printf("解析卡牌样式数据失败: %v", err)
		// 解析失败时，返回错误
		return nil, nil, game_error.New(-2, "数据处理错误")
	}

	// 从卡牌样式数据中提取两个独立的数组
	return cm.ExtractCardStyleArrays(cardStyles)
}


// ConvertToClientFormat 将数据库格式的卡牌样式数据转换为客户端需要的格式
// 例如：将[{"cardOwnStyle": 650041, "cardStyleExpiredTime": 0}, {"cardOwnStyle": 650051, "cardStyleExpiredTime": 0}]
// 转换为{"cardOwnStyle": [650041, 650051], "cardStyleExpiredTime": [0, 0]}
func (cm *CardStyleManager) ConvertToClientFormat(cardStylesJSON string) (map[string]interface{}, error) {
	// 解析卡牌样式数据
	cardOwnStyles, cardStyleExpiredTimes, err := cm.ParseCardStylesFromJSON(cardStylesJSON)
	if err != nil {
		log.Printf("解析卡牌样式数据失败: %v", err)
		return nil, err
	}

	// 构建客户端需要的格式
	result := map[string]interface{}{
		"cardOwnStyle":         cardOwnStyles,
		"cardStyleExpiredTime": cardStyleExpiredTimes,
	}

	return result, nil
}
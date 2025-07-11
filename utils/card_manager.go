// utils/card_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// Card 表示一个卡牌的结构
type Card struct {
	ID       int         `json:"id"`       // 卡牌ID
	Level    int         `json:"level"`    // 卡牌等级
	CurSkin  interface{} `json:"curSkin"`  // 当前使用的皮肤
	CurStyle interface{} `json:"curStyle"` // 当前使用的风格
}

// CardManager 提供卡牌数据的管理功能
type CardManager struct {}

// NewCardManager 创建一个新的卡牌管理器
func NewCardManager() *CardManager {
	return &CardManager{}
}

// GetCards 从数据库获取指定设备ID的所有卡牌数据
func (cm *CardManager) GetCards(deviceID string) ([]Card, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析卡牌数据
	var cards []Card
	if playerData.Cards == "" || playerData.Cards == "null" {
		// 如果没有卡牌数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的卡牌数据为空，使用默认数据", deviceID)
		defaultCards := cm.GetDefaultCards()
		
		// 保存默认数据到数据库
		err := cm.SaveCards(deviceID, defaultCards)
		if err != nil {
			log.Printf("保存默认卡牌数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultCards, nil
	}

	err := json.Unmarshal([]byte(playerData.Cards), &cards)
	if err != nil {
		log.Printf("解析卡牌数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认卡牌数据")
		return cm.GetDefaultCards(), nil
		// 注意：这里不返回错误，而是使用默认数据
	}

	return cards, nil
}

// SaveCards 保存卡牌数据到数据库
func (cm *CardManager) SaveCards(deviceID string, cards []Card) error {
	// 检查卡牌数据是否为空
	if len(cards) == 0 {
		// 如果为空，使用空数组而不是空字符串
		log.Printf("卡牌数据为空")
	//	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("cards", "[]")
	//	if result.Error != nil {
		//	log.Printf("更新卡牌数据失败: %v", result.Error)
			return game_error.New(-2, "数据库更新错误")
		//}
		//return nil
	}

	// 将卡牌数据序列化为JSON
	cardsJSON, err := json.Marshal(cards)
	if err != nil {
		log.Printf("序列化卡牌数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("cards", string(cardsJSON))
	if result.Error != nil {
		log.Printf("更新卡牌数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// ConvertArraysToCards 将四个独立的数组转换为卡牌对象数组
func (cm *CardManager) ConvertArraysToCards(cardIDs []int, cardLevels []int, cardCurSkins []interface{}, cardCurStyles []interface{}) []Card {
	// 确定数组长度，取四个数组中最小的长度
	length := len(cardIDs)
	if len(cardLevels) < length {
		length = len(cardLevels)
	}
	if len(cardCurSkins) < length {
		length = len(cardCurSkins)
	}
	if len(cardCurStyles) < length {
		length = len(cardCurStyles)
	}

	// 创建卡牌对象数组
	cards := make([]Card, length)
	for i := 0; i < length; i++ {
		cards[i] = Card{
			ID:       cardIDs[i],
			Level:    cardLevels[i],
			CurSkin:  cardCurSkins[i],
			CurStyle: cardCurStyles[i],
		}
	}

	return cards
}

// UpdateCardData 更新玩家的卡牌数据
func (cm *CardManager) UpdateCardData(deviceID string, cardIDs []int, cardLevels []int, cardCurSkins []interface{}, cardCurStyles []interface{}) error {
	// 将四个数组转换为卡牌对象数组
	cards := cm.ConvertArraysToCards(cardIDs, cardLevels, cardCurSkins, cardCurStyles)

	// 保存到数据库
	return cm.SaveCards(deviceID, cards)
}

// ExtractCardArrays 从卡牌对象数组中提取四个独立的数组
func (cm *CardManager) ExtractCardArrays(cards []Card) ([]int, []int, []interface{}, []interface{}, error) {
	cardIDs := make([]int, len(cards))
	cardLevels := make([]int, len(cards))
	cardCurSkins := make([]interface{}, len(cards))
	cardCurStyles := make([]interface{}, len(cards))

	for i, card := range cards {
		cardIDs[i] = card.ID
		cardLevels[i] = card.Level
		cardCurSkins[i] = card.CurSkin
		cardCurStyles[i] = card.CurStyle
	}

	return cardIDs, cardLevels, cardCurSkins, cardCurStyles, nil
}

// ParseCardsFromJSON 从JSON字符串解析卡牌数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (cm *CardManager) ParseCardsFromJSON(cardsJSON string) ([]int, []int, []interface{}, []interface{}, error) {
	// 解析卡牌数据
	var cards []Card
	if cardsJSON == "" || cardsJSON == "null" {
		// 如果没有卡牌数据，使用默认数据
		log.Printf("卡牌数据为空，使用默认数据")
		defaultCards := cm.GetDefaultCards()
		return cm.ExtractCardArrays(defaultCards)
	}

	err := json.Unmarshal([]byte(cardsJSON), &cards)
	if err != nil {
		log.Printf("解析卡牌数据失败: %v", err)
		// 解析失败时，使用默认数据
		log.Printf("使用默认卡牌数据")
		defaultCards := cm.GetDefaultCards()
		return cm.ExtractCardArrays(defaultCards)
	}

	// 从卡牌数据中提取四个独立的数组
	return cm.ExtractCardArrays(cards)
}

// GetDefaultCards 创建默认卡牌数据并返回
func (cm *CardManager) GetDefaultCards() []Card {
	// 创建默认卡牌数据 - 包含id、level、curSkin和curStyle属性
	defaultCards := []Card{
		{ID: 100, Level: 1, CurSkin: 50001, CurStyle: 650061},
		{ID: 101, Level: 1, CurSkin: 50002, CurStyle: 0},
		{ID: 102, Level: 1, CurSkin: 50003, CurStyle: 0},
		{ID: 103, Level: 1, CurSkin: 50004, CurStyle: 0},
		{ID: 104, Level: 1, CurSkin: 50005, CurStyle: 0},
		{ID: 105, Level: 1, CurSkin: 50006, CurStyle: 0},
		{ID: 106, Level: 1, CurSkin: 50007, CurStyle: 0},
		{ID: 107, Level: 1, CurSkin: 50008, CurStyle: 0},
		{ID: 108, Level: 1, CurSkin: 50009, CurStyle: 0},
		{ID: 109, Level: 1, CurSkin: 50010, CurStyle: 0},
		// 添加更多默认卡牌...
	}
	
	return defaultCards
}

// UpdateCardField 更新指定卡牌的特定字段
func (cm *CardManager) UpdateCardField(deviceID string, cardID int, fieldName string, fieldValue interface{}) error {
	// 获取当前卡牌数据
	cards, err := cm.GetCards(deviceID)
	if err != nil {
		return err
	}

	// 查找并更新指定卡牌的字段
	cardFound := false
	for i, card := range cards {
		if card.ID == cardID {
			switch fieldName {
			case "level":
				if level, ok := fieldValue.(int); ok {
					cards[i].Level = level
				} else {
					return game_error.New(-2, "字段类型错误")
				}
			case "curSkin":
				cards[i].CurSkin = fieldValue
			case "curStyle":
				cards[i].CurStyle = fieldValue
			default:
				return game_error.New(-2, "不支持的字段名称")
			}
			cardFound = true
			break
		}
	}

	// 如果没有找到指定卡牌，则添加新卡牌
	if !cardFound {
		newCard := Card{ID: cardID}
		switch fieldName {
		case "level":
			if level, ok := fieldValue.(int); ok {
				newCard.Level = level
			} else {
				return game_error.New(-2, "字段类型错误")
			}
		case "curSkin":
			newCard.CurSkin = fieldValue
		case "curStyle":
			newCard.CurStyle = fieldValue
		default:
			return game_error.New(-2, "不支持的字段名称")
		}
		cards = append(cards, newCard)
	}

	// 保存更新后的卡牌数据
	return cm.SaveCards(deviceID, cards)
}

// AddCard 添加新卡牌或更新现有卡牌
func (cm *CardManager) AddCard(deviceID string, card Card) error {
	// 获取当前卡牌数据
	cards, err := cm.GetCards(deviceID)
	if err != nil {
		return err
	}

	// 查找是否已存在相同ID的卡牌
	cardFound := false
	for i, existingCard := range cards {
		if existingCard.ID == card.ID {
			// 更新现有卡牌
			cards[i] = card
			cardFound = true
			break
		}
	}

	// 如果没有找到相同ID的卡牌，则添加新卡牌
	if !cardFound {
		cards = append(cards, card)
	}

	// 保存更新后的卡牌数据
	return cm.SaveCards(deviceID, cards)
}

// RemoveCard 移除指定ID的卡牌
func (cm *CardManager) RemoveCard(deviceID string, cardID int) error {
	// 获取当前卡牌数据
	cards, err := cm.GetCards(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除指定ID的卡牌
	found := false
	for i, card := range cards {
		if card.ID == cardID {
			// 移除卡牌（通过切片操作）
			cards = append(cards[:i], cards[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		// 如果没有找到指定ID的卡牌，返回错误
		return game_error.New(-2, "未找到指定ID的卡牌")
	}

	// 保存更新后的卡牌数据
	return cm.SaveCards(deviceID, cards)
}
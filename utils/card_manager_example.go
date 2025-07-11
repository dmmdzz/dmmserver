// utils/card_manager_example.go
package utils

import (
	"log"
	"dmmserver/game_error"
)

// 这个示例展示了如何使用CardManager来更新玩家的卡牌数据

// 示例1：将三个独立的数组合并为一个对象数组并保存到数据库
func ExampleUpdateCardData() {
	// 创建卡牌管理器
	cm := NewCardManager()

	// 假设这些是从客户端接收到的四个独立数组
	cardIDs := []int{100, 101, 102}
	cardLevels := []int{5, 3, 1}
	cardCurSkins := []interface{}{600985, 0, 0}
	cardCurStyles := []interface{}{650061, 0, 0}

	// 更新玩家的卡牌数据
	deviceID := "example_device_id"
	err := cm.UpdateCardData(deviceID, cardIDs, cardLevels, cardCurSkins, cardCurStyles)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新卡牌数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新卡牌数据失败: %v", err)
		}
		return
	}

	log.Printf("成功更新玩家 %s 的卡牌数据", deviceID)
}

// 示例2：从数据库读取卡牌数据并转换为三个独立的数组
func ExampleGetCardData() {
	// 创建卡牌管理器
	cm := NewCardManager()

	// 从数据库获取卡牌数据
	deviceID := "example_device_id"
	cards, err := cm.GetCards(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取卡牌数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取卡牌数据失败: %v", err)
		}
		return
	}

	// 从卡牌数据中提取四个独立的数组
	cardIDs, cardLevels, cardCurSkins, cardCurStyles, err := cm.ExtractCardArrays(cards)
	if err != nil {
	    log.Printf("提取卡牌数据失败: %v", err)
	    return
	}

	// 现在可以使用这四个数组
	log.Printf("玩家 %s 拥有 %d 个卡牌", deviceID, len(cardIDs))
	log.Printf("第一个卡牌ID: %d, 等级: %d, 当前皮肤: %v, 当前风格: %v", 
		cardIDs[0], cardLevels[0], cardCurSkins[0], cardCurStyles[0])
}

// 示例3：更新特定卡牌的特定字段
func ExampleUpdateCardField() {
	// 创建卡牌管理器
	cm := NewCardManager()

	// 更新卡牌等级
	deviceID := "example_device_id"
	cardID := 100
	newLevel := 10
	err := cm.UpdateCardField(deviceID, cardID, "level", newLevel)
	if err != nil {
		log.Printf("更新卡牌等级失败: %v", err)
		return
	}

	log.Printf("成功将玩家 %s 的卡牌 %d 等级更新为 %d", deviceID, cardID, newLevel)

	// 更新卡牌皮肤
	newSkin := 600986
	err = cm.UpdateCardField(deviceID, cardID, "curSkin", newSkin)
	if err != nil {
		log.Printf("更新卡牌皮肤失败: %v", err)
		return
	}

	log.Printf("成功将玩家 %s 的卡牌 %d 皮肤更新为 %d", deviceID, cardID, newSkin)
}

// 示例4：添加新卡牌或更新现有卡牌
func ExampleAddCard() {
	// 创建卡牌管理器
	cm := NewCardManager()

	// 创建一个新卡牌
	newCard := Card{
		ID:      110,
		Level:   5,
		CurSkin: 600990,
		CurStyle: 650062,
	}

	// 添加或更新卡牌
	deviceID := "example_device_id"
	err := cm.AddCard(deviceID, newCard)
	if err != nil {
		log.Printf("添加卡牌失败: %v", err)
		return
	}

	log.Printf("成功为玩家 %s 添加或更新卡牌 %d", deviceID, newCard.ID)
}

// 示例5：移除指定ID的卡牌
func ExampleRemoveCard() {
	// 创建卡牌管理器
	cm := NewCardManager()

	// 移除卡牌
	deviceID := "example_device_id"
	cardID := 110
	err := cm.RemoveCard(deviceID, cardID)
	if err != nil {
		log.Printf("移除卡牌失败: %v", err)
		return
	}

	log.Printf("成功从玩家 %s 的卡牌列表中移除卡牌 %d", deviceID, cardID)
}
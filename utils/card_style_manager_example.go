// utils/card_style_manager_example.go
package utils

import (
	"log"
)

// 这个示例展示了如何使用CardStyleManager来更新玩家的卡牌样式数据

// 示例1：将两个独立的数组合并为一个对象数组并保存到数据库
func ExampleUpdateCardStyleData() {
	// 创建卡牌样式管理器
	cm := NewCardStyleManager()

	// 假设这些是从客户端接收到的两个独立数组
	cardOwnStyles := []int{650041, 650051, 650031}
	cardStyleExpiredTimes := []int{0, 0, 1547274976}

	// 更新玩家的卡牌样式数据
	deviceID := "example_device_id"
	err := cm.UpdateCardStyleData(deviceID, cardOwnStyles, cardStyleExpiredTimes)
	if err != nil {
		log.Printf("更新卡牌样式数据失败: %v", err)
		return
	}

	log.Printf("成功更新玩家 %s 的卡牌样式数据", deviceID)
}

// 示例2：从数据库读取卡牌样式数据并转换为两个独立的数组
func ExampleGetCardStyleData() {
	// 创建卡牌样式管理器
	cm := NewCardStyleManager()

	// 从数据库获取卡牌样式数据
	deviceID := "example_device_id"
	cardStyles, err := cm.GetCardStyles(deviceID)
	if err != nil {
		log.Printf("获取卡牌样式数据失败: %v", err)
		return
	}

	// 从卡牌样式数据中提取两个独立的数组
	cardOwnStyles, cardStyleExpiredTimes, err := cm.ExtractCardStyleArrays(cardStyles)
	if err != nil {
		log.Printf("提取卡牌样式数据失败: %v", err)
		return
	}

	// 现在可以使用这两个数组
	log.Printf("玩家 %s 拥有 %d 个卡牌样式", deviceID, len(cardOwnStyles))
	log.Printf("第一个卡牌样式ID: %d, 过期时间: %d", 
		cardOwnStyles[0], cardStyleExpiredTimes[0])
}
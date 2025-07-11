// utils/skin_part_manager_example.go
package utils

import (
	"log"
	"dmmserver/game_error"
)

// 这个示例展示了如何使用SkinPartManager来更新玩家的皮肤部件数据

// 示例1：将四个独立的数组合并为一个对象数组并保存到数据库
func ExampleUpdateSkinPartData() {
	// 创建皮肤部件管理器
	sm := NewSkinPartManager()

	// 假设这些是从客户端接收到的四个独立数组
	skinPartIDs := []string{"1001", "1002", "1003"}
	skinPartColors := []string{"67", "103", "84"}
	expiredTimes := []int{0, 0, 1547274976}
	skinDecals := []interface{}{0, 0, 0}

	// 更新玩家的皮肤部件数据
	deviceID := "example_device_id"
	err := sm.UpdateSkinPartData(deviceID, skinPartIDs, skinPartColors, expiredTimes, skinDecals)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新皮肤部件数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新皮肤部件数据失败: %v", err)
		}
		return
	}

	log.Printf("成功更新玩家 %s 的皮肤部件数据", deviceID)
}

// 示例2：从数据库读取皮肤部件数据并转换为四个独立的数组
func ExampleGetSkinPartData() {
	// 创建皮肤部件管理器
	sm := NewSkinPartManager()

	// 从数据库获取皮肤部件数据
	deviceID := "example_device_id"
	skinParts, err := sm.GetSkinParts(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取皮肤部件数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取皮肤部件数据失败: %v", err)
		}
		return
	}

	// 从皮肤部件数据中提取四个独立的数组
	skinPartIDs, skinPartColors, expiredTimes, skinDecals,err := sm.ExtractSkinPartArrays(skinParts)

	// 现在可以使用这四个数组
	log.Printf("玩家 %s 拥有 %d 个皮肤部件", deviceID, len(skinPartIDs))
	log.Printf("第一个皮肤部件ID: %s, 颜色: %s, 过期时间: %d, 贴花: %v", 
		skinPartIDs[0], skinPartColors[0], expiredTimes[0], skinDecals[0])
}

// 示例3：更新单个皮肤部件的特定字段
func ExampleUpdateSkinPartField() {
	// 创建皮肤部件管理器
	sm := NewSkinPartManager()

	// 更新指定皮肤部件的颜色
	deviceID := "example_device_id"
	skinPartID := "1001"
	newColor := "75" // 新的颜色值

	err := sm.UpdateSkinPartField(deviceID, skinPartID, "skinPartColors", newColor)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新皮肤部件颜色失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新皮肤部件颜色失败: %v", err)
		}
		return
	}

	log.Printf("成功更新皮肤部件 %s 的颜色为 %s", skinPartID, newColor)
}

// 示例4：添加新的皮肤部件
func ExampleAddSkinPart() {
	// 创建皮肤部件管理器
	sm := NewSkinPartManager()

	// 创建新的皮肤部件
	newSkinPart := SkinPart{
		SkinPartIDs:    "3001",
		SkinPartColors: "42",
		ExpiredTime:    0,
		SkinDecals:     0,
	}

	// 添加新皮肤部件
	deviceID := "example_device_id"
	err := sm.AddSkinPart(deviceID, newSkinPart)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("添加皮肤部件失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("添加皮肤部件失败: %v", err)
		}
		return
	}

	log.Printf("成功添加新皮肤部件 %s", newSkinPart.SkinPartIDs)
}

// 示例5：移除皮肤部件
func ExampleRemoveSkinPart() {
	// 创建皮肤部件管理器
	sm := NewSkinPartManager()

	// 移除指定ID的皮肤部件
	deviceID := "example_device_id"
	skinPartID := "2001"

	err := sm.RemoveSkinPart(deviceID, skinPartID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("移除皮肤部件失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("移除皮肤部件失败: %v", err)
		}
		return
	}

	log.Printf("成功移除皮肤部件 %s", skinPartID)
}
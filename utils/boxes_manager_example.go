// utils/boxes_manager_example.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/game_error"
)

// 示例1：获取玩家的装饰框数据
func ExampleGetBoxesData(deviceID string) (map[string]interface{}, error) {
	// 创建装饰框管理器
	bm := NewBoxesManager()

	// 获取装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取装饰框数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取装饰框数据失败: %v", err)
		}
		return nil, err
	}

	// 将装饰框数据转换为map格式，以便在handler中使用
	result := map[string]interface{}{}
	boxesDataJSON, _ := json.Marshal(boxesData)
	json.Unmarshal(boxesDataJSON, &result)

	return result, nil
}

// 示例2：更新玩家拥有的头像框
func ExampleUpdateHeadBoxes(deviceID string, headBoxIDs []int, expiredTimes []int) error {
	// 创建装饰框管理器
	bm := NewBoxesManager()

	// 更新玩家拥有的头像框数据
	err := bm.UpdateHeadBoxes(deviceID, headBoxIDs, expiredTimes)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新头像框数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新头像框数据失败: %v", err)
		}
		return err
	}

	log.Printf("成功更新玩家 %s 的头像框数据", deviceID)
	return nil
}

// 示例3：添加新头像框
func ExampleAddNewHeadBox(deviceID string, headBoxID int, expiredTime int) error {
	// 创建装饰框管理器
	bm := NewBoxesManager()

	// 添加新头像框
	err := bm.AddHeadBox(deviceID, headBoxID, expiredTime)
	if err != nil {
		log.Printf("添加头像框失败: %v", err)
		return err
	}

	log.Printf("成功为玩家 %s 添加头像框 %d", deviceID, headBoxID)
	return nil
}

// 示例4：更新玩家拥有的聊天气泡
func ExampleUpdateBubbleBoxes(deviceID string, bubbleBoxIDs []int, expiredTimes []int) error {
	// 创建装饰框管理器
	bm := NewBoxesManager()

	// 更新玩家拥有的聊天气泡数据
	err := bm.UpdateBubbleBoxes(deviceID, bubbleBoxIDs, expiredTimes)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新聊天气泡数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新聊天气泡数据失败: %v", err)
		}
		return err
	}

	log.Printf("成功更新玩家 %s 的聊天气泡数据", deviceID)
	return nil
}

// 示例5：添加新聊天气泡
func ExampleAddNewBubbleBox(deviceID string, bubbleBoxID int, expiredTime int) error {
	// 创建装饰框管理器
	bm := NewBoxesManager()

	// 添加新聊天气泡
	err := bm.AddBubbleBox(deviceID, bubbleBoxID, expiredTime)
	if err != nil {
		log.Printf("添加聊天气泡失败: %v", err)
		return err
	}

	log.Printf("成功为玩家 %s 添加聊天气泡 %d", deviceID, bubbleBoxID)
	return nil
}
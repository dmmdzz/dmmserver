// utils/emotion_manager_example.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/game_error"
)

// 示例1：获取玩家的表情数据
func ExampleGetEmotionData(deviceID string) (map[string]interface{}, error) {
	// 创建表情管理器
	em := NewEmotionManager()

	// 获取表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取表情数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取表情数据失败: %v", err)
		}
		return nil, err
	}

	// 将表情数据转换为map格式，以便在handler中使用
	result := map[string]interface{}{}
	emotionDataJSON, _ := json.Marshal(emotionData)
	json.Unmarshal(emotionDataJSON, &result)

	return result, nil
}

// 示例2：更新玩家拥有的表情
func ExampleUpdateOwnedEmotions(deviceID string, ids []interface{}, expiredTimes []int) error {
	// 创建表情管理器
	em := NewEmotionManager()

	// 更新玩家拥有的表情数据
	err := em.UpdateOwnedEmotions(deviceID, ids, expiredTimes)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("更新表情数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("更新表情数据失败: %v", err)
		}
		return err
	}

	log.Printf("成功更新玩家 %s 的表情数据", deviceID)
	return nil
}

// 示例3：添加新表情
func ExampleAddNewEmotion(deviceID string, emotionID interface{}, expiredTime int) error {
	// 创建表情管理器
	em := NewEmotionManager()

	// 添加新表情
	err := em.AddOwnedEmotion(deviceID, emotionID, expiredTime)
	if err != nil {
		log.Printf("添加表情失败: %v", err)
		return err
	}

	log.Printf("成功为玩家 %s 添加表情 %v", deviceID, emotionID)
	return nil
}

// 示例4：更新角色的表情配置
func ExampleUpdateCharacterEmotionConfig(deviceID string, characterID int, config []interface{}) error {
	// 创建表情管理器
	em := NewEmotionManager()

	// 更新角色的表情配置
	err := em.UpdateEmotionConfig(deviceID, characterID, config)
	if err != nil {
		log.Printf("更新角色表情配置失败: %v", err)
		return err
	}

	log.Printf("成功更新玩家 %s 的角色 %d 表情配置", deviceID, characterID)
	return nil
}

// 示例5：获取所有角色的表情配置
func ExampleGetAllEmotionConfigs(deviceID string) ([]map[string]interface{}, error) {
	// 创建表情管理器
	em := NewEmotionManager()

	// 获取所有角色的表情配置
	configs, err := em.GetAllEmotionConfigs(deviceID)
	if err != nil {
		log.Printf("获取表情配置失败: %v", err)
		return nil, err
	}

	// 将配置数据转换为map格式，以便在handler中使用
	result := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		configMap := map[string]interface{}{}
		configJSON, _ := json.Marshal(config)
		json.Unmarshal(configJSON, &configMap)
		result = append(result, configMap)
	}

	return result, nil
}
// utils/lightness_manager_example.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/game_error"
)

// 示例1：获取玩家的炫光数据
func ExampleGetLightnessData(deviceID string) (map[string]interface{}, error) {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 获取炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取炫光数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取炫光数据失败: %v", err)
		}
		return nil, err
	}

	// 将炫光数据转换为map格式，以便在handler中使用
	result := map[string]interface{}{}
	lightnessDataJSON, _ := json.Marshal(lightnessData)
	json.Unmarshal(lightnessDataJSON, &result)

	return result, nil
}

// 示例2：更新玩家的炫光ID列表
func ExampleUpdateLightnessIDs(deviceID string, ids []int, expiredTimes []int) error {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 检查参数长度是否一致
	if len(ids) != len(expiredTimes) {
		log.Printf("参数长度不一致: ids=%d, expiredTimes=%d", len(ids), len(expiredTimes))
		return game_error.New(-2, "参数错误")
	}

	// 更新炫光ID列表
	err := lm.UpdateLightnessIDs(deviceID, ids, expiredTimes)
	if err != nil {
		log.Printf("更新炫光ID列表失败: %v", err)
		return err
	}

	log.Printf("成功更新玩家 %s 的炫光ID列表", deviceID)
	return nil
}

// 示例3：更新玩家的炫光配置
func ExampleUpdateLightnessConfig(deviceID string, config int) error {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 更新炫光配置
	err := lm.UpdateLightnessConfig(deviceID, config)
	if err != nil {
		log.Printf("更新炫光配置失败: %v", err)
		return err
	}

	log.Printf("成功更新玩家 %s 的炫光配置为 %d", deviceID, config)
	return nil
}

// 示例4：添加新炫光
func ExampleAddLightness(deviceID string, id int, expiredTime int) error {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 添加新炫光
	err := lm.AddLightness(deviceID, id, expiredTime)
	if err != nil {
		log.Printf("添加炫光失败: %v", err)
		return err
	}

	log.Printf("成功为玩家 %s 添加炫光 %d，过期时间 %d", deviceID, id, expiredTime)
	return nil
}

// 示例5：移除指定炫光
func ExampleRemoveLightness(deviceID string, id int) error {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 移除炫光
	err := lm.RemoveLightness(deviceID, id)
	if err != nil {
		log.Printf("移除炫光失败: %v", err)
		return err
	}

	log.Printf("成功从玩家 %s 的炫光列表中移除炫光 %d", deviceID, id)
	return nil
}

// 示例6：获取玩家的炫光ID列表
func ExampleGetLightnessIDs(deviceID string) ([]int, []int, error) {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 获取炫光ID列表
	ids, expiredTimes, err := lm.GetLightnessIDs(deviceID)
	if err != nil {
		log.Printf("获取炫光ID列表失败: %v", err)
		return nil, nil, err
	}

	log.Printf("玩家 %s 拥有 %d 个炫光", deviceID, len(ids))
	return ids, expiredTimes, nil
}

// 示例7：获取玩家的炫光配置
func ExampleGetLightnessConfig(deviceID string) (int, error) {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 获取炫光配置
	config, err := lm.GetLightnessConfig(deviceID)
	if err != nil {
		log.Printf("获取炫光配置失败: %v", err)
		return 0, err
	}

	log.Printf("玩家 %s 的炫光配置为 %d", deviceID, config)
	return config, nil
}

// 示例8：从数据库JSON字符串解析炫光数据
// 此方法用于直接解析数据库中存储的炫光数据，避免重复查询数据库
func ParseLightnessDataFromDBString(jsonData string) (map[string]interface{}, error) {
	// 创建炫光管理器
	lm := NewLightnessManager()

	// 解析JSON数据
	result, err := lm.ParseLightnessDataFromJSON(jsonData)
	if err != nil {
		log.Printf("解析炫光数据失败: %v", err)
		return nil, err
	}

	return result, nil
}
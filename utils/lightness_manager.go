// utils/lightness_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// LightnessData 表示完整的炫光数据结构
type LightnessData struct {
	ID          []int `json:"id"`          // 炫光ID
	ExpiredTime []int `json:"expiredTime"` // 过期时间
	Config      int   `json:"config"`      // 炫光配置
}

// LightnessManager 提供炫光数据的管理功能
type LightnessManager struct {}

// NewLightnessManager 创建一个新的炫光管理器
func NewLightnessManager() *LightnessManager {
	return &LightnessManager{}
}

// GetLightnessData 从数据库获取指定设备ID的炫光数据
func (lm *LightnessManager) GetLightnessData(deviceID string) (*LightnessData, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析炫光数据
	var lightnessData LightnessData
	if playerData.LightnessData == "" || playerData.LightnessData == "null" {
		// 如果没有炫光数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的炫光数据为空，使用默认数据", deviceID)
		defaultLightnessData := lm.GetDefaultLightnessData()
		
		// 保存默认数据到数据库
		err := lm.SaveLightnessData(deviceID, defaultLightnessData)
		if err != nil {
			log.Printf("保存默认炫光数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultLightnessData, nil
	}

	err := json.Unmarshal([]byte(playerData.LightnessData), &lightnessData)
	if err != nil {
		log.Printf("解析炫光数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认炫光数据")
		return lm.GetDefaultLightnessData(), nil
	}

	return &lightnessData, nil
}

// SaveLightnessData 保存炫光数据到数据库
func (lm *LightnessManager) SaveLightnessData(deviceID string, lightnessData *LightnessData) error {
	// 将炫光数据序列化为JSON
	lightnessDataJSON, err := json.Marshal(lightnessData)
	if err != nil {
		log.Printf("序列化炫光数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("lightness_data", string(lightnessDataJSON))
	if result.Error != nil {
		log.Printf("更新炫光数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// GetDefaultLightnessData 创建默认炫光数据并返回
func (lm *LightnessManager) GetDefaultLightnessData() *LightnessData {
	// 创建默认炫光数据
	defaultLightnessData := &LightnessData{
		ID:          []int{},
		ExpiredTime: []int{},
		Config:      0,
	}
	
	return defaultLightnessData
}

// UpdateLightnessIDs 更新玩家拥有的炫光ID列表
func (lm *LightnessManager) UpdateLightnessIDs(deviceID string, ids []int, expiredTimes []int) error {
	// 获取当前炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		return err
	}

	// 更新炫光ID和过期时间
	lightnessData.ID = ids
	lightnessData.ExpiredTime = expiredTimes

	// 保存到数据库
	return lm.SaveLightnessData(deviceID, lightnessData)
}

// UpdateLightnessConfig 更新玩家的炫光配置
func (lm *LightnessManager) UpdateLightnessConfig(deviceID string, config int) error {
	// 获取当前炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		return err
	}

	// 更新炫光配置
	lightnessData.Config = config

	// 保存到数据库
	return lm.SaveLightnessData(deviceID, lightnessData)
}

// AddLightness 添加一个新的炫光到玩家拥有的炫光列表
func (lm *LightnessManager) AddLightness(deviceID string, id int, expiredTime int) error {
	// 获取当前炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		return err
	}

	// 检查炫光是否已存在
	for i, lightnessID := range lightnessData.ID {
		if lightnessID == id {
			// 如果已存在，更新过期时间
			lightnessData.ExpiredTime[i] = expiredTime
			return lm.SaveLightnessData(deviceID, lightnessData)
		}
	}

	// 添加新炫光
	lightnessData.ID = append(lightnessData.ID, id)
	lightnessData.ExpiredTime = append(lightnessData.ExpiredTime, expiredTime)

	// 保存到数据库
	return lm.SaveLightnessData(deviceID, lightnessData)
}

// RemoveLightness 从玩家拥有的炫光列表中移除一个炫光
func (lm *LightnessManager) RemoveLightness(deviceID string, id int) error {
	// 获取当前炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除炫光
	found := false
	newIDs := []int{}
	newExpiredTimes := []int{}

	for i, lightnessID := range lightnessData.ID {
		if lightnessID != id {
			newIDs = append(newIDs, lightnessID)
			newExpiredTimes = append(newExpiredTimes, lightnessData.ExpiredTime[i])
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-2, "炫光不存在")
	}

	// 更新炫光数据
	lightnessData.ID = newIDs
	lightnessData.ExpiredTime = newExpiredTimes

	// 保存到数据库
	return lm.SaveLightnessData(deviceID, lightnessData)
}

// GetLightnessIDs 获取玩家拥有的炫光ID列表
func (lm *LightnessManager) GetLightnessIDs(deviceID string) ([]int, []int, error) {
	// 获取炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		return nil, nil, err
	}

	return lightnessData.ID, lightnessData.ExpiredTime, nil
}

// GetLightnessConfig 获取玩家的炫光配置
func (lm *LightnessManager) GetLightnessConfig(deviceID string) (int, error) {
	// 获取炫光数据
	lightnessData, err := lm.GetLightnessData(deviceID)
	if err != nil {
		return 0, err
	}

	return lightnessData.Config, nil
}

// ParseLightnessDataFromJSON 解析JSON格式的炫光数据，返回客户端需要的格式
// 此方法用于直接从数据库中获取的JSON字符串解析为客户端需要的格式
// 避免重复查询数据库
func (lm *LightnessManager) ParseLightnessDataFromJSON(jsonData string) (map[string]interface{}, error) {
	// 如果数据为空，返回默认数据
	if jsonData == "" || jsonData == "null" {
		defaultData := lm.GetDefaultLightnessData()
		// 将默认数据转换为map格式
		result := map[string]interface{}{}
		defaultDataJSON, _ := json.Marshal(defaultData)
		json.Unmarshal(defaultDataJSON, &result)
		return result, nil
	}

	// 解析JSON数据
	var lightnessData LightnessData
	err := json.Unmarshal([]byte(jsonData), &lightnessData)
	if err != nil {
		log.Printf("解析炫光数据失败: %v", err)
		// 解析失败时，使用默认数据
		defaultData := lm.GetDefaultLightnessData()
		// 将默认数据转换为map格式
		result := map[string]interface{}{}
		defaultDataJSON, _ := json.Marshal(defaultData)
		json.Unmarshal(defaultDataJSON, &result)
		return result, nil
	}

	// 将炫光数据转换为map格式，以便在handler中使用
	result := map[string]interface{}{}
	lightnessDataJSON, _ := json.Marshal(lightnessData)
	json.Unmarshal(lightnessDataJSON, &result)

	return result, nil
}
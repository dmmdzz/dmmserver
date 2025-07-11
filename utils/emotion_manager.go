// utils/emotion_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// OwnedEmotion 表示玩家拥有的表情结构
type OwnedEmotion struct {
	ID          []interface{} `json:"id"`          // 表情ID，可以是整数或字符串
	ExpiredTime []int         `json:"expiredTime"` // 过期时间
}

// EmotionConfig 表示角色的表情配置结构
type EmotionConfig struct {
	Character int           `json:"character"` // 角色ID
	Config    []interface{} `json:"config"`    // 表情配置，可以是整数或字符串
}

// EmotionData 表示完整的表情数据结构
type EmotionData struct {
	OwnedIngameEmotion   OwnedEmotion    `json:"ownedIngameEmotion"`   // 拥有的游戏内表情
	IngameEmotionConfigs []EmotionConfig `json:"ingameEmotionConfigs"` // 游戏内表情配置
}

// EmotionManager 提供表情数据的管理功能
type EmotionManager struct {}

// NewEmotionManager 创建一个新的表情管理器
func NewEmotionManager() *EmotionManager {
	return &EmotionManager{}
}

// GetEmotionData 从数据库获取指定设备ID的表情数据
func (em *EmotionManager) GetEmotionData(deviceID string) (*EmotionData, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析表情数据
	var emotionData EmotionData
	if playerData.EmotionData == "" || playerData.EmotionData == "null" {
		// 如果没有表情数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的表情数据为空，使用默认数据", deviceID)
		defaultEmotionData := em.GetDefaultEmotionData()
		
		// 保存默认数据到数据库
		err := em.SaveEmotionData(deviceID, defaultEmotionData)
		if err != nil {
			log.Printf("保存默认表情数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultEmotionData, nil
	}

	err := json.Unmarshal([]byte(playerData.EmotionData), &emotionData)
	if err != nil {
		log.Printf("解析表情数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认表情数据")
		return em.GetDefaultEmotionData(), nil
	}

	return &emotionData, nil
}

// SaveEmotionData 保存表情数据到数据库
func (em *EmotionManager) SaveEmotionData(deviceID string, emotionData *EmotionData) error {
	// 将表情数据序列化为JSON
	emotionDataJSON, err := json.Marshal(emotionData)
	if err != nil {
		log.Printf("序列化表情数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("emotion_data", string(emotionDataJSON))
	if result.Error != nil {
		log.Printf("更新表情数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// GetDefaultEmotionData 创建默认表情数据并返回
func (em *EmotionManager) GetDefaultEmotionData() *EmotionData {
	// 创建默认表情数据
	defaultEmotionData := &EmotionData{
		OwnedIngameEmotion: OwnedEmotion{
			ID:          []interface{}{950001, 960001, 960701},
			ExpiredTime: []int{0, 0, 0},
		},
		IngameEmotionConfigs: []EmotionConfig{
			{
				Character: 100,
				Config:    []interface{}{"950001", "960701", 0, 0, 0, 0},
			},
			{
				Character: 200,
				Config:    []interface{}{"950001", "960701", 0, 0, 0, 0},
			},
			{
				Character: 300,
				Config:    []interface{}{"950001", "960701", 0, 0, 0, 0},
			},
			{
				Character: 400,
				Config:    []interface{}{"950001", "960701", 0, 0, 0, 0},
			},
			{
				Character: 500,
				Config:    []interface{}{"950001", "960701", 0, 0, 0, 0},
			},
		},
	}
	
	return defaultEmotionData
}

// UpdateOwnedEmotions 更新玩家拥有的表情数据
func (em *EmotionManager) UpdateOwnedEmotions(deviceID string, ids []interface{}, expiredTimes []int) error {
	// 获取当前表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return err
	}

	// 更新拥有的表情数据
	emotionData.OwnedIngameEmotion.ID = ids
	emotionData.OwnedIngameEmotion.ExpiredTime = expiredTimes

	// 保存到数据库
	return em.SaveEmotionData(deviceID, emotionData)
}

// GetOwnedEmotions 获取玩家拥有的表情数据
func (em *EmotionManager) GetOwnedEmotions(deviceID string) ([]interface{}, []int, error) {
	// 获取表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return nil, nil, err
	}

	return emotionData.OwnedIngameEmotion.ID, emotionData.OwnedIngameEmotion.ExpiredTime, nil
}

// AddOwnedEmotion 添加一个新的表情到玩家拥有的表情列表
func (em *EmotionManager) AddOwnedEmotion(deviceID string, id interface{}, expiredTime int) error {
	// 获取当前表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return err
	}

	// 检查表情是否已存在
	for i, emotionID := range emotionData.OwnedIngameEmotion.ID {
		if emotionID == id {
			// 如果已存在，更新过期时间
			emotionData.OwnedIngameEmotion.ExpiredTime[i] = expiredTime
			return em.SaveEmotionData(deviceID, emotionData)
		}
	}

	// 添加新表情
	emotionData.OwnedIngameEmotion.ID = append(emotionData.OwnedIngameEmotion.ID, id)
	emotionData.OwnedIngameEmotion.ExpiredTime = append(emotionData.OwnedIngameEmotion.ExpiredTime, expiredTime)

	// 保存到数据库
	return em.SaveEmotionData(deviceID, emotionData)
}

// RemoveOwnedEmotion 从玩家拥有的表情列表中移除一个表情
func (em *EmotionManager) RemoveOwnedEmotion(deviceID string, id interface{}) error {
	// 获取当前表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除表情
	found := false
	newIDs := []interface{}{}
	newExpiredTimes := []int{}

	for i, emotionID := range emotionData.OwnedIngameEmotion.ID {
		if emotionID != id {
			newIDs = append(newIDs, emotionID)
			newExpiredTimes = append(newExpiredTimes, emotionData.OwnedIngameEmotion.ExpiredTime[i])
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-2, "表情不存在")
	}

	// 更新表情数据
	emotionData.OwnedIngameEmotion.ID = newIDs
	emotionData.OwnedIngameEmotion.ExpiredTime = newExpiredTimes

	// 保存到数据库
	return em.SaveEmotionData(deviceID, emotionData)
}

// UpdateEmotionConfig 更新指定角色的表情配置
func (em *EmotionManager) UpdateEmotionConfig(deviceID string, characterID int, config []interface{}) error {
	// 获取当前表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return err
	}

	// 查找并更新角色的表情配置
	found := false
	for i, emotionConfig := range emotionData.IngameEmotionConfigs {
		if emotionConfig.Character == characterID {
			emotionData.IngameEmotionConfigs[i].Config = config
			found = true
			break
		}
	}

	// 如果没有找到角色的配置，添加一个新的配置
	if !found {
		emotionData.IngameEmotionConfigs = append(emotionData.IngameEmotionConfigs, EmotionConfig{
			Character: characterID,
			Config:    config,
		})
	}

	// 保存到数据库
	return em.SaveEmotionData(deviceID, emotionData)
}

// GetEmotionConfig 获取指定角色的表情配置
func (em *EmotionManager) GetEmotionConfig(deviceID string, characterID int) ([]interface{}, error) {
	// 获取表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return nil, err
	}

	// 查找角色的表情配置
	for _, emotionConfig := range emotionData.IngameEmotionConfigs {
		if emotionConfig.Character == characterID {
			return emotionConfig.Config, nil
		}
	}

	// 如果没有找到，返回默认配置
	return []interface{}{0, 0, 0, 0, 0, 0}, nil
}

// GetAllEmotionConfigs 获取所有角色的表情配置
func (em *EmotionManager) GetAllEmotionConfigs(deviceID string) ([]EmotionConfig, error) {
	// 获取表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return nil, err
	}

	return emotionData.IngameEmotionConfigs, nil
}

// RemoveEmotionConfig 移除指定角色的表情配置
func (em *EmotionManager) RemoveEmotionConfig(deviceID string, characterID int) error {
	// 获取当前表情数据
	emotionData, err := em.GetEmotionData(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除角色的表情配置
	found := false
	newConfigs := []EmotionConfig{}

	for _, emotionConfig := range emotionData.IngameEmotionConfigs {
		if emotionConfig.Character != characterID {
			newConfigs = append(newConfigs, emotionConfig)
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-2, "角色表情配置不存在")
	}

	// 更新表情配置
	emotionData.IngameEmotionConfigs = newConfigs

	// 保存到数据库
	return em.SaveEmotionData(deviceID, emotionData)
}

// ParseEmotionDataFromJSON 从JSON字符串解析表情数据
// 此方法可以直接接收数据库中存储的原始JSON字符串，解析后返回EmotionData结构
func (em *EmotionManager) ParseEmotionDataFromJSON(jsonStr string) (*EmotionData, error) {
	// 如果JSON字符串为空或为"null"，返回默认数据
	if jsonStr == "" || jsonStr == "null" {
		log.Printf("表情数据为空，使用默认数据")
		return em.GetDefaultEmotionData(), nil
	}

	// 解析JSON字符串为EmotionData结构
	var emotionData EmotionData
	err := json.Unmarshal([]byte(jsonStr), &emotionData)
	if err != nil {
		log.Printf("解析表情数据失败: %v，使用默认数据", err)
		return em.GetDefaultEmotionData(), nil
	}

	return &emotionData, nil
}

// ConvertEmotionDataToMap 将EmotionData结构转换为map格式，以便在handler中使用
func (em *EmotionManager) ConvertEmotionDataToMap(emotionData *EmotionData) (map[string]interface{}, error) {
	// 将EmotionData结构序列化为JSON
	emotionDataJSON, err := json.Marshal(emotionData)
	if err != nil {
		log.Printf("序列化表情数据失败: %v", err)
		return nil, game_error.New(-2, "数据处理错误")
	}

	// 将JSON反序列化为map
	result := map[string]interface{}{}
	err = json.Unmarshal(emotionDataJSON, &result)
	if err != nil {
		log.Printf("反序列化表情数据失败: %v", err)
		return nil, game_error.New(-2, "数据处理错误")
	}

	return result, nil
}
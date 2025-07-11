// utils/skin_part_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// SkinPart 表示一个皮肤部件的结构
type SkinPart struct {
	SkinPartIDs    string      `json:"skinPartIDs"`    // 皮肤部件ID
	SkinPartColors string      `json:"skinPartColors"` // 皮肤部件颜色
	ExpiredTime    int         `json:"expiredTime"`    // 过期时间
	SkinDecals     interface{} `json:"skinDecals"`     // 皮肤贴花
}

// SkinPartManager 提供皮肤部件数据的管理功能
type SkinPartManager struct {}

// NewSkinPartManager 创建一个新的皮肤部件管理器
func NewSkinPartManager() *SkinPartManager {
	return &SkinPartManager{}
}

// GetSkinParts 从数据库获取指定设备ID的所有皮肤部件数据
func (sm *SkinPartManager) GetSkinParts(deviceID string) ([]SkinPart, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析皮肤部件数据
	var skinParts []SkinPart
	if playerData.OwnedSkins == "" || playerData.OwnedSkins == "null" {
		// 如果没有皮肤部件数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的皮肤部件数据为空，使用默认数据", deviceID)
		defaultSkinParts := sm.GetDefaultSkinParts()
		
		// 保存默认数据到数据库
		err := sm.SaveSkinParts(deviceID, defaultSkinParts)
		if err != nil {
			log.Printf("保存默认皮肤部件数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultSkinParts, nil
	}

	err := json.Unmarshal([]byte(playerData.OwnedSkins), &skinParts)
	if err != nil {
		log.Printf("解析皮肤部件数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认皮肤部件数据")
		return sm.GetDefaultSkinParts(), nil
		// 注意：这里不返回错误，而是使用默认数据
	}

	return skinParts, nil
}

// SaveSkinParts 保存皮肤部件数据到数据库
func (sm *SkinPartManager) SaveSkinParts(deviceID string, skinParts []SkinPart) error {
	// 检查皮肤部件数据是否为空
	if len(skinParts) == 0 {
		// 如果为空，使用空数组而不是空字符串
		log.Printf("皮肤部件数据为空，使用空数组 '[]' 代替")
		result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("owned_skins", "[]")
		if result.Error != nil {
			log.Printf("更新皮肤部件数据失败: %v", result.Error)
			return game_error.New(-2, "数据库更新错误")
		}
		return nil
	}

	// 将皮肤部件数据序列化为JSON
	skinPartsJSON, err := json.Marshal(skinParts)
	if err != nil {
		log.Printf("序列化皮肤部件数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("owned_skins", string(skinPartsJSON))
	if result.Error != nil {
		log.Printf("更新皮肤部件数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// ConvertArraysToSkinParts 将四个独立的数组转换为皮肤部件对象数组
func (sm *SkinPartManager) ConvertArraysToSkinParts(skinPartIDs []string, skinPartColors []string, expiredTimes []int, skinDecals []interface{}) []SkinPart {
	// 确定数组长度，取四个数组中最小的长度
	length := len(skinPartIDs)
	if len(skinPartColors) < length {
		length = len(skinPartColors)
	}
	if len(expiredTimes) < length {
		length = len(expiredTimes)
	}
	if len(skinDecals) < length {
		length = len(skinDecals)
	}

	// 创建皮肤部件对象数组
	skinParts := make([]SkinPart, length)
	for i := 0; i < length; i++ {
		skinParts[i] = SkinPart{
			SkinPartIDs:    skinPartIDs[i],
			SkinPartColors: skinPartColors[i],
			ExpiredTime:    expiredTimes[i],
			SkinDecals:     skinDecals[i],
		}
	}

	return skinParts
}

// UpdateSkinPartData 更新玩家的皮肤部件数据
func (sm *SkinPartManager) UpdateSkinPartData(deviceID string, skinPartIDs []string, skinPartColors []string, expiredTimes []int, skinDecals []interface{}) error {
	// 将四个数组转换为皮肤部件对象数组
	skinParts := sm.ConvertArraysToSkinParts(skinPartIDs, skinPartColors, expiredTimes, skinDecals)

	// 保存到数据库
	return sm.SaveSkinParts(deviceID, skinParts)
}

// ExtractSkinPartArrays 从皮肤部件对象数组中提取四个独立的数组
func (sm *SkinPartManager) ExtractSkinPartArrays(skinParts []SkinPart) ([]string, []string, []int, []interface{}, error) {
	skinPartIDs := make([]string, len(skinParts))
	skinPartColors := make([]string, len(skinParts))
	expiredTimes := make([]int, len(skinParts))
	skinDecals := make([]interface{}, len(skinParts))

	for i, part := range skinParts {
		skinPartIDs[i] = part.SkinPartIDs
		skinPartColors[i] = part.SkinPartColors
		expiredTimes[i] = part.ExpiredTime
		skinDecals[i] = part.SkinDecals
	}

	return skinPartIDs, skinPartColors, expiredTimes, skinDecals, nil
}

// GetDefaultSkinParts 创建默认皮肤部件数据并返回
func (sm *SkinPartManager) GetDefaultSkinParts() []SkinPart {
	// 创建默认皮肤部件数据 - 包含skinPartIDs、skinPartColors、expiredTime和skinDecals属性
	defaultSkinParts := []SkinPart{
		// 基础皮肤数据
		{SkinPartIDs: "1001", SkinPartColors: "67", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "1002", SkinPartColors: "103", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "1003", SkinPartColors: "84", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "1004", SkinPartColors: "110", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "1005", SkinPartColors: "119", ExpiredTime: 0, SkinDecals: 0},
		// 角色2皮肤
		{SkinPartIDs: "2001", SkinPartColors: "12", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "2002", SkinPartColors: "38", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "2003", SkinPartColors: "25", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "2004", SkinPartColors: "51", ExpiredTime: 0, SkinDecals: 0},
		{SkinPartIDs: "2005", SkinPartColors: "54", ExpiredTime: 0, SkinDecals: 0},
	}
	
	return defaultSkinParts
}

// UpdateSkinPartField 更新指定皮肤部件的特定字段
func (sm *SkinPartManager) UpdateSkinPartField(deviceID string, skinPartID string, fieldName string, fieldValue interface{}) error {
	// 获取当前皮肤部件数据
	skinParts, err := sm.GetSkinParts(deviceID)
	if err != nil {
		return err
	}

	// 查找并更新指定皮肤部件的字段
	partFound := false
	for i, part := range skinParts {
		if part.SkinPartIDs == skinPartID {
			switch fieldName {
			case "skinPartColors":
				if colors, ok := fieldValue.(string); ok {
					skinParts[i].SkinPartColors = colors
				} else {
					return game_error.New(-2, "字段类型错误")
				}
			case "expiredTime":
				if expTime, ok := fieldValue.(int); ok {
					skinParts[i].ExpiredTime = expTime
				} else {
					return game_error.New(-2, "字段类型错误")
				}
			case "skinDecals":
				skinParts[i].SkinDecals = fieldValue
			default:
				return game_error.New(-2, "不支持的字段名称")
			}
			partFound = true
			break
		}
	}

	// 如果没有找到指定皮肤部件，则添加新皮肤部件
	if !partFound {
		newPart := SkinPart{SkinPartIDs: skinPartID}
		switch fieldName {
		case "skinPartColors":
			if colors, ok := fieldValue.(string); ok {
				newPart.SkinPartColors = colors
			} else {
				return game_error.New(-2, "字段类型错误")
			}
		case "expiredTime":
			if expTime, ok := fieldValue.(int); ok {
				newPart.ExpiredTime = expTime
			} else {
				return game_error.New(-2, "字段类型错误")
			}
		case "skinDecals":
			newPart.SkinDecals = fieldValue
		default:
			return game_error.New(-2, "不支持的字段名称")
		}
		skinParts = append(skinParts, newPart)
	}

	// 保存更新后的皮肤部件数据
	return sm.SaveSkinParts(deviceID, skinParts)
}

// AddSkinPart 添加新皮肤部件或更新现有皮肤部件
func (sm *SkinPartManager) AddSkinPart(deviceID string, skinPart SkinPart) error {
	// 获取当前皮肤部件数据
	skinParts, err := sm.GetSkinParts(deviceID)
	if err != nil {
		return err
	}

	// 查找是否已存在相同ID的皮肤部件
	partFound := false
	for i, existingPart := range skinParts {
		if existingPart.SkinPartIDs == skinPart.SkinPartIDs {
			// 更新现有皮肤部件
			skinParts[i] = skinPart
			partFound = true
			break
		}
	}

	// 如果没有找到相同ID的皮肤部件，则添加新皮肤部件
	if !partFound {
		skinParts = append(skinParts, skinPart)
	}

	// 保存更新后的皮肤部件数据
	return sm.SaveSkinParts(deviceID, skinParts)
}

// RemoveSkinPart 移除指定ID的皮肤部件
func (sm *SkinPartManager) RemoveSkinPart(deviceID string, skinPartID string) error {
	// 获取当前皮肤部件数据
	skinParts, err := sm.GetSkinParts(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除指定ID的皮肤部件
	found := false
	for i, part := range skinParts {
		if part.SkinPartIDs == skinPartID {
			// 移除皮肤部件（通过切片操作）
			skinParts = append(skinParts[:i], skinParts[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		// 如果没有找到指定ID的皮肤部件，返回错误
		return game_error.New(-2, "未找到指定ID的皮肤部件")
	}

	// 保存更新后的皮肤部件数据
	return sm.SaveSkinParts(deviceID, skinParts)
}

// ParseSkinPartsFromDB 从数据库JSON字符串解析皮肤部件数据并转换为客户端所需格式
func (sm *SkinPartManager) ParseSkinPartsFromDB(dbJSONStr string) ([]string, []string, []int, []interface{}, error) {
	// 如果数据为空，返回默认数据
	if dbJSONStr == "" || dbJSONStr == "null" {
		defaultSkinParts := sm.GetDefaultSkinParts()
		return sm.ExtractSkinPartArrays(defaultSkinParts)
	}

	// 解析数据库中的JSON字符串
	var skinParts []SkinPart
	err := json.Unmarshal([]byte(dbJSONStr), &skinParts)
	if err != nil {
		log.Printf("解析皮肤部件数据失败: %v", err)
		// 解析失败时使用默认数据并返回错误
		return nil, nil, nil, nil, game_error.New(-2, "数据处理错误")
	}

	// 提取并返回客户端所需的四个数组
	return sm.ExtractSkinPartArrays(skinParts)
}

// ParseSkinPartsFromJSON 从JSON字符串解析皮肤部件数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (sm *SkinPartManager) ParseSkinPartsFromJSON(jsonStr string) ([]string, []string, []int, []interface{}, error) {
	// 如果JSON字符串为空或为"null"，返回默认数据
	if jsonStr == "" || jsonStr == "null" {
		log.Printf("皮肤部件数据为空，使用默认数据")
		defaultSkinParts := sm.GetDefaultSkinParts()
		return sm.ExtractSkinPartArrays(defaultSkinParts)
	}

	// 解析JSON字符串为SkinPart结构数组
	var skinParts []SkinPart
	err := json.Unmarshal([]byte(jsonStr), &skinParts)
	if err != nil {
		log.Printf("解析皮肤部件数据失败: %v，使用默认数据", err)
		// 解析失败时，使用默认数据
		defaultSkinParts := sm.GetDefaultSkinParts()
		return sm.ExtractSkinPartArrays(defaultSkinParts)
	}

	// 从皮肤部件数据中提取四个独立的数组
	return sm.ExtractSkinPartArrays(skinParts)
}
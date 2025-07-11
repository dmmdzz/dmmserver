// utils/character_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/model"
	"dmmserver/db"
	"dmmserver/game_error"
)

// Character 表示一个游戏角色的结构
type Character struct {
	CharacterID        int                    `json:"characterID"`
	ExpiredTime        int                    `json:"expiredTime"`
	CurrentSkinInfo    CharacterSkinInfo      `json:"currentSkinInfo"`
	ExpLevel           int                    `json:"ExpLevel"`
	ExpPoint           int                    `json:"ExpPoint"`
	TalentPointRemained int                   `json:"TalentPointRemained"`
	TalentLevels       []int                  `json:"TalentLevels"`
	WeaponSkinID       int                    `json:"weaponSkinID"`
}

// CharacterSkinInfo 表示角色皮肤信息的结构
type CharacterSkinInfo struct {
	SkinPartIDs     []interface{} `json:"skinPartIDs"`     // 可以是字符串或整数
	SkinPartColors  []string      `json:"skinPartColors"`
	SkinDecals      []int         `json:"skinDecals"`
	ExpiredTime     []int         `json:"expiredTime"`
}

// CharacterManager 提供角色数据的管理功能
type CharacterManager struct {}

// NewCharacterManager 创建一个新的角色管理器
func NewCharacterManager() *CharacterManager {
	return &CharacterManager{}
}

// GetCharacters 从数据库获取指定设备ID的所有角色数据
func (cm *CharacterManager) GetCharacters(deviceID string) ([]Character, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, result.Error
	}

	// 解析角色数据
	var characters []Character
	if playerData.OwnedCharacters == "" || playerData.OwnedCharacters == "null" {
		// 如果没有角色数据，返回空数组
		return []Character{}, nil
	}

	err := json.Unmarshal([]byte(playerData.OwnedCharacters), &characters)
	if err != nil {
		log.Printf("解析角色数据失败: %v", err)
		return nil, err
	}

	return characters, nil
}

// GetCharacterByID 获取指定设备ID和角色ID的角色数据
func (cm *CharacterManager) GetCharacterByID(deviceID string, characterID int) (*Character, error) {
	characters, err := cm.GetCharacters(deviceID)
	if err != nil {
		return nil, err
	}

	for i := range characters {
		if characters[i].CharacterID == characterID {
			return &characters[i], nil
		}
	}

	return nil, game_error.New(-31, "角色不存在")
}

// SaveCharacters 保存角色数据到数据库
func (cm *CharacterManager) SaveCharacters(deviceID string, characters []Character) error {
	// 将角色数据序列化为JSON
	charactersJSON, err := json.Marshal(characters)
	if err != nil {
		log.Printf("序列化角色数据失败: %v", err)
		return err
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("owned_characters", string(charactersJSON))
	if result.Error != nil {
		log.Printf("更新角色数据失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// AddCharacter 添加一个新角色
func (cm *CharacterManager) AddCharacter(deviceID string, character Character) error {
	characters, err := cm.GetCharacters(deviceID)
	if err != nil {
		return err
	}

	// 检查角色是否已存在
	for i := range characters {
		if characters[i].CharacterID == character.CharacterID {
			return game_error.New(-8, "角色已存在")
		}
	}

	// 添加新角色
	characters = append(characters, character)

	// 保存到数据库
	return cm.SaveCharacters(deviceID, characters)
}

// UpdateCharacter 更新角色数据
func (cm *CharacterManager) UpdateCharacter(deviceID string, character Character) error {
	characters, err := cm.GetCharacters(deviceID)
	if err != nil {
		return err
	}

	// 查找并更新角色
	found := false
	for i := range characters {
		if characters[i].CharacterID == character.CharacterID {
			characters[i] = character
			found = true
			break
		}
	}

	if !found {
		return game_error.New(-31, "角色不存在")
	}

	// 保存到数据库
	return cm.SaveCharacters(deviceID, characters)
}

// DeleteCharacter 删除角色
func (cm *CharacterManager) DeleteCharacter(deviceID string, characterID int) error {
	characters, err := cm.GetCharacters(deviceID)
	if err != nil {
		return err
	}

	// 查找并删除角色
	found := false
	newCharacters := []Character{}
	for i := range characters {
		if characters[i].CharacterID != characterID {
			newCharacters = append(newCharacters, characters[i])
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-31, "角色不存在")
	}

	// 保存到数据库
	return cm.SaveCharacters(deviceID, newCharacters)
}

// ParseCharactersFromDB 从数据库JSON字符串解析角色数据
func (cm *CharacterManager) ParseCharactersFromDB(dbJSONStr string) ([]Character, error) {
	// 如果数据为空，返回空数组
	if dbJSONStr == "" || dbJSONStr == "null" {
		return []Character{}, nil
	}

	// 解析数据库中的JSON字符串
	var characters []Character
	err := json.Unmarshal([]byte(dbJSONStr), &characters)
	if err != nil {
		log.Printf("解析角色数据失败: %v", err)
		return nil, err
	}

	return characters, nil
}

// ParseCharactersFromJSON 从JSON字符串解析角色数据并返回客户端需要的格式
func (cm *CharacterManager) ParseCharactersFromJSON(charactersJSON string) ([]Character, error) {
	// 如果数据为空，使用默认数据
	if charactersJSON == "" || charactersJSON == "null" {
		log.Printf("角色数据为空，使用默认数据")
		defaultCharacters := cm.GetDefaultCharacters()
		return defaultCharacters, nil
	}

	// 解析JSON字符串
	var characters []Character
	err := json.Unmarshal([]byte(charactersJSON), &characters)
	if err != nil {
		log.Printf("解析角色数据失败: %v，使用默认值", err)
		// 解析失败时，使用默认数据
		defaultCharacters := cm.GetDefaultCharacters()
		return defaultCharacters, nil
	}

	// 返回客户端需要的格式
	return characters, nil
}

// GetDefaultCharacter 获取默认的角色数据
func (cm *CharacterManager) GetDefaultCharacter(characterID int) Character {
	// 根据角色ID返回默认配置
	switch characterID {
	case 100:
		return Character{
			CharacterID: 100,
			ExpiredTime: 0,
			CurrentSkinInfo: CharacterSkinInfo{
				SkinPartIDs:    []interface{}{"1001", "1002", "1003", "1004", "1005"},
				SkinPartColors: []string{"67", "103", "84", "110", "119"},
				SkinDecals:     []int{0, 0, 0, 0, 0},
				ExpiredTime:    []int{0, 0, 0, 0, 0},
			},
			ExpLevel:           999999999,
			ExpPoint:           0,
			TalentPointRemained: 0,
			TalentLevels:       []int{4, 4, 4},
			WeaponSkinID:       0,
		}
	case 200:
		return Character{
			CharacterID: 200,
			ExpiredTime: 0,
			CurrentSkinInfo: CharacterSkinInfo{
				SkinPartIDs:    []interface{}{2001, 2002, 2003, 2004, 2005},
				SkinPartColors: []string{"12", "38", "25", "51", "54"},
				SkinDecals:     []int{0, 0, 0, 0, 0},
				ExpiredTime:    []int{0, 0, 0, 0, 0},
			},
			ExpLevel:           9999,
			ExpPoint:           127,
			TalentPointRemained: 2,
			TalentLevels:       []int{99999, 99999, 99999},
			WeaponSkinID:       0,
		}
	default:
		// 默认角色配置
		return Character{
			CharacterID: characterID,
			ExpiredTime: 0,
			CurrentSkinInfo: CharacterSkinInfo{
				SkinPartIDs:    []interface{}{},
				SkinPartColors: []string{},
				SkinDecals:     []int{},
				ExpiredTime:    []int{},
			},
			ExpLevel:           1,
			ExpPoint:           0,
			TalentPointRemained: 0,
			TalentLevels:       []int{1, 1, 1},
			WeaponSkinID:       0,
		}
	}
}

// GetDefaultCharacters 获取默认的角色数据列表
func (cm *CharacterManager) GetDefaultCharacters() []Character {
	return []Character{
		cm.GetDefaultCharacter(100),
		cm.GetDefaultCharacter(200),
	}
}
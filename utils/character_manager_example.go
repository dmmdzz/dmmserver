// utils/character_manager_example.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/model"
)

// 示例：如何在handler中使用CharacterManager

// 示例1：获取玩家的所有角色数据
func ExampleGetAllCharacters(deviceID string) ([]map[string]interface{}, error) {
	// 创建角色管理器
	cm := NewCharacterManager()

	// 获取角色数据
	characters, err := cm.GetCharacters(deviceID)
	if err != nil {
		log.Printf("获取角色数据失败: %v", err)
		return nil, err
	}

	// 将角色数据转换为map格式，以便在handler中使用
	result := make([]map[string]interface{}, 0, len(characters))
	for _, character := range characters {
		characterMap := map[string]interface{}{}
		characterJSON, _ := json.Marshal(character)
		json.Unmarshal(characterJSON, &characterMap)
		result = append(result, characterMap)
	}

	return result, nil
}

// 示例2：更新角色的经验值
func ExampleUpdateCharacterExp(deviceID string, characterID int, expLevel int, expPoint int) error {
	// 创建角色管理器
	cm := NewCharacterManager()

	// 获取指定角色
	character, err := cm.GetCharacterByID(deviceID, characterID)
	if err != nil {
		log.Printf("获取角色数据失败: %v", err)
		return err
	}

	// 更新经验值
	character.ExpLevel = expLevel
	character.ExpPoint = expPoint

	// 保存更新后的角色数据
	return cm.UpdateCharacter(deviceID, *character)
}

// 示例3：为新玩家创建默认角色
func ExampleCreateDefaultCharactersForNewPlayer(playerData *model.PlayerData) error {
	// 创建角色管理器
	cm := NewCharacterManager()

	// 获取默认角色列表
	defaultCharacters := cm.GetDefaultCharacters()

	// 将角色数据序列化为JSON并保存到PlayerData
	charactersJSON, err := json.Marshal(defaultCharacters)
	if err != nil {
		log.Printf("序列化角色数据失败: %v", err)
		return err
	}

	// 更新PlayerData
	playerData.OwnedCharacters = string(charactersJSON)

	return nil
}

// 示例4：更新角色皮肤
func ExampleUpdateCharacterSkin(deviceID string, characterID int, skinPartIDs []interface{}, skinPartColors []string) error {
	// 创建角色管理器
	cm := NewCharacterManager()

	// 获取指定角色
	character, err := cm.GetCharacterByID(deviceID, characterID)
	if err != nil {
		log.Printf("获取角色数据失败: %v", err)
		return err
	}

	// 更新皮肤信息
	character.CurrentSkinInfo.SkinPartIDs = skinPartIDs
	character.CurrentSkinInfo.SkinPartColors = skinPartColors

	// 保存更新后的角色数据
	return cm.UpdateCharacter(deviceID, *character)
}
// utils/assets_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// Asset 表示玩家拥有的单个资产结构
type Asset struct {
	ItemID    int `json:"itemID"`    // 物品ID
	ItemCount int `json:"itemCount"` // 物品数量
}

// AssetsData 表示完整的资产数据结构
type AssetsData struct {
	OwnedAssets []Asset `json:"ownedAssets"` // 拥有的资产列表
}

// AssetsManager 提供资产数据的管理功能
type AssetsManager struct {}

// NewAssetsManager 创建一个新的资产管理器
func NewAssetsManager() *AssetsManager {
	return &AssetsManager{}
}

// GetAssetsData 从数据库获取指定设备ID的资产数据
func (am *AssetsManager) GetAssetsData(deviceID string) (*AssetsData, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析资产数据
	var assetsData AssetsData
	if playerData.AssetsData == "" || playerData.AssetsData == "null" {
		// 如果没有资产数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的资产数据为空，使用默认数据", deviceID)
		defaultAssetsData := am.GetDefaultAssetsData()
		
		// 保存默认数据到数据库
		err := am.SaveAssetsData(deviceID, defaultAssetsData)
		if err != nil {
			log.Printf("保存默认资产数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultAssetsData, nil
	}

	err := json.Unmarshal([]byte(playerData.AssetsData), &assetsData)
	if err != nil {
		log.Printf("解析资产数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认资产数据")
		return am.GetDefaultAssetsData(), nil
	}

	return &assetsData, nil
}

// SaveAssetsData 保存资产数据到数据库
func (am *AssetsManager) SaveAssetsData(deviceID string, assetsData *AssetsData) error {
	// 将资产数据序列化为JSON
	assetsDataJSON, err := json.Marshal(assetsData)
	if err != nil {
		log.Printf("序列化资产数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("assets_data", string(assetsDataJSON))
	if result.Error != nil {
		log.Printf("更新资产数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// GetDefaultAssetsData 创建默认资产数据并返回
func (am *AssetsManager) GetDefaultAssetsData() *AssetsData {
	// 创建默认资产数据
	defaultAssetsData := &AssetsData{
		OwnedAssets: []Asset{
			{ItemID: 29, ItemCount: 4},
			{ItemID: 30, ItemCount: 20},
			{ItemID: 55, ItemCount: 2},
		},
	}
	
	return defaultAssetsData
}

// UpdateAssets 更新玩家拥有的资产数据
func (am *AssetsManager) UpdateAssets(deviceID string, assets []Asset) error {
	// 获取当前资产数据
	assetsData, err := am.GetAssetsData(deviceID)
	if err != nil {
		return err
	}

	// 更新拥有的资产数据
	assetsData.OwnedAssets = assets

	// 保存到数据库
	return am.SaveAssetsData(deviceID, assetsData)
}

// GetAssets 获取玩家拥有的资产数据
func (am *AssetsManager) GetAssets(deviceID string) ([]Asset, error) {
	// 获取资产数据
	assetsData, err := am.GetAssetsData(deviceID)
	if err != nil {
		return nil, err
	}

	return assetsData.OwnedAssets, nil
}

// AddAsset 添加一个新的资产到玩家拥有的资产列表或更新现有资产数量
func (am *AssetsManager) AddAsset(deviceID string, itemID int, itemCount int) error {
	// 获取当前资产数据
	assetsData, err := am.GetAssetsData(deviceID)
	if err != nil {
		return err
	}

	// 检查资产是否已存在
	for i, asset := range assetsData.OwnedAssets {
		if asset.ItemID == itemID {
			// 如果已存在，更新数量
			assetsData.OwnedAssets[i].ItemCount += itemCount
			return am.SaveAssetsData(deviceID, assetsData)
		}
	}

	// 添加新资产
	assetsData.OwnedAssets = append(assetsData.OwnedAssets, Asset{ItemID: itemID, ItemCount: itemCount})

	// 保存到数据库
	return am.SaveAssetsData(deviceID, assetsData)
}

// UpdateAssetCount 更新指定资产的数量
func (am *AssetsManager) UpdateAssetCount(deviceID string, itemID int, itemCount int) error {
	// 获取当前资产数据
	assetsData, err := am.GetAssetsData(deviceID)
	if err != nil {
		return err
	}

	// 查找并更新资产数量
	found := false
	for i, asset := range assetsData.OwnedAssets {
		if asset.ItemID == itemID {
			assetsData.OwnedAssets[i].ItemCount = itemCount
			found = true
			break
		}
	}

	if !found {
		// 如果资产不存在，添加新资产
		assetsData.OwnedAssets = append(assetsData.OwnedAssets, Asset{ItemID: itemID, ItemCount: itemCount})
	}

	// 保存到数据库
	return am.SaveAssetsData(deviceID, assetsData)
}

// RemoveAsset 从玩家拥有的资产列表中移除一个资产
func (am *AssetsManager) RemoveAsset(deviceID string, itemID int) error {
	// 获取当前资产数据
	assetsData, err := am.GetAssetsData(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除资产
	found := false
	newAssets := []Asset{}

	for _, asset := range assetsData.OwnedAssets {
		if asset.ItemID != itemID {
			newAssets = append(newAssets, asset)
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-2, "资产不存在")
	}

	// 更新资产数据
	assetsData.OwnedAssets = newAssets

	// 保存到数据库
	return am.SaveAssetsData(deviceID, assetsData)
}

// GetAssetCount 获取指定资产的数量
func (am *AssetsManager) GetAssetCount(deviceID string, itemID int) (int, error) {
	// 获取当前资产数据
	assetsData, err := am.GetAssetsData(deviceID)
	if err != nil {
		return 0, err
	}

	// 查找资产
	for _, asset := range assetsData.OwnedAssets {
		if asset.ItemID == itemID {
			return asset.ItemCount, nil
		}
	}

	// 资产不存在
	return 0, nil
}

// ParseAssetsFromJSON 从JSON字符串解析资产数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (am *AssetsManager) ParseAssetsFromJSON(jsonStr string) ([]Asset, error) {
	// 如果JSON字符串为空或为"null"，返回默认数据
	if jsonStr == "" || jsonStr == "null" {
		log.Printf("资产数据为空，使用默认数据")
		defaultAssetsData := am.GetDefaultAssetsData()
		return defaultAssetsData.OwnedAssets, nil
	}

	// 解析JSON字符串为AssetsData结构
	var assetsData AssetsData
	err := json.Unmarshal([]byte(jsonStr), &assetsData)
	if err != nil {
		log.Printf("解析资产数据失败: %v", err)
		// 解析失败时，使用默认数据
		log.Printf("使用默认资产数据")
		defaultAssetsData := am.GetDefaultAssetsData()
		return defaultAssetsData.OwnedAssets, nil
	}

	return assetsData.OwnedAssets, nil
}
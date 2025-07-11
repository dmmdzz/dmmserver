// utils/assets_manager_example.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/game_error"
)

// 示例1：获取玩家的资产数据
func ExampleGetAssetsData(deviceID string) ([]map[string]interface{}, error) {
	// 创建资产管理器
	am := NewAssetsManager()

	// 获取资产数据
	assets, err := am.GetAssets(deviceID)
	if err != nil {
		// 检查是否为GameError类型
		if gameErr, ok := err.(*game_error.GameError); ok {
			log.Printf("获取资产数据失败: 错误码 %d, 错误信息: %s", gameErr.Code, gameErr.Message)
		} else {
			log.Printf("获取资产数据失败: %v", err)
		}
		return nil, err
	}

	// 将资产数据转换为map格式，以便在handler中使用
	result := make([]map[string]interface{}, 0, len(assets))
	for _, asset := range assets {
		assetMap := map[string]interface{}{}
		assetJSON, _ := json.Marshal(asset)
		json.Unmarshal(assetJSON, &assetMap)
		result = append(result, assetMap)
	}

	return result, nil
}

// 示例2：更新玩家的资产数据
func ExampleUpdateAssetsData(deviceID string, itemIDs []int, itemCounts []int) error {
	// 创建资产管理器
	am := NewAssetsManager()

	// 检查参数长度是否一致
	if len(itemIDs) != len(itemCounts) {
		log.Printf("参数长度不一致: itemIDs=%d, itemCounts=%d", len(itemIDs), len(itemCounts))
		return game_error.New(-2, "参数错误")
	}

	// 构建资产列表
	assets := make([]Asset, 0, len(itemIDs))
	for i, itemID := range itemIDs {
		assets = append(assets, Asset{
			ItemID:    itemID,
			ItemCount: itemCounts[i],
		})
	}

	// 更新资产数据
	err := am.UpdateAssets(deviceID, assets)
	if err != nil {
		log.Printf("更新资产数据失败: %v", err)
		return err
	}

	log.Printf("成功更新玩家 %s 的资产数据", deviceID)
	return nil
}

// 示例3：添加新资产或更新现有资产数量
func ExampleAddAsset(deviceID string, itemID int, itemCount int) error {
	// 创建资产管理器
	am := NewAssetsManager()

	// 添加资产
	err := am.AddAsset(deviceID, itemID, itemCount)
	if err != nil {
		log.Printf("添加资产失败: %v", err)
		return err
	}

	log.Printf("成功为玩家 %s 添加资产 %d，数量 %d", deviceID, itemID, itemCount)
	return nil
}

// 示例4：更新指定资产的数量
func ExampleUpdateAssetCount(deviceID string, itemID int, itemCount int) error {
	// 创建资产管理器
	am := NewAssetsManager()

	// 更新资产数量
	err := am.UpdateAssetCount(deviceID, itemID, itemCount)
	if err != nil {
		log.Printf("更新资产数量失败: %v", err)
		return err
	}

	log.Printf("成功将玩家 %s 的资产 %d 数量更新为 %d", deviceID, itemID, itemCount)
	return nil
}

// 示例5：移除指定资产
func ExampleRemoveAsset(deviceID string, itemID int) error {
	// 创建资产管理器
	am := NewAssetsManager()

	// 移除资产
	err := am.RemoveAsset(deviceID, itemID)
	if err != nil {
		log.Printf("移除资产失败: %v", err)
		return err
	}

	log.Printf("成功从玩家 %s 的资产列表中移除资产 %d", deviceID, itemID)
	return nil
}

// 示例6：获取指定资产的数量
func ExampleGetAssetCount(deviceID string, itemID int) (int, error) {
	// 创建资产管理器
	am := NewAssetsManager()

	// 获取资产数量
	count, err := am.GetAssetCount(deviceID, itemID)
	if err != nil {
		log.Printf("获取资产数量失败: %v", err)
		return 0, err
	}

	log.Printf("玩家 %s 的资产 %d 数量为 %d", deviceID, itemID, count)
	return count, nil
}
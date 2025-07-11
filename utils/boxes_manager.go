// utils/boxes_manager.go
package utils

import (
	"encoding/json"
	"log"

	"dmmserver/db"
	"dmmserver/model"
	"dmmserver/game_error"
)

// OwnedHeadBox 表示玩家拥有的头像框结构
type OwnedHeadBox struct {
	HeadBoxID   []int `json:"headBoxID"`   // 头像框ID
	ExpiredTime []int `json:"expiredTime"` // 过期时间
}

// OwnedBubbleBox 表示玩家拥有的聊天气泡结构
type OwnedBubbleBox struct {
	BubbleBoxID []int `json:"bubbleBoxID"` // 聊天气泡ID
	ExpiredTime []int `json:"expiredTime"` // 过期时间
}

// BoxesData 表示完整的装饰框数据结构
type BoxesData struct {
	OwnedHeadBoxes   OwnedHeadBox   `json:"ownedHeadBoxes"`   // 拥有的头像框
	OwnedBubbleBoxes OwnedBubbleBox `json:"ownedBubbleBoxes"` // 拥有的聊天气泡
}

// BoxesManager 提供装饰框数据的管理功能
type BoxesManager struct {}

// NewBoxesManager 创建一个新的装饰框管理器
func NewBoxesManager() *BoxesManager {
	return &BoxesManager{}
}

// GetBoxesData 从数据库获取指定设备ID的装饰框数据
func (bm *BoxesManager) GetBoxesData(deviceID string) (*BoxesData, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析装饰框数据
	var boxesData BoxesData
	if playerData.BoxesData == "" || playerData.BoxesData == "null" {
		// 如果没有装饰框数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的装饰框数据为空，使用默认数据", deviceID)
		defaultBoxesData := bm.GetDefaultBoxesData()
		
		// 保存默认数据到数据库
		err := bm.SaveBoxesData(deviceID, defaultBoxesData)
		if err != nil {
			log.Printf("保存默认装饰框数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}
		
		return defaultBoxesData, nil
	}

	err := json.Unmarshal([]byte(playerData.BoxesData), &boxesData)
	if err != nil {
		log.Printf("解析装饰框数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认装饰框数据")
		return bm.GetDefaultBoxesData(), nil
	}

	return &boxesData, nil
}

// SaveBoxesData 保存装饰框数据到数据库
func (bm *BoxesManager) SaveBoxesData(deviceID string, boxesData *BoxesData) error {
	// 将装饰框数据序列化为JSON
	boxesDataJSON, err := json.Marshal(boxesData)
	if err != nil {
		log.Printf("序列化装饰框数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("boxes_data", string(boxesDataJSON))
	if result.Error != nil {
		log.Printf("更新装饰框数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// GetDefaultBoxesData 创建默认装饰框数据并返回
func (bm *BoxesManager) GetDefaultBoxesData() *BoxesData {
	// 创建默认装饰框数据
	defaultBoxesData := &BoxesData{
		OwnedHeadBoxes: OwnedHeadBox{
			HeadBoxID:   []int{900001},
			ExpiredTime: []int{0},
		},
		OwnedBubbleBoxes: OwnedBubbleBox{
			BubbleBoxID: []int{910001},
			ExpiredTime: []int{0},
		},
	}
	
	return defaultBoxesData
}

// 头像框相关方法
// ===============================

// UpdateHeadBoxes 更新玩家拥有的头像框数据
func (bm *BoxesManager) UpdateHeadBoxes(deviceID string, headBoxIDs []int, expiredTimes []int) error {
	// 获取当前装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return err
	}

	// 更新拥有的头像框数据
	boxesData.OwnedHeadBoxes.HeadBoxID = headBoxIDs
	boxesData.OwnedHeadBoxes.ExpiredTime = expiredTimes

	// 保存到数据库
	return bm.SaveBoxesData(deviceID, boxesData)
}

// GetHeadBoxes 获取玩家拥有的头像框数据
func (bm *BoxesManager) GetHeadBoxes(deviceID string) ([]int, []int, error) {
	// 获取装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return nil, nil, err
	}

	return boxesData.OwnedHeadBoxes.HeadBoxID, boxesData.OwnedHeadBoxes.ExpiredTime, nil
}

// AddHeadBox 添加一个新的头像框到玩家拥有的头像框列表
func (bm *BoxesManager) AddHeadBox(deviceID string, headBoxID int, expiredTime int) error {
	// 获取当前装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return err
	}

	// 检查头像框是否已存在
	for i, id := range boxesData.OwnedHeadBoxes.HeadBoxID {
		if id == headBoxID {
			// 如果已存在，更新过期时间
			boxesData.OwnedHeadBoxes.ExpiredTime[i] = expiredTime
			return bm.SaveBoxesData(deviceID, boxesData)
		}
	}

	// 添加新头像框
	boxesData.OwnedHeadBoxes.HeadBoxID = append(boxesData.OwnedHeadBoxes.HeadBoxID, headBoxID)
	boxesData.OwnedHeadBoxes.ExpiredTime = append(boxesData.OwnedHeadBoxes.ExpiredTime, expiredTime)

	// 保存到数据库
	return bm.SaveBoxesData(deviceID, boxesData)
}

// RemoveHeadBox 从玩家拥有的头像框列表中移除一个头像框
func (bm *BoxesManager) RemoveHeadBox(deviceID string, headBoxID int) error {
	// 获取当前装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除头像框
	found := false
	newIDs := []int{}
	newExpiredTimes := []int{}

	for i, id := range boxesData.OwnedHeadBoxes.HeadBoxID {
		if id != headBoxID {
			newIDs = append(newIDs, id)
			newExpiredTimes = append(newExpiredTimes, boxesData.OwnedHeadBoxes.ExpiredTime[i])
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-2, "头像框不存在")
	}

	// 更新头像框数据
	boxesData.OwnedHeadBoxes.HeadBoxID = newIDs
	boxesData.OwnedHeadBoxes.ExpiredTime = newExpiredTimes

	// 保存到数据库
	return bm.SaveBoxesData(deviceID, boxesData)
}

// 聊天气泡相关方法
// ===============================

// UpdateBubbleBoxes 更新玩家拥有的聊天气泡数据
func (bm *BoxesManager) UpdateBubbleBoxes(deviceID string, bubbleBoxIDs []int, expiredTimes []int) error {
	// 获取当前装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return err
	}

	// 更新拥有的聊天气泡数据
	boxesData.OwnedBubbleBoxes.BubbleBoxID = bubbleBoxIDs
	boxesData.OwnedBubbleBoxes.ExpiredTime = expiredTimes

	// 保存到数据库
	return bm.SaveBoxesData(deviceID, boxesData)
}

// GetBubbleBoxes 获取玩家拥有的聊天气泡数据
func (bm *BoxesManager) GetBubbleBoxes(deviceID string) ([]int, []int, error) {
	// 获取装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return nil, nil, err
	}

	return boxesData.OwnedBubbleBoxes.BubbleBoxID, boxesData.OwnedBubbleBoxes.ExpiredTime, nil
}

// AddBubbleBox 添加一个新的聊天气泡到玩家拥有的聊天气泡列表
func (bm *BoxesManager) AddBubbleBox(deviceID string, bubbleBoxID int, expiredTime int) error {
	// 获取当前装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return err
	}

	// 检查聊天气泡是否已存在
	for i, id := range boxesData.OwnedBubbleBoxes.BubbleBoxID {
		if id == bubbleBoxID {
			// 如果已存在，更新过期时间
			boxesData.OwnedBubbleBoxes.ExpiredTime[i] = expiredTime
			return bm.SaveBoxesData(deviceID, boxesData)
		}
	}

	// 添加新聊天气泡
	boxesData.OwnedBubbleBoxes.BubbleBoxID = append(boxesData.OwnedBubbleBoxes.BubbleBoxID, bubbleBoxID)
	boxesData.OwnedBubbleBoxes.ExpiredTime = append(boxesData.OwnedBubbleBoxes.ExpiredTime, expiredTime)

	// 保存到数据库
	return bm.SaveBoxesData(deviceID, boxesData)
}

// RemoveBubbleBox 从玩家拥有的聊天气泡列表中移除一个聊天气泡
func (bm *BoxesManager) RemoveBubbleBox(deviceID string, bubbleBoxID int) error {
	// 获取当前装饰框数据
	boxesData, err := bm.GetBoxesData(deviceID)
	if err != nil {
		return err
	}

	// 查找并移除聊天气泡
	found := false
	newIDs := []int{}
	newExpiredTimes := []int{}

	for i, id := range boxesData.OwnedBubbleBoxes.BubbleBoxID {
		if id != bubbleBoxID {
			newIDs = append(newIDs, id)
			newExpiredTimes = append(newExpiredTimes, boxesData.OwnedBubbleBoxes.ExpiredTime[i])
		} else {
			found = true
		}
	}

	if !found {
		return game_error.New(-2, "聊天气泡不存在")
	}

	// 更新聊天气泡数据
	boxesData.OwnedBubbleBoxes.BubbleBoxID = newIDs
	boxesData.OwnedBubbleBoxes.ExpiredTime = newExpiredTimes

	// 保存到数据库
	return bm.SaveBoxesData(deviceID, boxesData)
}

// ParseBoxesDataFromJSON 从JSON字符串解析装饰框数据
// 此方法可以直接接收数据库中存储的原始JSON字符串，解析后返回BoxesData结构
func (bm *BoxesManager) ParseBoxesDataFromJSON(jsonStr string) (*BoxesData, error) {
	// 如果JSON字符串为空或为"null"，返回默认数据
	if jsonStr == "" || jsonStr == "null" {
		log.Printf("装饰框数据为空，使用默认数据")
		return bm.GetDefaultBoxesData(), nil
	}

	// 解析JSON字符串为BoxesData结构
	var boxesData BoxesData
	err := json.Unmarshal([]byte(jsonStr), &boxesData)
	if err != nil {
		log.Printf("解析装饰框数据失败: %v，使用默认数据", err)
		return bm.GetDefaultBoxesData(), nil
	}

	return &boxesData, nil
}



// ConvertToClientFormat 将数据库格式的装饰框数据转换为客户端需要的格式
// 此方法直接接收数据库中存储的JSON字符串，解析后返回客户端需要的格式
func (bm *BoxesManager) ConvertToClientFormat(boxesDataJSON string) (map[string]interface{}, error) {
	// 解析装饰框数据
	boxesData, err := bm.ParseBoxesDataFromJSON(boxesDataJSON)
	if err != nil {
		log.Printf("解析装饰框数据失败: %v", err)
		return nil, err
	}

	// 构建客户端需要的格式
	result := map[string]interface{}{
		"ownedHeadBoxes": map[string]interface{}{
			"headBoxID":   boxesData.OwnedHeadBoxes.HeadBoxID,
			"expiredTime": boxesData.OwnedHeadBoxes.ExpiredTime,
		},
		"ownedBubbleBoxes": map[string]interface{}{
			"bubbleBoxID": boxesData.OwnedBubbleBoxes.BubbleBoxID,
			"expiredTime": boxesData.OwnedBubbleBoxes.ExpiredTime,
		},
	}

	return result, nil
}
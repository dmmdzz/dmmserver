// utils/public_info_manager.go
package utils

import (
//	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
)

// PublicInfo 表示玩家公开信息的结构
type PublicInfo struct {
	Name              string           `json:"name"`
	Icon              interface{}      `json:"icon"` // 可以是整数或字符串
	Sex               string           `json:"sex"`
	Age               string           `json:"age"`
	Area              string           `json:"area"`
	Province          string           `json:"province"`
	GradeThief        string           `json:"gradeThief"`
	GradePointThief   string           `json:"gradePointThief"`
	GradePolice       string           `json:"gradePolice"`
	GradePointPolice  string           `json:"gradePointPolice"`
	Gm                bool             `json:"gm"`
	MembershipInfo    []MembershipInfo `json:"membershipInfo"`
	ActiveHeadBoxID   int              `json:"activeHeadBoxID"`
	ActiveBubbleBoxID int              `json:"activeBubbleBoxID"`
	UnionID           int              `json:"unionID"`
	UnionName         string           `json:"unionName"`
	UnionBadge        UnionBadge       `json:"unionBadge"`
	IP                string           `json:"ip,omitempty"` // 实时生成，不存储
}

// MembershipInfo 表示会员信息的结构
type MembershipInfo struct {
	MembershipID int `json:"membershipID"`
	ExpireTime   int `json:"expireTime"`
}

// UnionBadge 表示公会徽章的结构
type UnionBadge struct {
	IconID  int `json:"iconID"`
	FrameID int `json:"frameID"`
	FlageID int `json:"flageID"` // 注意：原始数据中是flageID而不是flagID
}

// PublicInfoManager 提供玩家公开信息的管理功能
type PublicInfoManager struct{}

// NewPublicInfoManager 创建一个新的玩家公开信息管理器
func NewPublicInfoManager() *PublicInfoManager {
	return &PublicInfoManager{}
}

// GetPublicInfo 从数据库获取指定设备ID的玩家公开信息
func (pm *PublicInfoManager) GetPublicInfo(deviceID string) (*PublicInfo, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析玩家公开信息数据
	var publicInfo PublicInfo
	if playerData.PublicInfo == "" || playerData.PublicInfo == "null" {
		// 如果没有公开信息数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的公开信息数据为空，使用默认数据", deviceID)
		publicInfo = pm.GetDefaultPublicInfo()

		// 保存默认数据到数据库
		err := pm.SavePublicInfo(deviceID, &publicInfo)
		if err != nil {
			log.Printf("保存默认公开信息数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}

		return &publicInfo, nil
	}

	// 解析键值对格式
	publicInfo, err := pm.ParseKeyValuePublicInfo(playerData.PublicInfo)
	if err != nil {
		log.Printf("解析公开信息数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认公开信息数据")
		defaultPublicInfo := pm.GetDefaultPublicInfo()
		return &defaultPublicInfo, nil
	}

	return &publicInfo, nil
}

// GetPublicInfoByRoleID 通过角色ID获取玩家公开信息
func (pm *PublicInfoManager) GetPublicInfoByRoleID(roleID int) (*PublicInfo, error) {
	var playerData model.PlayerData
	result := db.DB.Where("role_id = ?", roleID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析玩家公开信息数据
	var publicInfo PublicInfo
	if playerData.PublicInfo == "" || playerData.PublicInfo == "null" {
		// 如果没有公开信息数据，使用默认数据并保存到数据库
		log.Printf("玩家角色ID %d 的公开信息数据为空，使用默认数据", roleID)
		publicInfo = pm.GetDefaultPublicInfo()

		// 保存默认数据到数据库
		err := pm.SavePublicInfoByRoleID(roleID, &publicInfo)
		if err != nil {
			log.Printf("保存默认公开信息数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}

		return &publicInfo, nil
	}

	// 解析键值对格式
	publicInfo, err := pm.ParseKeyValuePublicInfo(playerData.PublicInfo)
	if err != nil {
		log.Printf("解析公开信息数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认公开信息数据")
		defaultPublicInfo := pm.GetDefaultPublicInfo()
		return &defaultPublicInfo, nil
	}

	return &publicInfo, nil
}

// SavePublicInfo 保存玩家公开信息到数据库
func (pm *PublicInfoManager) SavePublicInfo(deviceID string, publicInfo *PublicInfo) error {
	// 将玩家公开信息转换为键值对格式
	publicInfoStr, err := pm.ConvertPublicInfoToKeyValue(publicInfo)
	if err != nil {
		log.Printf("转换公开信息数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("public_info", publicInfoStr)
	if result.Error != nil {
		log.Printf("更新公开信息数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// SavePublicInfoByRoleID 通过角色ID保存玩家公开信息到数据库
func (pm *PublicInfoManager) SavePublicInfoByRoleID(roleID int, publicInfo *PublicInfo) error {
	// 将玩家公开信息转换为键值对格式
	publicInfoStr, err := pm.ConvertPublicInfoToKeyValue(publicInfo)
	if err != nil {
		log.Printf("转换公开信息数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("role_id = ?", roleID).Update("public_info", publicInfoStr)
	if result.Error != nil {
		log.Printf("更新公开信息数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// GetDefaultPublicInfo 获取默认的玩家公开信息
func (pm *PublicInfoManager) GetDefaultPublicInfo() PublicInfo {
	return PublicInfo{
		Name:              "Notitle",
		Gm:                true,
		Area:              "中国 四川",
		GradeThief:        "1",
		GradePointThief:   "0",
		GradePolice:       "1",
		GradePointPolice:  "0",
		Age:               "16",
		Sex:               "0",
		Icon:              0,
		UnionID:           0,
		Province:          "2",
		UnionName:         "",
		UnionBadge:        UnionBadge{IconID: 0, FlageID: 0, FrameID: 0},
		MembershipInfo:    []MembershipInfo{{MembershipID: 1, ExpireTime: 0}, {MembershipID: 2, ExpireTime: 0}},
		ActiveHeadBoxID:   900001,
		ActiveBubbleBoxID: 910001,
	}
}

// ParseKeyValuePublicInfo 解析键值对格式的公开信息
func (pm *PublicInfoManager) ParseKeyValuePublicInfo(data string) (PublicInfo, error) {
	publicInfo := pm.GetDefaultPublicInfo() // 使用默认值初始化
	
	// 直接解析键值对格式，不再尝试JSON格式解析
	log.Printf("解析键值对格式的公开信息数据")
	
	// 处理键值对格式
	lines := strings.Split(data, "\n")
	
	// 处理键值对
	key := ""
	value := ""
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("无效的键值对格式: %s", line)
			continue
		}
		
		key = strings.TrimSpace(parts[0])
		value = strings.TrimSpace(parts[1])
		
		// 去除可能存在的引号
		if len(value) > 1 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		
		switch key {
		case "name":
			publicInfo.Name = value
		case "icon":
			// 尝试将icon转换为整数
			if intValue, err := strconv.Atoi(value); err == nil {
				publicInfo.Icon = intValue
			} else {
				// 如果转换失败，则保持为字符串
				publicInfo.Icon = value
			}
		case "sex":
			publicInfo.Sex = value
		case "age":
			publicInfo.Age = value
		case "area":
			publicInfo.Area = value
		case "province":
			publicInfo.Province = value
		case "gradeThief":
			publicInfo.GradeThief = value
		case "gradePointThief":
			publicInfo.GradePointThief = value
		case "gradePolice":
			publicInfo.GradePolice = value
		case "gradePointPolice":
			publicInfo.GradePointPolice = value
		case "gm":
			publicInfo.Gm, _ = strconv.ParseBool(value)
		case "activeHeadBoxID":
			publicInfo.ActiveHeadBoxID, _ = strconv.Atoi(value)
		case "activeBubbleBoxID":
			publicInfo.ActiveBubbleBoxID, _ = strconv.Atoi(value)
		case "unionID":
			publicInfo.UnionID, _ = strconv.Atoi(value)
		case "unionName":
			publicInfo.UnionName = value
		default:
			if strings.HasPrefix(key, "unionBadge.") {
				field := strings.TrimPrefix(key, "unionBadge.")
				switch field {
				case "iconID":
					publicInfo.UnionBadge.IconID, _ = strconv.Atoi(value)
				case "frameID":
					publicInfo.UnionBadge.FrameID, _ = strconv.Atoi(value)
				case "flageID":
					publicInfo.UnionBadge.FlageID, _ = strconv.Atoi(value)
				}
			} else if strings.HasPrefix(key, "membershipInfo.") {
				membershipIDStr := strings.TrimPrefix(key, "membershipInfo.")
				membershipID, err := strconv.Atoi(membershipIDStr)
				if err != nil {
					log.Printf("无效的membershipID: %s", membershipIDStr)
					continue
				}
				expireTime, err := strconv.Atoi(value)
				if err != nil {
					log.Printf("无效的expireTime: %s", value)
					continue
				}

				// 查找是否已存在该membershipID
				found := false
				for i, mi := range publicInfo.MembershipInfo {
					if mi.MembershipID == membershipID {
						publicInfo.MembershipInfo[i].ExpireTime = expireTime
						found = true
						break
					}
				}
				if !found {
					publicInfo.MembershipInfo = append(publicInfo.MembershipInfo, MembershipInfo{MembershipID: membershipID, ExpireTime: expireTime})
				}
			}
		}
	}

	return publicInfo, nil
}

// ParsePublicInfoFromJSON 从键值对字符串解析公开信息数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (pm *PublicInfoManager) ParsePublicInfoFromJSON(publicInfoJSON string) (*PublicInfo, error) {
	// 如果数据为空，使用默认数据
	if publicInfoJSON == "" || publicInfoJSON == "null" {
		log.Printf("公开信息数据为空，使用默认数据")
		defaultPublicInfo := pm.GetDefaultPublicInfo()
		return &defaultPublicInfo, nil
	}

	// 解析键值对格式
	publicInfo, err := pm.ParseKeyValuePublicInfo(publicInfoJSON)
	if err != nil {
		log.Printf("解析公开信息数据失败: %v，使用默认值", err)
		// 解析失败时，使用默认数据
		defaultPublicInfo := pm.GetDefaultPublicInfo()
		return &defaultPublicInfo, nil
	}

	// 返回客户端需要的格式
	return &publicInfo, nil
}

// ConvertPublicInfoToKeyValue 将PublicInfo结构体转换为键值对格式的字符串
func (pm *PublicInfoManager) ConvertPublicInfoToKeyValue(publicInfo *PublicInfo) (string, error) {
	if publicInfo == nil {
		return "", game_error.New(-2, "公开信息数据为空")
	}

	// 直接使用键值对格式，不再尝试JSON序列化
	log.Printf("将PublicInfo转换为键值对格式")

	// 构建键值对格式的字符串
	var sb strings.Builder

	// 添加基本字段
	sb.WriteString(fmt.Sprintf("name=\"%s\"\n\n", publicInfo.Name))
	
	// 处理gm字段（布尔值）
	sb.WriteString(fmt.Sprintf("gm=%v\n\n", publicInfo.Gm))
	
	// 处理字符串字段
	sb.WriteString(fmt.Sprintf("area=\"%s\"\n\n", publicInfo.Area))
	
	// 处理等级和积分字段
	if intValue, err := strconv.Atoi(publicInfo.GradeThief); err == nil {
		sb.WriteString(fmt.Sprintf("gradeThief=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("gradeThief=\"%s\"\n\n", publicInfo.GradeThief))
	}
	
	if intValue, err := strconv.Atoi(publicInfo.GradePointThief); err == nil {
		sb.WriteString(fmt.Sprintf("gradePointThief=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("gradePointThief=\"%s\"\n\n", publicInfo.GradePointThief))
	}
	
	if intValue, err := strconv.Atoi(publicInfo.GradePolice); err == nil {
		sb.WriteString(fmt.Sprintf("gradePolice=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("gradePolice=\"%s\"\n\n", publicInfo.GradePolice))
	}
	
	if intValue, err := strconv.Atoi(publicInfo.GradePointPolice); err == nil {
		sb.WriteString(fmt.Sprintf("gradePointPolice=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("gradePointPolice=\"%s\"\n\n", publicInfo.GradePointPolice))
	}
	
	// 处理年龄和性别字段
	if intValue, err := strconv.Atoi(publicInfo.Age); err == nil {
		sb.WriteString(fmt.Sprintf("age=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("age=\"%s\"\n\n", publicInfo.Age))
	}
	
	if intValue, err := strconv.Atoi(publicInfo.Sex); err == nil {
		sb.WriteString(fmt.Sprintf("sex=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("sex=\"%s\"\n\n", publicInfo.Sex))
	}
	
	// 处理icon字段（可能是整数或字符串）
	switch v := publicInfo.Icon.(type) {
	case int:
		sb.WriteString(fmt.Sprintf("icon=%d\n\n", v))
	case float64:
		sb.WriteString(fmt.Sprintf("icon=%d\n\n", int(v)))
	case string:
		sb.WriteString(fmt.Sprintf("icon=\"%s\"\n\n", v))
	default:
		sb.WriteString(fmt.Sprintf("icon=\"%v\"\n\n", publicInfo.Icon))
	}
	
	// 处理整数字段
	sb.WriteString(fmt.Sprintf("unionID=%d\n\n", publicInfo.UnionID))
	
	// 处理省份字段
	if intValue, err := strconv.Atoi(publicInfo.Province); err == nil {
		sb.WriteString(fmt.Sprintf("province=%d\n\n", intValue))
	} else {
		sb.WriteString(fmt.Sprintf("province=\"%s\"\n\n", publicInfo.Province))
	}
	
	// 处理公会名称
	sb.WriteString(fmt.Sprintf("unionName=\"%s\"\n\n", publicInfo.UnionName))
	
	// 处理UnionBadge结构体
	sb.WriteString(fmt.Sprintf("unionBadge.iconID=%d\n\n", publicInfo.UnionBadge.IconID))
	sb.WriteString(fmt.Sprintf("unionBadge.flageID=%d\n\n", publicInfo.UnionBadge.FlageID))
	sb.WriteString(fmt.Sprintf("unionBadge.frameID=%d\n\n", publicInfo.UnionBadge.FrameID))
	
	// 处理MembershipInfo数组
	for _, info := range publicInfo.MembershipInfo {
		sb.WriteString(fmt.Sprintf("membershipInfo.%d=%d\n\n", info.MembershipID, info.ExpireTime))
	}
	
	// 处理其他整数字段
	sb.WriteString(fmt.Sprintf("activeHeadBoxID=%d\n\n", publicInfo.ActiveHeadBoxID))
	sb.WriteString(fmt.Sprintf("activeBubbleBoxID=%d", publicInfo.ActiveBubbleBoxID))
	
	return sb.String(), nil
}

// utils/radar_manager.go
package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"encoding/json"

	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
)

// RadarInfo 表示玩家雷达信息的结构
type RadarInfo struct {
	RadarThief            []int `json:"radarThief"`
	RadarPolice           []int `json:"radarPolice"`
	RadarRemainRoundPolice int   `json:"radarRemainRoundPolice"`
	RadarRemainRoundThief  int   `json:"radarRemainRoundThief"`
}

// RadarManager 提供玩家雷达信息的管理功能
type RadarManager struct{}

// NewRadarManager 创建一个新的玩家雷达信息管理器
func NewRadarManager() *RadarManager {
	return &RadarManager{}
}

// GetRadarInfo 从数据库获取指定设备ID的玩家雷达信息
func (rm *RadarManager) GetRadarInfo(deviceID string) (*RadarInfo, error) {
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析玩家雷达信息数据
	var radarInfo RadarInfo
	if playerData.PlayerRadar == "" || playerData.PlayerRadar == "null" {
		// 如果没有雷达信息数据，使用默认数据并保存到数据库
		log.Printf("玩家 %s 的雷达信息数据为空，使用默认数据", deviceID)
		radarInfo = rm.GetDefaultRadarInfo()

		// 保存默认数据到数据库
		err := rm.SaveRadarInfo(deviceID, &radarInfo)
		if err != nil {
			log.Printf("保存默认雷达信息数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}

		return &radarInfo, nil
	}

	// 解析键值对格式
	radarInfo, err := rm.ParseKeyValueRadarInfo(playerData.PlayerRadar)
	if err != nil {
		log.Printf("解析雷达信息数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认雷达信息数据")
		defaultRadarInfo := rm.GetDefaultRadarInfo()
		return &defaultRadarInfo, nil
	}

	return &radarInfo, nil
}

// GetRadarInfoByRoleID 通过角色ID获取玩家雷达信息
func (rm *RadarManager) GetRadarInfoByRoleID(roleID int) (*RadarInfo, error) {
	var playerData model.PlayerData
	result := db.DB.Where("role_id = ?", roleID).First(&playerData)
	if result.Error != nil {
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 解析玩家雷达信息数据
	var radarInfo RadarInfo
	if playerData.PlayerRadar == "" || playerData.PlayerRadar == "null" {
		// 如果没有雷达信息数据，使用默认数据并保存到数据库
		log.Printf("玩家角色ID %d 的雷达信息数据为空，使用默认数据", roleID)
		radarInfo = rm.GetDefaultRadarInfo()

		// 保存默认数据到数据库
		err := rm.SaveRadarInfoByRoleID(roleID, &radarInfo)
		if err != nil {
			log.Printf("保存默认雷达信息数据失败: %v", err)
			// 即使保存失败，仍然返回默认数据
		}

		return &radarInfo, nil
	}

	// 解析键值对格式
	radarInfo, err := rm.ParseKeyValueRadarInfo(playerData.PlayerRadar)
	if err != nil {
		log.Printf("解析雷达信息数据失败: %v", err)
		// 解析失败时，使用默认数据但不保存到数据库
		log.Printf("使用默认雷达信息数据")
		defaultRadarInfo := rm.GetDefaultRadarInfo()
		return &defaultRadarInfo, nil
	}

	return &radarInfo, nil
}

// SaveRadarInfo 保存玩家雷达信息到数据库
func (rm *RadarManager) SaveRadarInfo(deviceID string, radarInfo *RadarInfo) error {
	// 将玩家雷达信息转换为键值对格式
	radarInfoStr, err := rm.ConvertRadarInfoToKeyValue(radarInfo)
	if err != nil {
		log.Printf("转换雷达信息数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("device_id = ?", deviceID).Update("player_radar", radarInfoStr)
	if result.Error != nil {
		log.Printf("更新雷达信息数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// SaveRadarInfoByRoleID 通过角色ID保存玩家雷达信息到数据库
func (rm *RadarManager) SaveRadarInfoByRoleID(roleID int, radarInfo *RadarInfo) error {
	// 将玩家雷达信息转换为键值对格式
	radarInfoStr, err := rm.ConvertRadarInfoToKeyValue(radarInfo)
	if err != nil {
		log.Printf("转换雷达信息数据失败: %v", err)
		return game_error.New(-2, "数据处理错误")
	}

	// 更新数据库
	result := db.DB.Model(&model.PlayerData{}).Where("role_id = ?", roleID).Update("player_radar", radarInfoStr)
	if result.Error != nil {
		log.Printf("更新雷达信息数据失败: %v", result.Error)
		return game_error.New(-2, "数据库更新错误")
	}

	return nil
}

// GetDefaultRadarInfo 获取默认的玩家雷达信息
func (rm *RadarManager) GetDefaultRadarInfo() RadarInfo {
	return RadarInfo{
		RadarThief:            []int{64, 64, 36, 4, 4},
		RadarPolice:           []int{},
		RadarRemainRoundPolice: 1,
		RadarRemainRoundThief:  0,
	}
}

// ParseKeyValueRadarInfo 解析键值对格式的雷达信息
func (rm *RadarManager) ParseKeyValueRadarInfo(data string) (RadarInfo, error) {
	radarInfo := rm.GetDefaultRadarInfo() // 使用默认值初始化
	
	// 处理键值对格式
	log.Printf("解析键值对格式的雷达信息数据")
	
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
		
		switch key {
		case "radarThief":
			// 解析雷达数组格式 [64,64,36,4,4]
			if len(value) >= 2 && value[0] == '[' && value[len(value)-1] == ']' {
				// 去除方括号
				value = value[1 : len(value)-1]
				// 分割数组元素
				elements := strings.Split(value, ",")
				
				// 转换为整数数组
				radarThief := make([]int, 0, len(elements))
				for _, elem := range elements {
					elem = strings.TrimSpace(elem)
					if elem == "" {
						continue
					}
					
					val, err := strconv.Atoi(elem)
					if err != nil {
						log.Printf("无效的雷达值: %s", elem)
						continue
					}
					
					radarThief = append(radarThief, val)
				}
				
				radarInfo.RadarThief = radarThief
			} else {
				log.Printf("无效的雷达数组格式: %s", value)
			}
			
		case "radarPolice":
			// 解析雷达数组格式 []
			if len(value) >= 2 && value[0] == '[' && value[len(value)-1] == ']' {
				// 去除方括号
				value = value[1 : len(value)-1]
				
				// 如果数组为空，则设置为空数组
				if value == "" {
					radarInfo.RadarPolice = []int{}
					continue
				}
				
				// 分割数组元素
				elements := strings.Split(value, ",")
				
				// 转换为整数数组
				radarPolice := make([]int, 0, len(elements))
				for _, elem := range elements {
					elem = strings.TrimSpace(elem)
					if elem == "" {
						continue
					}
					
					val, err := strconv.Atoi(elem)
					if err != nil {
						log.Printf("无效的雷达值: %s", elem)
						continue
					}
					
					radarPolice = append(radarPolice, val)
				}
				
				radarInfo.RadarPolice = radarPolice
			} else {
				log.Printf("无效的雷达数组格式: %s", value)
			}
			
		case "radarRemainRoundPolice":
			// 解析剩余回合数
			val, err := strconv.Atoi(value)
			if err != nil {
				log.Printf("无效的剩余回合数: %s", value)
				return radarInfo, game_error.New(-2, "雷达剩余回合数据解析失败")
			}
			radarInfo.RadarRemainRoundPolice = val
			
		case "radarRemainRoundThief":
			// 解析剩余回合数
			val, err := strconv.Atoi(value)
			if err != nil {
				log.Printf("无效的剩余回合数: %s", value)
				return radarInfo, game_error.New(-2, "雷达剩余回合数据解析失败")
			}
			radarInfo.RadarRemainRoundThief = val
		}
	}

	return radarInfo, nil
}

// ParseRadarInfoFromJSON 从JSON字符串解析雷达信息数据并返回客户端需要的格式
// 此方法用于当isSelf为true时，直接使用playerData中的数据而不再查询数据库
func (rm *RadarManager) ParseRadarInfoFromJSON(radarInfoJSON string) ([]int, []int, int, int, error) {
	// 如果数据为空，使用默认数据
	if radarInfoJSON == "" || radarInfoJSON == "null" {
		log.Printf("雷达信息数据为空，使用默认数据")
		defaultRadarInfo := rm.GetDefaultRadarInfo()
		return defaultRadarInfo.RadarThief, defaultRadarInfo.RadarPolice, defaultRadarInfo.RadarRemainRoundPolice, defaultRadarInfo.RadarRemainRoundThief, nil
	}

	// 解析键值对格式
	radarInfo, err := rm.ParseKeyValueRadarInfo(radarInfoJSON)
	if err != nil {
		log.Printf("解析雷达信息数据失败: %v，使用默认值", err)
		// 解析失败时，使用默认数据
		defaultRadarInfo := rm.GetDefaultRadarInfo()
		return defaultRadarInfo.RadarThief, defaultRadarInfo.RadarPolice, defaultRadarInfo.RadarRemainRoundPolice, defaultRadarInfo.RadarRemainRoundThief, nil
	}

	// 返回客户端需要的格式
	return radarInfo.RadarThief, radarInfo.RadarPolice, radarInfo.RadarRemainRoundPolice, radarInfo.RadarRemainRoundThief, nil
}

// ConvertRadarInfoToKeyValue 将RadarInfo结构体转换为键值对格式的字符串
func (rm *RadarManager) ConvertRadarInfoToKeyValue(radarInfo *RadarInfo) (string, error) {
	if radarInfo == nil {
		return "", game_error.New(-2, "雷达信息数据为空")
	}

	// 构建键值对格式的字符串
	var sb strings.Builder

	// 处理雷达数组
	radarThiefJSON, err := json.Marshal(radarInfo.RadarThief)
	if err != nil {
		log.Printf("序列化雷达数据失败: %v", err)
		return "", game_error.New(-2, "数据处理错误")
	}
	sb.WriteString(fmt.Sprintf("radarThief=%s\n\n", string(radarThiefJSON)))

	radarPoliceJSON, err := json.Marshal(radarInfo.RadarPolice)
	if err != nil {
		log.Printf("序列化雷达数据失败: %v", err)
		return "", game_error.New(-2, "数据处理错误")
	}
	sb.WriteString(fmt.Sprintf("radarPolice=%s\n\n", string(radarPoliceJSON)))

	// 处理剩余回合数
	sb.WriteString(fmt.Sprintf("radarRemainRoundPolice=%d\n\n", radarInfo.RadarRemainRoundPolice))
	sb.WriteString(fmt.Sprintf("radarRemainRoundThief=%d", radarInfo.RadarRemainRoundThief))

	return sb.String(), nil
}
// internal/model/playerinfo.go
package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

)

// PlayerInfo 表示dmm_playerinfo表的结构
type PlayerInfo struct {
	DeviceID     string `gorm:"primaryKey"` // 设置为主键
	RoleID       int    `gorm:"uniqueIndex"` // 设置为唯一索引（从键）
	RealDeviceID string `gorm:"type:text"` // 存储历史realDeviceID，使用双换行符分隔
	DeviceInfo   string `gorm:"type:text"` // 存储历史deviceInfo，使用双换行符分隔
	IP           string `gorm:"type:text"` // 存储历史IP地址，使用双换行符分隔
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// ValidateRoleID 验证roleID是否有效，确保与dmm_playerdata表中的roleID匹配
// 注意：此函数不应直接在model层进行数据库操作，仅提供验证逻辑
// 实际的数据库查询和验证应在服务层或控制器中实现
func (p *PlayerInfo) ValidateRoleID(roleID int) bool {
	// 检查roleID是否为正整数
	if roleID <= 0 {
		return false
	}
	
	// 检查roleID是否与当前实例的roleID匹配
	// 如果当前实例的roleID已设置（非零值），则必须匹配
	if p.RoleID > 0 && p.RoleID != roleID {
		return false
	}
	
	return true
}

// AddRealDeviceID 添加新的realDeviceID到历史记录中，使用双换行符分隔
func (p *PlayerInfo) AddRealDeviceID(newRealDeviceID string) error {
	// 处理空值或无效值的情况
	if p.RealDeviceID == "" || p.RealDeviceID == "Invalid value." || p.RealDeviceID == "null" {
		p.RealDeviceID = newRealDeviceID
		return nil
	}
	
	// 尝试解析旧的JSON格式数据
	var realDeviceIDs []string
	if err := json.Unmarshal([]byte(p.RealDeviceID), &realDeviceIDs); err == nil {
		// 如果成功解析为JSON，则将其转换为新格式
		// 检查是否已存在相同的realDeviceID
		for _, id := range realDeviceIDs {
			if id == newRealDeviceID {
				// 已存在，无需添加，但需要转换格式
				p.RealDeviceID = convertArrayToText(realDeviceIDs)
				return nil
			}
		}
		
		// 添加新的realDeviceID并转换格式
		realDeviceIDs = append(realDeviceIDs, newRealDeviceID)
		p.RealDeviceID = convertArrayToText(realDeviceIDs)
		return nil
	}
	
	// 如果不是JSON格式，则按照新格式处理
	// 将现有内容按双换行符分割
	realDeviceIDList := splitTextByDoubleNewline(p.RealDeviceID)
	
	// 检查是否已存在相同的realDeviceID
	for _, id := range realDeviceIDList {
		if id == newRealDeviceID {
			// 已存在，无需添加
			return nil
		}
	}
	
	// 添加新的realDeviceID
	if p.RealDeviceID != "" {
		p.RealDeviceID = p.RealDeviceID + "\n\n" + newRealDeviceID
	} else {
		p.RealDeviceID = newRealDeviceID
	}
	
	return nil
}

// 辅助函数：将字符串数组转换为双换行符分隔的文本
func convertArrayToText(items []string) string {
	if len(items) == 0 {
		return ""
	}
	
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += "\n\n" + items[i]
	}
	
	return result
}

// 辅助函数：将双换行符分隔的文本分割为字符串数组
func splitTextByDoubleNewline(text string) []string {
	if text == "" {
		return []string{}
	}
	
	// 使用双换行符分割文本
	var result []string
	parts := strings.Split(text, "\n\n")
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	
	return result
}

// AddDeviceInfo 添加新的deviceInfo到历史记录中，使用双换行符分隔
// 在添加前会先调用GetStableDeviceInfo将deviceInfo转换为稳定格式，移除会变化的内存值部分
func (p *PlayerInfo) AddDeviceInfo(newDeviceInfo string) error {
	// 将deviceInfo转换为稳定格式，移除会变化的内存值部分
	newDeviceInfo = GetStableDeviceInfo(newDeviceInfo)
	// 处理空值或无效值的情况
	if p.DeviceInfo == "" || p.DeviceInfo == "Invalid value." || p.DeviceInfo == "null" {
		p.DeviceInfo = newDeviceInfo
		return nil
	}
	
	// 尝试解析旧的JSON格式数据
	var deviceInfos []string
	if err := json.Unmarshal([]byte(p.DeviceInfo), &deviceInfos); err == nil {
		// 如果成功解析为JSON，则将其转换为新格式
		// 检查是否已存在相同的deviceInfo
		for _, info := range deviceInfos {
			if info == newDeviceInfo {
				// 已存在，无需添加，但需要转换格式
				p.DeviceInfo = convertArrayToText(deviceInfos)
				return nil
			}
		}
		
		// 添加新的deviceInfo并转换格式
		deviceInfos = append(deviceInfos, newDeviceInfo)
		p.DeviceInfo = convertArrayToText(deviceInfos)
		return nil
	}
	
	// 如果不是JSON格式，则按照新格式处理
	// 将现有内容按双换行符分割
	deviceInfoList := splitTextByDoubleNewline(p.DeviceInfo)
	
	// 检查是否已存在相同的deviceInfo
	for _, info := range deviceInfoList {
		if info == newDeviceInfo {
			// 已存在，无需添加
			return nil
		}
	}
	
	// 添加新的deviceInfo
	if p.DeviceInfo != "" {
		p.DeviceInfo = p.DeviceInfo + "\n\n" + newDeviceInfo
	} else {
		p.DeviceInfo = newDeviceInfo
	}
	
	return nil
}

// GetRealDeviceIDCount 获取realDeviceID历史记录的数量
func (p *PlayerInfo) GetRealDeviceIDCount() (int, error) {
	if p.RealDeviceID == "" || p.RealDeviceID == "Invalid value." || p.RealDeviceID == "null" {
		return 0, nil
	}
	
	// 尝试解析旧的JSON格式数据
	var realDeviceIDs []string
	if err := json.Unmarshal([]byte(p.RealDeviceID), &realDeviceIDs); err == nil {
		// 如果成功解析为JSON，返回数组长度
		return len(realDeviceIDs), nil
	}
	
	// 如果不是JSON格式，则按照新格式处理
	// 将现有内容按双换行符分割并计算数量
	realDeviceIDList := splitTextByDoubleNewline(p.RealDeviceID)
	return len(realDeviceIDList), nil
}

// GetDeviceInfoCount 获取deviceInfo历史记录的数量
func (p *PlayerInfo) GetDeviceInfoCount() (int, error) {
	if p.DeviceInfo == "" || p.DeviceInfo == "Invalid value." || p.DeviceInfo == "null" {
		return 0, nil
	}
	
	// 尝试解析旧的JSON格式数据
	var deviceInfos []string
	if err := json.Unmarshal([]byte(p.DeviceInfo), &deviceInfos); err == nil {
		// 如果成功解析为JSON，返回数组长度
		return len(deviceInfos), nil
	}
	
	// 如果不是JSON格式，则按照新格式处理
	// 将现有内容按双换行符分割并计算数量
	deviceInfoList := splitTextByDoubleNewline(p.DeviceInfo)
	return len(deviceInfoList), nil
}

// AddIP 添加新的IP到历史记录中，使用双换行符分隔
func (p *PlayerInfo) AddIP(newIP string) error {
	// 处理空值或无效值的情况
	if p.IP == "" || p.IP == "Invalid value." || p.IP == "null" {
		p.IP = newIP
		return nil
	}
	
	// 尝试解析旧的JSON格式数据
	var ips []string
	if err := json.Unmarshal([]byte(p.IP), &ips); err == nil {
		// 如果成功解析为JSON，则将其转换为新格式
		// 检查是否已存在相同的IP
		for _, ip := range ips {
			if ip == newIP {
				// 已存在，无需添加，但需要转换格式
				p.IP = convertArrayToText(ips)
				return nil
			}
		}
		
		// 添加新的IP并转换格式
		ips = append(ips, newIP)
		p.IP = convertArrayToText(ips)
		return nil
	}
	
	// 如果不是JSON格式，则按照新格式处理
	// 将现有内容按双换行符分割
	ipList := splitTextByDoubleNewline(p.IP)
	
	// 检查是否已存在相同的IP
	for _, ip := range ipList {
		if ip == newIP {
			// 已存在，无需添加
			return nil
		}
	}
	
	// 添加新的IP
	if p.IP != "" {
		p.IP = p.IP + "\n\n" + newIP
	} else {
		p.IP = newIP
	}
	
	return nil
}

// GetIPCount 获取IP历史记录的数量
func (p *PlayerInfo) GetIPCount() (int, error) {
	if p.IP == "" || p.IP == "Invalid value." || p.IP == "null" {
		return 0, nil
	}
	
	// 尝试解析旧的JSON格式数据
	var ips []string
	if err := json.Unmarshal([]byte(p.IP), &ips); err == nil {
		// 如果成功解析为JSON，返回数组长度
		return len(ips), nil
	}
	
	// 如果不是JSON格式，则按照新格式处理
	// 将现有内容按双换行符分割并计算数量
	ipList := splitTextByDoubleNewline(p.IP)
	return len(ipList), nil
}

// TableName 指定表名
func (PlayerInfo) TableName() string {
	return "dmm_playerinfo"
}

// ValidateAndSetRoleID 验证并设置roleID，确保与dmm_playerdata表中的roleID匹配
// 此函数提供验证逻辑，但不直接进行数据库操作
// 参数说明：
// - deviceID: 设备ID
// - roleIDFromPlayerData: 从dmm_playerdata表中查询到的roleID
// - checkDuplicate: 是否检查roleID在dmm_playerinfo表中是否已存在（用于新记录创建时）
// 返回值：
// - error: 如果验证失败，返回错误信息
func (p *PlayerInfo) ValidateAndSetRoleID(deviceID string, roleIDFromPlayerData int, checkDuplicate bool) error {
	// 检查参数有效性
	if deviceID == "" {
		return fmt.Errorf("设备ID不能为空")
	}
	
	// 检查roleID是否有效
	if roleIDFromPlayerData <= 0 {
		return fmt.Errorf("无效的角色ID: %d", roleIDFromPlayerData)
	}
	
	// 如果当前实例已有roleID，确保与传入的roleID匹配
	if p.RoleID > 0 && p.RoleID != roleIDFromPlayerData {
		return fmt.Errorf("角色ID不匹配: 当前=%d, 传入=%d", p.RoleID, roleIDFromPlayerData)
	}
	
	// 注意：此处不进行数据库操作，仅提供验证逻辑
	// 实际的数据库查询应在服务层或控制器中实现
	// checkDuplicate参数用于提示调用者是否需要在数据库中检查roleID唯一性
	
	// 设置roleID
	p.RoleID = roleIDFromPlayerData
	
	return nil
}
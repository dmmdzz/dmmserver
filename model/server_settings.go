// internal/model/server_settings.go
package model

import (
	"time"
)

// ServerSettings 表示dmm_settings表的结构
// 注意：ServerID、ServerVersion和ServerTimeZoneOffset已从结构体中移除，将在代码中实时生成
type ServerSettings struct {
	GraphicsOptions string // JSON格式存储
	MiscOptions     string // JSON格式存储
	ServerIP        *string
	ServerPort      *string
	ServerOverDayTimeStamp int64
	PlaytimeSettings string // JSON格式存储游戏时长设置
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (ServerSettings) TableName() string {
	return "dmm_settings"
}

// DefaultGraphicsOptions 返回默认的图形选项JSON字符串
func DefaultGraphicsOptions() string {
	return `[
		{"level": 1, "isDefault": 0, "shadow": 0, "maxParticles": 1000, "renderScale": 0.8},
		{"level": 3, "isDefault": 1, "shadow": 1, "maxParticles": 5000, "renderScale": 1}
	]`
}

// DefaultMiscOptions 返回默认的杂项选项JSON字符串
func DefaultMiscOptions() string {
	return `{"outline": 1, "HFR": 1, "BRHFR": 1, "HFX": 1}`
}

// DefaultPlaytimeSettings 返回默认的游戏时长设置JSON字符串
func DefaultPlaytimeSettings() string {
	return `{
		"freeUserDailyLimit": 5400,
		"luckyUserExtraTime": 2700,
		"paidUserDailyLimit": 36000
	}`
}
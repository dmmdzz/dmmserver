// internal/model/ban.go
package model

import "time"

type BanDeviceID struct {
	DeviceID    string `gorm:"primaryKey"`
	BanningTime time.Time
}

func (BanDeviceID) TableName() string {
	return "Ban_DeviceID"
}

type BanIP struct {
	IP          string `gorm:"primaryKey"`
	BanningTime time.Time
}

func (BanIP) TableName() string {
	return "Ban_IP"
}

type BanRealDeviceID struct {
	RealDeviceID string `gorm:"primaryKey"`
	BanningTime  time.Time
}

func (BanRealDeviceID) TableName() string {
	return "Ban_realDeviceID"
}

type BanDeviceInfo struct {
	DeviceInfo  string `gorm:"primaryKey"`
	BanningTime time.Time
}

func (BanDeviceInfo) TableName() string {
	return "Ban_DeviceInfo"
}
// utils/radar_manager_example.go
package utils

import (
	"fmt"
	"log"
)

// 示例：如何使用RadarManager获取和保存玩家雷达信息
func ExampleRadarManager() {
	// 创建雷达管理器
	rm := NewRadarManager()

	// 示例1：获取玩家雷达信息
	deviceID := "example_device_id"
	radarInfo, err := rm.GetRadarInfo(deviceID)
	if err != nil {
		log.Printf("获取玩家雷达信息失败: %v", err)
		return
	}

	// 打印雷达信息
	fmt.Printf("玩家雷达信息:\n")
	fmt.Printf("  盗贼雷达: %v\n", radarInfo.RadarThief)
	fmt.Printf("  警察雷达: %v\n", radarInfo.RadarPolice)
	fmt.Printf("  警察雷达剩余回合: %d\n", radarInfo.RadarRemainRoundPolice)
	fmt.Printf("  盗贼雷达剩余回合: %d\n", radarInfo.RadarRemainRoundThief)

	// 示例2：修改玩家雷达信息
	radarInfo.RadarThief = []int{70, 60, 40, 10, 5}
	radarInfo.RadarRemainRoundThief = 3

	// 保存修改后的雷达信息
	err = rm.SaveRadarInfo(deviceID, radarInfo)
	if err != nil {
		log.Printf("保存玩家雷达信息失败: %v", err)
		return
	}

	// 示例3：通过角色ID获取玩家雷达信息
	roleID := 12345
	radarInfoByRole, err := rm.GetRadarInfoByRoleID(roleID)
	if err != nil {
		log.Printf("通过角色ID获取玩家雷达信息失败: %v", err)
		return
	}

	// 打印通过角色ID获取的雷达信息
	fmt.Printf("通过角色ID获取的玩家雷达信息:\n")
	fmt.Printf("  盗贼雷达: %v\n", radarInfoByRole.RadarThief)
	fmt.Printf("  警察雷达: %v\n", radarInfoByRole.RadarPolice)
	fmt.Printf("  警察雷达剩余回合: %d\n", radarInfoByRole.RadarRemainRoundPolice)
	fmt.Printf("  盗贼雷达剩余回合: %d\n", radarInfoByRole.RadarRemainRoundThief)
}
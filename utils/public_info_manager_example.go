// utils/public_info_manager_example.go
package utils

import (
	"fmt"
	"log"
)

// 这个文件提供了PublicInfoManager的使用示例

// PublicInfoManagerExample 展示如何使用PublicInfoManager
func PublicInfoManagerExample() {
	// 创建一个PublicInfoManager实例
	pm := NewPublicInfoManager()

	// 示例1: 获取玩家的公开信息
	deviceID := "example_device_id"
	publicInfo, err := pm.GetPublicInfo(deviceID)
	if err != nil {
		log.Printf("获取玩家公开信息失败: %v", err)
		return
	}

	// 打印玩家名称和其他信息
	fmt.Printf("玩家名称: %s\n", publicInfo.Name)
	fmt.Printf("盗贼等级: %s\n", publicInfo.GradeThief)
	fmt.Printf("警察等级: %s\n", publicInfo.GradePolice)

	// 示例2: 修改玩家公开信息并保存
	// 直接修改PublicInfo结构体的字段
	publicInfo.Name = "新玩家名称"
	publicInfo.GradeThief = "5"
	publicInfo.GradePointThief = "1000"
	publicInfo.GradePolice = "4"
	publicInfo.GradePointPolice = "800"

	// 保存修改后的公开信息
	err = pm.SavePublicInfo(deviceID, publicInfo)
	if err != nil {
		log.Printf("保存玩家公开信息失败: %v", err)
	} else {
		fmt.Println("玩家信息已更新")

		// 重新获取玩家信息以验证更新是否成功
		updatedInfo, err := pm.GetPublicInfo(deviceID)
		if err != nil {
			log.Printf("获取更新后的玩家信息失败: %v", err)
		} else {
			fmt.Printf("更新后的玩家名称: %s\n", updatedInfo.Name)
			fmt.Printf("更新后的盗贼等级: %s\n", updatedInfo.GradeThief)
		}
	}

	// 示例3: 通过角色ID获取和更新公开信息
	roleID := 123
	publicInfoByRole, err := pm.GetPublicInfoByRoleID(roleID)
	if err != nil {
		log.Printf("通过角色ID获取玩家公开信息失败: %v", err)
	} else {
		fmt.Printf("角色ID %d 的玩家名称: %s\n", roleID, publicInfoByRole.Name)

		// 修改信息
		publicInfoByRole.Name = "通过角色ID修改的名称"
		publicInfoByRole.Area = "中国 北京"
		publicInfoByRole.ActiveHeadBoxID = 900002
		
		// 保存修改后的信息
		err = pm.SavePublicInfoByRoleID(roleID, publicInfoByRole)
		if err != nil {
			log.Printf("通过角色ID保存玩家公开信息失败: %v", err)
		} else {
			fmt.Println("通过角色ID更新玩家信息成功")
		}
	}

	// 示例4: 展示如何处理会员信息
	deviceID = "vip_player_device_id"
	vipInfo, err := pm.GetPublicInfo(deviceID)
	if err != nil {
		log.Printf("获取VIP玩家信息失败: %v", err)
	} else {
		// 打印当前会员信息
		fmt.Println("当前会员信息:")
		for _, mi := range vipInfo.MembershipInfo {
			fmt.Printf("会员ID: %d, 过期时间: %d\n", mi.MembershipID, mi.ExpireTime)
		}

		// 更新会员信息
		vipInfo.MembershipInfo = []MembershipInfo{
			{MembershipID: 1, ExpireTime: 1735689600}, // 2025年1月1日
			{MembershipID: 2, ExpireTime: 1767225600}, // 2026年1月1日
		}

		// 保存更新后的会员信息
		err = pm.SavePublicInfo(deviceID, vipInfo)
		if err != nil {
			log.Printf("保存VIP玩家信息失败: %v", err)
		} else {
			fmt.Println("VIP玩家信息已更新")
		}
	}

	// 示例5: 展示如何处理公会徽章信息
	deviceID = "union_player_device_id"
	unionPlayerInfo, err := pm.GetPublicInfo(deviceID)
	if err != nil {
		log.Printf("获取公会玩家信息失败: %v", err)
	} else {
		// 更新公会信息
		unionPlayerInfo.UnionID = 12345
		unionPlayerInfo.UnionName = "超级战队"
		unionPlayerInfo.UnionBadge = UnionBadge{
			IconID:  101,
			FrameID: 202,
			FlageID: 303,
		}

		// 保存更新后的公会信息
		err = pm.SavePublicInfo(deviceID, unionPlayerInfo)
		if err != nil {
			log.Printf("保存公会玩家信息失败: %v", err)
		} else {
			fmt.Println("公会玩家信息已更新")
		}
	}
}

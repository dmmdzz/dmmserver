
package model

import (
	"regexp"
	"strings"
)

// GetStableDeviceInfo 从完整的deviceInfo中提取稳定部分
// 移除会变化的内存值部分（如"249868MB"）
// 例如：将"OnePlus PJF110_Android OS 14 / API-34 (UP1A.231005.007/U.19fe0de^3e4c^3e4d)_15315MB_ARM64 FP ASIMD AES_Adreno (TM) 732_OpenGLES3_4096MB_8_5606_249868MB_50"
// 转换为"OnePlus PJF110_Android OS 14 / API-34 (UP1A.231005.007/U.19fe0de^3e4c^3e4d)_15315MB_ARM64 FP ASIMD AES_Adreno (TM) 732_OpenGLES3_4096MB_8_5606_50"
func GetStableDeviceInfo(deviceInfo string) string {
	if deviceInfo == "" {
		return ""
	}

	// 使用正则表达式匹配格式：数字+MB+下划线+数字
	// 这将匹配类似"249868MB_50"的部分
	pattern := regexp.MustCompile(`(\d+MB)_\d+$`)
	parts := pattern.FindStringSubmatch(deviceInfo)

	if len(parts) >= 2 {
		// 找到了变化的内存值部分
		// 提取最后的数字（如"50"）
		lastPart := deviceInfo[strings.LastIndex(deviceInfo, "_")+1:]
		
		// 找到倒数第二个下划线的位置
		beforePart := deviceInfo[:strings.LastIndex(deviceInfo, "_")]
		beforePart = beforePart[:strings.LastIndex(beforePart, "_")+1]
		
		// 组合稳定部分
		return beforePart + lastPart
	}

	// 如果没有匹配到预期格式，返回原始字符串
	return deviceInfo
}
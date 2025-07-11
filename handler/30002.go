// internal/handler/30002.go
package handler

import (
	"encoding/json"
//	"fmt"
	"log"
	"time"

	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
	"dmmserver/services/serversettings"
	"dmmserver/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	Register("30002", handle30002)
}

// handle30002 处理获取玩家完整档案请求
func handle30002(c *gin.Context, msgData map[string]interface{}) (map[string]interface{}, error) {
	// 创建所需的管理器实例，避免重复创建
	pm := utils.NewPublicInfoManager()
	cardManager := utils.NewCardManager()
	cardSkinManager := utils.NewCardSkinManager()
	cardStyleManager := utils.NewCardStyleManager()
	radarManager := utils.NewRadarManager()
	emotionManager := utils.NewEmotionManager()
	boxesManager := utils.NewBoxesManager()
	log.Printf("Executing handler for msg_id=30002. Received msgData: %+v", msgData)

	// 1. 参数解析和验证
	// ----------------------
	// 验证必须的参数字段
	requiredFields := []string{"requestRoleID", "authKey", "accountName", "roleID", "pfID", "deviceID", "version", "baseVerCode", "compVerCode", "sv", "sequenceID"}
	for _, field := range requiredFields {
		if _, exists := msgData[field]; !exists {
			log.Printf("错误：msg_id=30002 缺少 '%s' 参数", field)
			return nil, game_error.New(-5, "参数丢失，请重新登录")
		}
	}

	deviceID, ok := msgData["deviceID"].(string)
	if !ok || deviceID == "" {
		log.Println("错误：msg_id=30002 缺少 'deviceID' 参数")
		return nil, game_error.New(-5, "缺少 'deviceID' 参数")
	}

	authKey, ok := msgData["authKey"].(string)
	if !ok || authKey == "" {
		log.Println("错误：msg_id=30002 缺少 'authKey' 参数")
		return nil, game_error.New(-5, "缺少 'authKey' 参数")
	}

	// 获取accountName和roleID
	accountName, ok := msgData["accountName"].(string)
	if !ok || accountName == "" {
		log.Println("错误：msg_id=30002 缺少 'accountName' 参数")
		return nil, game_error.New(-5, "缺少 'accountName' 参数")
	}

	roleIDFloat, ok := msgData["roleID"].(float64)
	if !ok {
		log.Println("错误：msg_id=30002 'roleID' 参数类型错误")
		return nil, game_error.New(-13, "非法参数")
	}
	roleID := int(roleIDFloat)

	// 获取requestRoleID参数
	requestRoleIDFloat, ok := msgData["requestRoleID"].(float64)
	if !ok {
		log.Println("错误：msg_id=30002 'requestRoleID' 参数类型错误")
		return nil, game_error.New(-13, "非法参数")
	}
	requestRoleID := int(requestRoleIDFloat)

	// 2. 业务逻辑
	// -------------------------------------------------------------
	// 根据 deviceID 查询 dmm_playerdata
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		log.Printf("未找到 deviceID 为 '%s' 的玩家", deviceID)
		return nil, game_error.New(-3, "未找到玩家数据")
	}

	// 3. 验证 authKey 是否过期且匹配
	currentTime := time.Now().Unix()
	if playerData.AuthKeyExpire < currentTime {
		log.Printf("authKey 已过期，过期时间: %d, 当前时间: %d", playerData.AuthKeyExpire, currentTime)
		return nil, game_error.New(-12, "登录秘钥失效，请重新登录")
	}

	if playerData.AuthKey != authKey {
		log.Printf("authKey 不匹配，请求的 authKey: %s, 数据库中的 authKey: %s", authKey, playerData.AuthKey)
		return nil, game_error.New(-11, "登录验证错误，账号或已在别处登录")
	}

	// 验证accountName和roleID是否与数据库中的匹配
	// 使用PublicInfoManager获取玩家名字
	publicInfoObj, err := pm.GetPublicInfo(deviceID)
	if err != nil {
		log.Printf("获取玩家公开信息失败: %v", err)
		return nil, game_error.New(-3, "获取玩家数据失败")
	}
	
	// 检查publicInfoObj.Name是否为空，如果为空则更新为请求中的accountName
	if publicInfoObj.Name == "" {
		log.Printf("数据库中的 accountName 为空，使用请求中的 accountName: %s 更新数据库", accountName)
		publicInfoObj.Name = accountName
		// 保存更新后的公开信息
		err = pm.SavePublicInfo(deviceID, publicInfoObj)
		if err != nil {
			log.Printf("更新玩家公开信息失败: %v", err)
			// 即使更新失败，仍然继续处理请求
		}
	} else if publicInfoObj.Name != accountName {
		log.Printf("accountName 不匹配，请求的 accountName: %s, 数据库中的 accountName: %s", accountName, publicInfoObj.Name)
		return nil, game_error.New(-13, "非法参数")
	}

	if playerData.RoleID != roleID {
		log.Printf("roleID 不匹配，请求的 roleID: %d, 数据库中的 roleID: %d", roleID, playerData.RoleID)
		return nil, game_error.New(-13, "非法参数")
	}
	log.Printf("Current Playerdata : %s\n", playerData)
	log.Printf("Current result : %s", result)
	// 4. 获取服务器设置
	// 在处理请求前检查设置是否过期
	serversettings.CheckAndRefreshIfStale()
	// 使用缓存的服务器设置
	// serverSettings := serversettings.GetSettings() // 移除未使用的变量

	// 5. 获取被请求查看的玩家数据
	// 否则查看其他玩家的档案
	var requestedPlayerData model.PlayerData
	isSelf := requestRoleID == roleID

	if isSelf {
		// 查看自己的档案，直接使用已验证的playerData
		requestedPlayerData = playerData
	}
	// 注意：这里不再使用原始的数据库查询，而是统一使用PublicInfoManager获取玩家信息

	// 6. 构建响应数据
	// 这里我们需要从数据库中读取玩家数据，并构建响应数据
	// 获取服务器跨天时间戳
	var serverOverDayTimeStamp int64 = time.Now().Unix() // 默认使用当前时间
	// 从服务器设置中获取serverOverDayTimeStamp
	serverSettings := serversettings.GetSettings()
	serverOverDayTimeStamp = serverSettings.ServerOverDayTimeStamp
	if serverOverDayTimeStamp == 0 {
		log.Printf("serverOverDayTimeStamp为0，使用当前时间")
		serverOverDayTimeStamp = 1660924815 // 2022-08-16 00:00:00	
	}

	// 根据roleID获取玩家公开信息
	if isSelf {
		// 如果是查询自己的信息，直接使用playerData中的数据
		publicInfoObj, err = pm.ParsePublicInfoFromJSON(requestedPlayerData.PublicInfo)
	} else {
		// 如果是查询他人信息，使用requestRoleID获取
		publicInfoObj, err = pm.GetPublicInfoByRoleID(requestRoleID)
	}

	if err != nil {
		log.Printf("获取玩家公开信息失败: %v，使用默认值", err)
		// 使用默认公开信息
		defaultInfo := pm.GetDefaultPublicInfo()
		publicInfoObj = &defaultInfo
	}

	// 添加实时生成的IP地址
	publicInfoObj.IP = utils.GetClientIP(c)

	// 将PublicInfo对象直接作为JSON对象返回给客户端
	// 不再转换为键值对格式，因为键值对格式只用于数据库存储
	publicInfoJSON, err := json.Marshal(publicInfoObj)
	if err != nil {
		log.Printf("转换公开信息数据为JSON格式失败: %v，返回错误", err)
		return nil, game_error.New(-65, "玩家数据格式错误，请联系客服")
	}

	// 将JSON字符串解析为map，以便添加到响应中
	var publicInfoMap map[string]interface{}
	if err := json.Unmarshal(publicInfoJSON, &publicInfoMap); err != nil {
		log.Printf("解析公开信息JSON数据失败: %v，返回错误", err)
		return nil, game_error.New(-65, "玩家数据格式错误，请联系客服")
	}

	// 使用解析后的JSON对象作为publicInfo
	
	// 获取雷达数据 - 统一获取雷达信息，避免重复代码
	var radarThief, radarPolice []int
	var radarRemainRoundPolice, radarRemainRoundThief int
	
	if isSelf {
		// 如果是查询自己的信息，直接使用playerData中的数据
		radarThief, radarPolice, radarRemainRoundPolice, radarRemainRoundThief, err = radarManager.ParseRadarInfoFromJSON(requestedPlayerData.PlayerRadar)
		if err != nil {
			log.Printf("解析雷达数据失败: %v，使用默认值", err)
			// 如果解析失败，使用默认雷达数据
			defaultRadarInfo := radarManager.GetDefaultRadarInfo()
			radarThief = defaultRadarInfo.RadarThief
			radarPolice = defaultRadarInfo.RadarPolice
			radarRemainRoundPolice = defaultRadarInfo.RadarRemainRoundPolice
			radarRemainRoundThief = defaultRadarInfo.RadarRemainRoundThief
		}
	} else {
		// 如果是查询他人信息，从数据库获取
		radarInfo, err := radarManager.GetRadarInfoByRoleID(requestRoleID)
		if err != nil {
			log.Printf("获取雷达数据失败: %v，使用默认值", err)
			// 如果获取失败，使用默认雷达数据
			defaultRadarInfo := radarManager.GetDefaultRadarInfo()
			radarThief = defaultRadarInfo.RadarThief
			radarPolice = defaultRadarInfo.RadarPolice
			radarRemainRoundPolice = defaultRadarInfo.RadarRemainRoundPolice
			radarRemainRoundThief = defaultRadarInfo.RadarRemainRoundThief
		} else {
			radarThief = radarInfo.RadarThief
			radarPolice = radarInfo.RadarPolice
			radarRemainRoundPolice = radarInfo.RadarRemainRoundPolice
			radarRemainRoundThief = radarInfo.RadarRemainRoundThief
		}
	}

	// 构建ownedCharacters - 从数据库中读取角色数据并转换为客户端需要的格式
	ownedCharacters := func() []map[string]interface{} {
		// 创建角色管理器
		cm := utils.NewCharacterManager()

		// 获取角色数据
		var characters []utils.Character
		var err error
		
		if isSelf {
			// 如果是查询自己的信息，直接使用playerData中的数据
			characters, err = cm.ParseCharactersFromDB(requestedPlayerData.OwnedCharacters)
		} 
		// else {
		// 	// 如果是查询他人信息，从数据库获取
		// 	characters, err = cm.GetCharacters(requestedPlayerData.DeviceID)
		// }
		
		if err != nil || len(characters) == 0 {
			log.Printf("获取角色数据失败: %v，使用默认值", err)
			// 使用默认角色数据
			defaultCharacters := cm.GetDefaultCharacters()
			if len(defaultCharacters) == 0 {
				log.Printf("角色数据为空，返回错误")
				// 不再使用默认值，而是直接返回空数组
				log.Printf("角色信息不存在，请重试")
				return []map[string]interface{}{}
			}
			characters = defaultCharacters
		}

		// 将角色数据转换为map格式，以便在响应中使用
		result := make([]map[string]interface{}, 0, len(characters))
		for _, character := range characters {
			characterMap := map[string]interface{}{}
			characterJSON, _ := json.Marshal(character)
			json.Unmarshal(characterJSON, &characterMap)
			result = append(result, characterMap)
		}
		return result
	}()

	// 构建ownedSkins - 使用SkinPartManager从数据库中读取皮肤数据并转换为客户端需要的格式
	ownedSkins := func() map[string]interface{} {
		// 创建皮肤部件管理器
		sm := utils.NewSkinPartManager()

		// 获取皮肤部件数据
		var skinPartIDs []string
		var skinPartColors []string
		var expiredTimes []int
		var skinDecals []interface{}
		var err error

		if isSelf {
			// 如果是查询自己的信息，直接使用playerData中的数据
			skinPartIDs, skinPartColors, expiredTimes, skinDecals, err = sm.ParseSkinPartsFromJSON(requestedPlayerData.OwnedSkins)
			if err != nil {
				log.Printf("解析皮肤部件数据失败: %v，使用默认值", err)
				// 使用默认皮肤部件数据
				defaultSkinParts := sm.GetDefaultSkinParts()
				skinPartIDs, skinPartColors, expiredTimes, skinDecals, _ = sm.ExtractSkinPartArrays(defaultSkinParts)
			}
		} 
		// else {
		// 	// 如果是查询他人信息，从数据库获取
		// 	// 获取对应的deviceID
		// 	var targetDeviceID string
		// 	if requestedPlayerData.DeviceID != "" {
		// 		targetDeviceID = requestedPlayerData.DeviceID
		// 	} else {
		// 		log.Printf("未找到玩家的deviceID，使用默认皮肤部件数据")
		// 		defaultSkinParts := sm.GetDefaultSkinParts()
		// 		skinPartIDs, skinPartColors, expiredTimes, skinDecals, _ = sm.ExtractSkinPartArrays(defaultSkinParts)
		// 		goto buildResult
		// 	}

		// 	// 获取皮肤部件数据
		// 	skinParts, err := sm.GetSkinParts(targetDeviceID)
		// 	if err != nil {
		// 		log.Printf("获取皮肤部件数据失败: %v，使用默认值", err)
		// 		defaultSkinParts := sm.GetDefaultSkinParts()
		// 		skinPartIDs, skinPartColors, expiredTimes, skinDecals, _ = sm.ExtractSkinPartArrays(defaultSkinParts)
		// 	} else {
		// 		skinPartIDs, skinPartColors, expiredTimes, skinDecals, _ = sm.ExtractSkinPartArrays(skinParts)
		// 	}
		// }

//	buildResult:
		// 构建结果
		return map[string]interface{}{
			"skinPartIDs":    skinPartIDs,
			"skinPartColors": skinPartColors,
			"expiredTime":    expiredTimes,
			"skinDecals":     skinDecals,
		}
	}()

	// 构建响应
	// 如果是查询自己的信息，返回完整数据；如果是查询他人信息，只返回公开信息和在线状态
	var responseData map[string]interface{}

	// 预先获取卡牌皮肤数据和卡牌样式数据，避免重复获取
	var cardSkins []utils.CardSkin
	var cardOwnSkins []int
	var cardSkinExpiredTimes []int
	var cardStyles []utils.CardStyle
	var cardOwnStyles []int
	var cardStyleExpiredTimes []int
	var cards []utils.Card
	
	// 预先获取卡牌数据，避免重复获取
	var cardIDs []int
	var cardLevels []int
	var cardCurSkins []interface{}
	var cardCurStyles []interface{}
	var interfaceCardIDs []interface{}

	if isSelf {
		// 直接使用playerData中的数据，而不是再次查询数据库
		var err error
		// 解析卡牌皮肤数据
		cardOwnSkins, cardSkinExpiredTimes, err = cardSkinManager.ParseCardSkinsFromJSON(requestedPlayerData.CardSkins)
		if err != nil {
			log.Printf("解析卡牌皮肤数据失败: %v，使用默认值", err)
			// 如果解析失败，使用默认卡牌皮肤数据
			cardSkins = cardSkinManager.GetDefaultCardSkins()
			cardOwnSkins, cardSkinExpiredTimes, _ = cardSkinManager.ExtractCardSkinArrays(cardSkins)
		}
		
		// 解析卡牌数据以获取cardIDs、cardLevels、cardCurSkins和cardCurStyles
		cardIDs, cardLevels, cardCurSkins, cardCurStyles, err = cardManager.ParseCardsFromJSON(requestedPlayerData.Cards)
		if err != nil {
			log.Printf("解析卡牌数据失败: %v，使用默认值", err)
			// 如果解析失败，使用默认卡牌数据
			cards = cardManager.GetDefaultCards()
			cardIDs, cardLevels, cardCurSkins, cardCurStyles, _ = cardManager.ExtractCardArrays(cards)
		}
		
		// 将int类型的cardIDs转换为interface{}类型
		interfaceCardIDs = make([]interface{}, len(cardIDs))
		for i, id := range cardIDs {
			interfaceCardIDs[i] = id
		}

		// 解析卡牌样式数据
		cardOwnStyles, cardStyleExpiredTimes, err = cardStyleManager.ParseCardStylesFromJSON(requestedPlayerData.CardStyles)
		if err != nil {
			log.Printf("解析卡牌样式数据失败: %v，使用默认值", err)
			// 如果解析失败，使用默认卡牌样式数据
			cardStyles = cardStyleManager.GetDefaultCardStyles()
			cardOwnStyles, cardStyleExpiredTimes, err = cardStyleManager.ExtractCardStyleArrays(cardStyles)
		}
	}

	if isSelf {
		// 返回完整的个人信息
		responseData = map[string]interface{}{
			"publicInfo":             publicInfoMap,
			"serverOverDayTimeStamp": serverOverDayTimeStamp,
			"gold":                   "170899",
			// 使用已获取的雷达信息
			"radarThief":             radarThief,
			"radarPolice":            radarPolice,
			"radarRemainRoundPolice": radarRemainRoundPolice,
			"radarRemainRoundThief":  radarRemainRoundThief,
			"description":            "",
			"province":               publicInfoMap["province"], // 从 publicInfoMap 获取 province
			"ownedCharacters":        ownedCharacters,
			"ownedSkins":             ownedSkins,
			"ownedSkinIDs":           []int{},
			"ownedSkinExpiredTime":   []int{},
			"coloringAgentNum":       "20",
			"personality":            "869",
			"personalityRank":        0,
			"onlineState":            1, // 实时生成的在线状态
			"diamonds":               2147483647,
			"tickets":                2147483647,
			"normalFortuneCards":     0,
			"advanceFortuneCards":    0,
			"activeCharacterID":      []int{100, 200},
			"activeRoleType":         "1",
			"recordVisible":          false,
			"likeCount":              0,
			"ownedAssets": func() []map[string]interface{} {
				// 创建资产管理器
				am := utils.NewAssetsManager()
				
				// 获取资产数据
				var assets []utils.Asset
				var err error
				
				if isSelf {
					// 如果是查询自己的信息，直接解析playerData中的数据
					assets, err = am.ParseAssetsFromJSON(requestedPlayerData.AssetsData)
				} else {
					// 如果是查询他人信息，使用默认资产数据
					defaultAssetsData := am.GetDefaultAssetsData()
					assets = defaultAssetsData.OwnedAssets
				}
				
				if err != nil {
					log.Printf("解析资产数据失败: %v，使用默认值", err)
					// 使用默认资产数据
					defaultAssetsData := am.GetDefaultAssetsData()
					assets = defaultAssetsData.OwnedAssets
				}
				
				// 将资产数据转换为map数组
				result := make([]map[string]interface{}, 0, len(assets))
				for _, asset := range assets {
					assetMap := map[string]interface{}{
						"itemID": asset.ItemID,
						"itemCount": asset.ItemCount,
					}
					result = append(result, assetMap)
				}
				

				
				return result
			}(),
			"activeHeadBoxID":   publicInfoObj.ActiveHeadBoxID,
			"activeBubbleBoxID": publicInfoObj.ActiveBubbleBoxID,
			// 使用BoxesManager获取装饰框数据
			"ownedHeadBoxes": func() map[string]interface{} {
				// 获取装饰框数据
				var boxesData *utils.BoxesData
				var err error
				
				if isSelf {
					// 如果是查询自己的信息，直接解析playerData中的数据
					boxesData, err = boxesManager.ParseBoxesDataFromJSON(requestedPlayerData.BoxesData)
				} else {
					// 如果是查询他人信息，使用默认装饰框数据
					boxesData = boxesManager.GetDefaultBoxesData()
				}
				
				if err != nil {
					log.Printf("解析装饰框数据失败: %v，使用默认值", err)
					boxesData = boxesManager.GetDefaultBoxesData()
				}
				
				// 将OwnedHeadBoxes转换为map
				return map[string]interface{}{
					"headBoxID":   boxesData.OwnedHeadBoxes.HeadBoxID,
					"expiredTime": boxesData.OwnedHeadBoxes.ExpiredTime,
				}
			}(),
			"ownedBubbleBoxes": func() map[string]interface{} {
				// 获取装饰框数据
				var boxesData *utils.BoxesData
				var err error
				
				if isSelf {
					// 如果是查询自己的信息，直接解析playerData中的数据
					boxesData, err = boxesManager.ParseBoxesDataFromJSON(requestedPlayerData.BoxesData)
				} else {
					// 如果是查询他人信息，使用默认装饰框数据
					boxesData = boxesManager.GetDefaultBoxesData()
				}
				
				if err != nil {
					log.Printf("解析装饰框数据失败: %v，使用默认值", err)
					boxesData = boxesManager.GetDefaultBoxesData()
				}
				
				// 将OwnedBubbleBoxes转换为map
				return map[string]interface{}{
					"bubbleBoxID": boxesData.OwnedBubbleBoxes.BubbleBoxID,
					"expiredTime": boxesData.OwnedBubbleBoxes.ExpiredTime,
				}
			}(),
			// 使用EmotionManager获取表情数据
			"ownedIngameEmotion": func() map[string]interface{} {
				// 获取表情数据
				var emotionData *utils.EmotionData
				var err error
				
				if isSelf {
					// 如果是查询自己的信息，直接解析playerData中的数据
					emotionData, err = emotionManager.ParseEmotionDataFromJSON(requestedPlayerData.EmotionData)
				} else {
					// 如果是查询他人信息，使用默认表情数据
					emotionData = emotionManager.GetDefaultEmotionData()
				}
				
				if err != nil {
					log.Printf("解析表情数据失败: %v，使用默认值", err)
					emotionData = emotionManager.GetDefaultEmotionData()
				}
				
				// 将OwnedIngameEmotion转换为map
				return map[string]interface{}{
					"id":          emotionData.OwnedIngameEmotion.ID,
					"expiredTime": emotionData.OwnedIngameEmotion.ExpiredTime,
				}
			}(),
			"ingameEmotionConfigs": func() []map[string]interface{} {
				// 获取表情数据
				var emotionData *utils.EmotionData
				var err error
				
				if isSelf {
					// 如果是查询自己的信息，直接解析playerData中的数据
					emotionData, err = emotionManager.ParseEmotionDataFromJSON(requestedPlayerData.EmotionData)
				} else {
					// 如果是查询他人信息，使用默认表情数据
					emotionData = emotionManager.GetDefaultEmotionData()
				}
				
				if err != nil {
					log.Printf("解析表情数据失败: %v，使用默认值", err)
					emotionData = emotionManager.GetDefaultEmotionData()
				}
				
				// 将EmotionConfig数组转换为map数组
				result := make([]map[string]interface{}, 0, len(emotionData.IngameEmotionConfigs))
				for _, config := range emotionData.IngameEmotionConfigs {
					configMap := map[string]interface{}{
						"character": config.Character,
						"config":    config.Config,
					}
					result = append(result, configMap)
				}
				return result
			}(),
			"hotPoint":           0,
			"sendGiftPoint":      0,
			"giftWars":           0,
			"giftIDs":            []int{},
			"giftCounts":         []int{},
			"followingNum":       0,
			"followerNum":        0,
			"visitorNum":         0,
			"hotPointLevel":      1,
			"sendGiftPointLevel": 1,
			"lightness": func() map[string]interface{} {
				// 创建炫光管理器
				lm := utils.NewLightnessManager()
				
				// 获取炫光数据
				var lightnessResult map[string]interface{}
				var err error
				
				if isSelf {
					// 如果是查询自己的信息，直接解析playerData中的数据
					lightnessResult, err = lm.ParseLightnessDataFromJSON(requestedPlayerData.LightnessData)
				} else {
					// 如果是查询他人信息，使用默认炫光数据
					defaultLightnessData := lm.GetDefaultLightnessData()
					// 将默认数据转换为map格式
					lightnessResult = map[string]interface{}{}
					defaultDataJSON, _ := json.Marshal(defaultLightnessData)
					json.Unmarshal(defaultDataJSON, &lightnessResult)
				}
				
				if err != nil {
					log.Printf("获取炫光数据失败: %v，使用默认值", err)
					// 使用默认炫光数据
					defaultLightnessData := lm.GetDefaultLightnessData()
					// 将默认数据转换为map格式
					lightnessResult = map[string]interface{}{}
					defaultDataJSON, _ := json.Marshal(defaultLightnessData)
					json.Unmarshal(defaultDataJSON, &lightnessResult)
				}
				
				return lightnessResult
			}(),
			"examGrade":             0,
			"passLevel":             "0",
			"reputationNum":         "100/100",
			"banCardSkin":           []int{},
			"banCharacterPartSkin":  []int{},
			"banCharacterSuitSkin":  []int{},
			"banCharacterGroupSkin": []int{},
			"customBuyConfig":       []int{1, 1, 1, 1, 1, 1, 1},
			"heros":                 []int{},
			"cardIDs":               interfaceCardIDs,
			"cardLevels":            cardLevels,
			"cardCurSkin":           cardCurSkins,
			"cardCurStyle": cardCurStyles,
			"cardPiece": []map[string]interface{}{
				{"cardID": 105, "num": 611},
				{"cardID": 200, "num": 1087},
				{"cardID": 103, "num": 774},
				{"cardID": 102, "num": 639},
				{"cardID": 108, "num": 560},
				{"cardID": 109, "num": 98},
				{"cardID": 101, "num": 0},
				{"cardID": 104, "num": 175},
				{"cardID": 210, "num": 72},
				{"cardID": 107, "num": 121},
				{"cardID": 100, "num": 507},
				{"cardID": 106, "num": 306},
				{"cardID": 110, "num": 62},
				{"cardID": 111, "num": 105},
				{"cardID": 112, "num": 168},
				{"cardID": 114, "num": 313},
				{"cardID": 113, "num": 8},
				{"cardID": 115, "num": 103},
				{"cardID": 230, "num": 1021},
				{"cardID": 116, "num": 7},
				{"cardID": 117, "num": 257},
				{"cardID": 118, "num": 239},
				{"cardID": 119, "num": 148},
				{"cardID": 120, "num": 70},
				{"cardID": 240, "num": 152},
				{"cardID": 121, "num": 0},
				{"cardID": 250, "num": 136},
				{"cardID": 122, "num": 76},
				{"cardID": 123, "num": 151},
				{"cardID": 124, "num": 270},
				{"cardID": 125, "num": 862},
				{"cardID": 280, "num": 351},
				{"cardID": 126, "num": 777},
				{"cardID": 127, "num": 973},
				{"cardID": 290, "num": 1352},
				{"cardID": 128, "num": 15},
				{"cardID": 129, "num": 40},
			},
			// 获取卡牌皮肤数据 - 只获取一次数据
			"cardOwnSkin": cardOwnSkins,
			"cardSkinExpiredTime": cardSkinExpiredTimes,
			"cardOwnStyle": cardOwnStyles,
			"cardStyleExpiredTime": cardStyleExpiredTimes,
			"isSelf": isSelf, // 实时生成的是否为自己
		}
	} else {
		// 返回其他玩家的公开信息、在线状态和雷达数据
		// 获取雷达数据
		
		responseData = map[string]interface{}{
			"publicInfo":  publicInfoMap,
			"onlineState": 1, // 实时生成的在线状态
			"isSelf":      isSelf,
			"radarThief":             radarThief,
			"radarPolice":            radarPolice,
			"radarRemainRoundPolice": radarRemainRoundPolice,
			"radarRemainRoundThief":  radarRemainRoundThief,
		}
	}

	return responseData, nil
}

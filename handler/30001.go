// internal/handler/30001.go
package handler

import (
	"crypto/md5"
	"dmmserver/db"
	"dmmserver/game_error"
	"dmmserver/model"
	"dmmserver/services/playtime"
	"dmmserver/services/serversettings"
	"dmmserver/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	//	"gorm.io/gorm" // Add gorm import
)

func init() {
	Register("30001", handle30001)
}

// createDefaultCards 创建默认卡牌数据并返回JSON字符串
func createDefaultCards() (string, error) {
	// 使用CardManager获取默认卡牌数据
	cm := utils.NewCardManager()
	defaultCards := cm.GetDefaultCards()

	// 将卡牌数据转换为JSON字符串
	cardsJSON, err := json.Marshal(defaultCards)
	if err != nil {
		log.Printf("序列化默认卡牌数据失败: %v", err)
		return "", err
	}

	return string(cardsJSON), nil
}

// createDefaultSkins 创建默认皮肤数据并返回JSON字符串
func createDefaultSkins() (string, error) {
	// 使用SkinPartManager获取默认皮肤部件数据
	sm := utils.NewSkinPartManager()
	defaultSkinParts := sm.GetDefaultSkinParts()

	// 将皮肤部件数据转换为JSON字符串
	skinsJSON, err := json.Marshal(defaultSkinParts)
	if err != nil {
		log.Printf("序列化默认皮肤数据失败: %v", err)
		return "", err
	}

	return string(skinsJSON), nil
}

// 注意：createDefaultPublicInfo函数已被移除，使用utils.PublicInfoManager代替
// PublicInfoManager提供了完整的玩家公开信息管理功能，包括获取、保存和更新公开信息

// 注意：createDefaultCharacters函数已被移除，使用utils.CharacterManager代替

// handle30001 处理登录请求并返回账户/服务器信息。
func handle30001(c *gin.Context, msgData map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing handler for msg_id=30001. Received msgData: %+v", msgData)

	// 1. 参数解析
	// ----------------------
	deviceID, _ := msgData["deviceID"].(string)
	loginKeyReq, _ := msgData["loginKey"].(string) // 请求中的 loginKey
	// openIDReq, _ := msgData["openID"].(string) // 请求中的 openID，可能为空

	// 基本验证
	if deviceID == "" {
		log.Println("错误：msg_id=30001 缺少 'deviceID' 参数")
		return nil, game_error.New(-5, "缺少 'deviceID' 参数")
	}

	// 在处理请求前检查设置是否过期
	serversettings.CheckAndRefreshIfStale()

	// 2. 业务逻辑 (数据库交互和数据生成)
	// -------------------------------------------------------------
	// 根据 deviceID 查询 dmm_playerdata
	var playerData model.PlayerData
	result := db.DB.Where("device_id = ?", deviceID).First(&playerData)
	if result.Error != nil {
		log.Printf("未找到 deviceID 为 '%s' 的玩家。正在创建默认数据。", deviceID)
		//		playerExists = false

		// 创建默认玩家数据

		// 查询当前最大的 RoleID
		var maxRoleID int
		if err := db.DB.Model(&model.PlayerData{}).Select("COALESCE(MAX(role_id), 0)").Row().Scan(&maxRoleID); err != nil {
			log.Printf("查询最大 RoleID 失败: %v", err)
			return nil, game_error.New(-2, "数据库查询错误")
		}

		// 生成初始 authKey
		hasher1 := md5.New()
		hasher1.Write([]byte(deviceID + "_salt1"))
		part1 := hex.EncodeToString(hasher1.Sum(nil))

		hasher2 := md5.New()
		hasher2.Write([]byte(fmt.Sprintf("%s_%d_salt2", loginKeyReq, time.Now().Unix())))
		part2 := hex.EncodeToString(hasher2.Sum(nil))

		// 创建默认卡牌数据
		cardsJSON, err := createDefaultCards()
		if err != nil {
			log.Printf("创建默认卡牌数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 创建默认皮肤数据
		skinsJSON, err := createDefaultSkins()
		if err != nil {
			log.Printf("创建默认皮肤数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用CharacterManager创建默认角色数据
		cm := utils.NewCharacterManager()
		defaultCharacters := cm.GetDefaultCharacters()
		charactersJSON, err := json.Marshal(defaultCharacters)
		if err != nil {
			log.Printf("序列化默认角色数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用CardSkinManager创建默认卡牌皮肤数据
		csm := utils.NewCardSkinManager()
		defaultCardSkins := csm.GetDefaultCardSkins()
		cardSkinsJSON, err := json.Marshal(defaultCardSkins)
		if err != nil {
			log.Printf("序列化默认卡牌皮肤数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用CardStyleManager创建默认卡牌样式数据
		cstm := utils.NewCardStyleManager()
		defaultCardStyles := cstm.GetDefaultCardStyles()
		cardStylesJSON, err := json.Marshal(defaultCardStyles)
		if err != nil {
			log.Printf("序列化默认卡牌样式数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用PublicInfoManager创建默认公开信息
		pm := utils.NewPublicInfoManager()
		defaultPublicInfo := pm.GetDefaultPublicInfo()
		publicInfoStr, err := pm.ConvertPublicInfoToKeyValue(&defaultPublicInfo)
		if err != nil {
			log.Printf("转换默认公开信息为键值对格式失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用EmotionManager创建默认表情数据
		em := utils.NewEmotionManager()
		defaultEmotionData := em.GetDefaultEmotionData()
		emotionDataJSON, err := json.Marshal(defaultEmotionData)
		if err != nil {
			log.Printf("序列化默认表情数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 初始化玩家游玩时长数据
		playtimeData, err := playtime.GetPlayerPlaytimeData(deviceID)
		if err != nil {
			log.Printf("获取玩家游玩时长数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}
		playtimeDataJSON, err := json.Marshal(playtimeData)
		if err != nil {
			log.Printf("序列化游玩时长数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用AssetsManager创建默认资产数据
		am := utils.NewAssetsManager()
		defaultAssetsData := am.GetDefaultAssetsData()
		assetsDataJSON, err := json.Marshal(defaultAssetsData)
		if err != nil {
			log.Printf("序列化资产数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用BoxesManager创建默认装饰框数据
		bm := utils.NewBoxesManager()
		defaultBoxesData := bm.GetDefaultBoxesData()
		boxesDataJSON, err := json.Marshal(defaultBoxesData)
		if err != nil {
			log.Printf("序列化装饰框数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 使用LightnessManager创建默认炫光数据
		lm := utils.NewLightnessManager()
		defaultLightnessData := lm.GetDefaultLightnessData()
		lightnessDataJSON, err := json.Marshal(defaultLightnessData)
		if err != nil {
			log.Printf("序列化炫光数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 创建新玩家数据
		playerData = model.PlayerData{
			DeviceID:            deviceID,
			RoleID:              maxRoleID + 1,
			OpenID:              "324438392",
			AuthKey:             fmt.Sprintf("%s_%s", part1, part2),
			AuthKeyExpire:       time.Now().Add(2 * time.Hour).Unix(),
			Audit:               1,
			CreateAccountTime:   time.Now().Unix(),
			AccountSafe:         false,
			NotSafe:             false,
			OpenIDMatched:       true,
			CustomAccount:       "",
			GuideLevel:          0,
			IsSetPwd:            false,
			Mail:                "",
			ReputationScore:     100,
			ReputationLimitTime: 0,
			InspectorLevel:      0,
			PublicInfo:          publicInfoStr,         // 添加公开信息数据
			Cards:               string(cardsJSON),      // 添加卡牌数据
			OwnedSkins:          string(skinsJSON),      // 添加皮肤数据
			OwnedCharacters:     string(charactersJSON), // 添加角色数据
			CardSkins:           string(cardSkinsJSON),  // 添加卡牌皮肤数据
			CardStyles:          string(cardStylesJSON), // 添加卡牌样式数据
			EmotionData:         string(emotionDataJSON), // 添加表情数据
			PlaytimeData:        string(playtimeDataJSON), // 添加游玩时长数据
			AssetsData:          string(assetsDataJSON),   // 添加资产数据
			BoxesData:           string(boxesDataJSON),    // 添加装饰框数据
			LightnessData:       string(lightnessDataJSON), // 添加炫光数据
		}

		if result := db.DB.Create(&playerData); result.Error != nil {
			log.Printf("创建新玩家数据失败: %v", result.Error)
			return nil, game_error.New(-2, "数据库写入错误")
		}
		log.Printf("成功创建 deviceID 为 '%s' 的新玩家数据", deviceID)
	} else {
		log.Printf("找到 deviceID 为 '%s' 的玩家", deviceID)
	}

	// 查询 dmm_settings 获取全局服务器设置
	// var serverSettings model.ServerSettings
	// result = db.DB.First(&serverSettings)

	// 使用缓存的服务器设置
	serverSettings := serversettings.GetSettings()

	// if result.Error != nil {
	//	// 检查是否是记录未找到错误
	//	if result.Error == gorm.ErrRecordNotFound {
	//		log.Println("未找到服务器设置。正在创建默认设置。")

	//		// 定义默认的 GraphicsOptions 和 MiscOptions
	//		defaultGraphicsOptions := `[
	//			{"level": 1, "isDefault": 0, "shadow": 0, "maxParticles": 1000, "renderScale": 0.8},
	//			{"level": 3, "isDefault": 1, "shadow": 1, "maxParticles": 5000, "renderScale": 1}
	//		]`
	//		defaultMiscOptions := `{
	//			"outline": 1,
	//			"HFR": 1,
	//			"BRHFR": 1,
	//			"HFX": 1
	//		}`

	//		// 创建默认服务器设置数据
	//		serverSettings = model.ServerSettings{
	//			GraphicsOptions: defaultGraphicsOptions,
	//			MiscOptions:     defaultMiscOptions,
	//		}

	//		// 将新设置数据持久化到数据库
	//		createResult := db.DB.Create(&serverSettings)
	//		if createResult.Error != nil {
	//			log.Printf("创建默认服务器设置失败: %v", createResult.Error)
	//			return nil, game_error.New(-2, "数据库写入错误")
	//		}
	//		log.Println("成功创建默认服务器设置")

	//	} else {
	//		// 如果是其他数据库错误，则返回错误
	//		log.Printf("获取服务器设置失败: %v", result.Error)
	//		return nil, game_error.New(-1, "数据库连接错误，请重新登录")
	//	}
	// }

	// 解析JSON格式的GraphicsOptions和MiscOptions
	// var graphicsOptions []map[string]interface{}
	// var miscOptions map[string]interface{}

	// if err := json.Unmarshal([]byte(serverSettings.GraphicsOptions), &graphicsOptions); err != nil {
	//	// 如果解析失败，使用默认值
	//	log.Printf("解析GraphicsOptions失败: %v，使用默认值", err)
	//	graphicsOptions = []map[string]interface{}{
	//		{"level": 1, "isDefault": 0, "shadow": 0, "maxParticles": 1000, "renderScale": 0.8},
	//		{"level": 3, "isDefault": 1, "shadow": 1, "maxParticles": 5000, "renderScale": 1},
	//	}
	// }

	// if err := json.Unmarshal([]byte(serverSettings.MiscOptions), &miscOptions); err != nil {
	//	// 如果解析失败，使用默认值
	//	log.Printf("解析MiscOptions失败: %v，使用默认值", err)
	//	miscOptions = map[string]interface{}{
	//		"outline": 1,
	//		"HFR":     1,
	//		"BRHFR":   1,
	//		"HFX":     1,
	//	}
	// }

	// 使用serversettings服务中的解析函数
	graphicsOptions := serversettings.ParseGraphicsOptions(serverSettings.GraphicsOptions)
	miscOptions := serversettings.ParseMiscOptions(serverSettings.MiscOptions)

	// 生成动态数据
	serverTimeStamp := time.Now().Unix()

	// 生成 authKey (deviceID + loginKey + 时间戳 + 盐的MD5值)
	// 第一部分是 MD5(deviceID + salt1)
	hasher1 := md5.New()
	hasher1.Write([]byte(deviceID + "_salt1"))
	part1 := hex.EncodeToString(hasher1.Sum(nil))

	// 第二部分是 MD5(loginKey + 时间戳 + salt2)
	hasher2 := md5.New()
	hasher2.Write([]byte(fmt.Sprintf("%s_%d_salt2", loginKeyReq, serverTimeStamp)))
	part2 := hex.EncodeToString(hasher2.Sum(nil))
	authKey := fmt.Sprintf("%s_%s", part1, part2)

	// 计算 authKey 失效时间戳 (当前时间 + 1.9小时)
	authKeyExpire := time.Now().Add(time.Duration(1.9 * float64(time.Hour))).Unix()

	// 更新玩家数据中的 authKey 和失效时间戳
	playerData.AuthKey = authKey
	playerData.AuthKeyExpire = authKeyExpire

	// 检查cards字段是否为空，如果为空则设置默认卡组数据
	if playerData.Cards == "" || playerData.Cards == "null" {
		// 创建默认卡牌数据
		cardsJSON, err := createDefaultCards()
		if err != nil {
			log.Printf("创建默认卡牌数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的卡牌数据
		playerData.Cards = cardsJSON
		log.Printf("玩家 '%s' 的卡牌数据为空，已设置默认卡组数据", deviceID)
	}

	// 检查OwnedSkins字段是否为空，如果为空则设置默认皮肤数据
	if playerData.OwnedSkins == "" || playerData.OwnedSkins == "null" {
		// 创建默认皮肤数据
		skinsJSON, err := createDefaultSkins()
		if err != nil {
			log.Printf("创建默认皮肤数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的皮肤数据
		playerData.OwnedSkins = skinsJSON
		log.Printf("玩家 '%s' 的皮肤数据为空，已设置默认皮肤数据", deviceID)
	}

	// 检查OwnedCharacters字段是否为空，如果为空则设置默认角色数据
	if playerData.OwnedCharacters == "" || playerData.OwnedCharacters == "null" {
		// 使用CharacterManager设置默认角色数据
		cm := utils.NewCharacterManager()
		defaultCharacters := cm.GetDefaultCharacters()
		charactersJSON, err := json.Marshal(defaultCharacters)
		if err != nil {
			log.Printf("序列化角色数据失败: %v", err)
			// 如果序列化失败，使用硬编码的默认值
			charactersJSON = []byte(`[{"characterID":100,"expiredTime":0,"currentSkinInfo":{"skinPartIDs":["1001","1002","1003","1004","1005"],"skinPartColors":["67","103","84","110","119"],"skinDecals":[0,0,0,0,0]},"ExpLevel":999999999,"ExpPoint":0,"TalentPointRemained":0,"TalentLevels":[4,4,4],"weaponSkinID":0}]`)
		}
		playerData.OwnedCharacters = string(charactersJSON)
		log.Printf("玩家 '%s' 的角色数据为空，已设置默认角色数据", deviceID)
	}

	// 检查CardSkins字段是否为空，如果为空则设置默认卡牌皮肤数据
	if playerData.CardSkins == "" || playerData.CardSkins == "null" {
		// 使用CardSkinManager设置默认卡牌皮肤数据
		csm := utils.NewCardSkinManager()
		defaultCardSkins := csm.GetDefaultCardSkins()
		cardSkinsJSON, err := json.Marshal(defaultCardSkins)
		if err != nil {
			log.Printf("序列化卡牌皮肤数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的卡牌皮肤数据
		playerData.CardSkins = string(cardSkinsJSON)
		log.Printf("玩家 '%s' 的卡牌皮肤数据为空，已设置默认卡牌皮肤数据", deviceID)
	}

	// 检查CardStyles字段是否为空，如果为空则设置默认卡牌样式数据
	if playerData.CardStyles == "" || playerData.CardStyles == "null" {
		// 使用CardStyleManager设置默认卡牌样式数据
		cstm := utils.NewCardStyleManager()
		defaultCardStyles := cstm.GetDefaultCardStyles()
		cardStylesJSON, err := json.Marshal(defaultCardStyles)
		if err != nil {
			log.Printf("序列化卡牌样式数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的卡牌样式数据
		playerData.CardStyles = string(cardStylesJSON)
		log.Printf("玩家 '%s' 的卡牌样式数据为空，已设置默认卡牌样式数据", deviceID)
	}

	// 检查PublicInfo字段是否为空，如果为空则设置默认公开信息数据
	if playerData.PublicInfo == "" || playerData.PublicInfo == "null" {
		// 使用PublicInfoManager创建默认公开信息
		pm := utils.NewPublicInfoManager()
		defaultPublicInfo := pm.GetDefaultPublicInfo()
		publicInfoStr, err := pm.ConvertPublicInfoToKeyValue(&defaultPublicInfo)
		if err != nil {
			log.Printf("转换默认公开信息为键值对格式失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}
		
		

		// 更新玩家的公开信息数据
		playerData.PublicInfo = publicInfoStr
		log.Printf("玩家 '%s' 的公开信息数据为空，已设置默认公开信息数据", deviceID)
	}

	// 检查PlaytimeData字段是否为空，如果为空则设置默认游玩时长数据
	if playerData.PlaytimeData == "" || playerData.PlaytimeData == "null" {
		// 初始化玩家游玩时长数据
		playtimeData, err := playtime.GetPlayerPlaytimeData(deviceID)
		if err != nil {
			log.Printf("获取玩家游玩时长数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}
		playtimeDataJSON, err := json.Marshal(playtimeData)
		if err != nil {
			log.Printf("序列化游玩时长数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的游玩时长数据
		playerData.PlaytimeData = string(playtimeDataJSON)
		log.Printf("玩家 '%s' 的游玩时长数据为空，已设置默认游玩时长数据", deviceID)
	}

	// 检查PlayerRadar字段是否为空，如果为空则设置默认雷达数据
	if playerData.PlayerRadar == "" || playerData.PlayerRadar == "null" {
		// 使用RadarManager创建默认雷达信息
		rm := utils.NewRadarManager()
		defaultRadarInfo := rm.GetDefaultRadarInfo()
		radarInfoStr, err := rm.ConvertRadarInfoToKeyValue(&defaultRadarInfo)
		if err != nil {
			log.Printf("转换默认雷达信息为键值对格式失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的雷达数据
		playerData.PlayerRadar = radarInfoStr
		log.Printf("玩家 '%s' 的雷达数据为空，已设置默认雷达数据", deviceID)
	}
	
	// 检查AssetsData字段是否为空，如果为空则设置默认资产数据
	if playerData.AssetsData == "" || playerData.AssetsData == "null" {
		// 使用AssetsManager创建默认资产数据
		am := utils.NewAssetsManager()
		defaultAssetsData := am.GetDefaultAssetsData()
		assetsDataJSON, err := json.Marshal(defaultAssetsData)
		if err != nil {
			log.Printf("序列化资产数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的资产数据
		playerData.AssetsData = string(assetsDataJSON)
		log.Printf("玩家 '%s' 的资产数据为空，已设置默认资产数据", deviceID)
	}

	// 检查BoxesData字段是否为空，如果为空则设置默认装饰框数据
	if playerData.BoxesData == "" || playerData.BoxesData == "null" {
		// 使用BoxesManager创建默认装饰框数据
		bm := utils.NewBoxesManager()
		defaultBoxesData := bm.GetDefaultBoxesData()
		boxesDataJSON, err := json.Marshal(defaultBoxesData)
		if err != nil {
			log.Printf("序列化装饰框数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}
		
		// 更新玩家的装饰框数据
		playerData.BoxesData = string(boxesDataJSON)
		log.Printf("玩家 '%s' 的装饰框数据为空，已设置默认装饰框数据", deviceID)
	}

	// 检查LightnessData字段是否为空，如果为空则设置默认炫光数据
	if playerData.LightnessData == "" || playerData.LightnessData == "null" {
		// 使用LightnessManager创建默认炫光数据
		lm := utils.NewLightnessManager()
		defaultLightnessData := lm.GetDefaultLightnessData()
		lightnessDataJSON, err := json.Marshal(defaultLightnessData)
		if err != nil {
			log.Printf("序列化炫光数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}
		
		// 更新玩家的炫光数据
		playerData.LightnessData = string(lightnessDataJSON)
		log.Printf("玩家 '%s' 的炫光数据为空，已设置默认炫光数据", deviceID)
	}

	// 检查EmotionData字段是否为空，如果为空则设置默认表情数据
	if playerData.EmotionData == "" || playerData.EmotionData == "null" {
		// 使用EmotionManager创建默认表情数据
		em := utils.NewEmotionManager()
		defaultEmotionData := em.GetDefaultEmotionData()
		emotionDataJSON, err := json.Marshal(defaultEmotionData)
		if err != nil {
			log.Printf("序列化默认表情数据失败: %v", err)
			return nil, game_error.New(-2, "数据处理错误")
		}

		// 更新玩家的表情数据
		playerData.EmotionData = string(emotionDataJSON)
		log.Printf("玩家 '%s' 的表情数据为空，已设置默认表情数据", deviceID)
	}

	// 将更新后的玩家数据保存到数据库
	updateResult := db.DB.Save(&playerData)
	if updateResult.Error != nil {
		log.Printf("更新玩家数据失败: %v", updateResult.Error)
		return nil, game_error.New(-2, "数据库更新错误")
	}

	// 3. 成功响应构建
	// --------------------------------
	// 使用PublicInfoManager获取玩家名字和年龄
	pm := utils.NewPublicInfoManager()
	publicInfo, err := pm.GetPublicInfo(deviceID)
	if err != nil {
		log.Printf("获取玩家公开信息失败: %v", err)
		return nil, game_error.New(-3, "未找到玩家数据")
	}
	
	responseData := map[string]interface{}{
		"serverTimeStamp":      serverTimeStamp,
		"serverIp":             serverSettings.ServerIP,
		"serverPort":           serverSettings.ServerPort,
		"roleID":               playerData.RoleID,
		"openID":               playerData.OpenID,
		"authKey":              authKey,
		"accountName":          publicInfo.Name,
		"loginKey":             loginKeyReq, // 返回请求中的 loginKey
		"audit":                playerData.Audit,
		"age":                  publicInfo.Age,
		"createAccountTime":    fmt.Sprintf("%d", playerData.CreateAccountTime),
		"accountSafe":          playerData.AccountSafe,
		"notSafe":              playerData.NotSafe,
		"openIDMatched":        playerData.OpenIDMatched,
		"customAccount":        playerData.CustomAccount,
		"serverId":             "15",     // 实时生成的ServerID
		"serverVersion":        20240101, // 实时生成的ServerVersion
		"guideLevel":           playerData.GuideLevel,
		"graphicsOptions":      graphicsOptions,
		"miscOptions":          miscOptions,
		"isSetPwd":             playerData.IsSetPwd,
		"serverTimeZoneOffset": 8, // 实时生成的ServerTimeZoneOffset，这里设置为东八区,
		"mail":                 playerData.Mail,
		"reputationScore":      playerData.ReputationScore,
		"reputationLimitTime":  playerData.ReputationLimitTime,
		"inspector_level":      playerData.InspectorLevel,
	}

	log.Printf("Successfully processed msg_id=30001 for deviceID '%s'. Response: %+v", deviceID, responseData)
	return responseData, nil
}

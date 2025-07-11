// internal/model/player.go
package model

import (
	"time"
)

// PlayerData 表示dmm_playerdata表的结构
type PlayerData struct {
	DeviceID           string `gorm:"primaryKey"`
	RoleID             int
	OpenID             string
//	AccountName        string
	Audit              int
	CreateAccountTime  int64
	AccountSafe        bool
	NotSafe            bool
	OpenIDMatched      bool
	CustomAccount      string
	GuideLevel         int
	IsSetPwd           bool
	Mail               string
	ReputationScore    int
	ReputationLimitTime int
	InspectorLevel     int
	// 玩家公开信息，使用键值对格式存储，包含name、icon、sex等公开属性
	PublicInfo         string `gorm:"type:text"` // 格式为 name="Notitle"\n\nicon=0\n\nsex=0\n\n...
	// 卡牌数据，使用JSON格式存储，每个卡牌包含id、level和curSkin三个属性
	Cards              string `gorm:"type:json"` // 格式为 [{"id":100,"level":13,"curSkin":600981}, ...]
	// 皮肤数据，使用JSON格式存储，每个皮肤包含skinPartIDs、skinPartColors、expiredTime和skinDecals四个属性
	OwnedSkins         string `gorm:"type:json"` // 格式为 [{"skinPartIDs":"1001","skinPartColors":"67","expiredTime":0,"skinDecals":0}, ...]
	// 角色数据，使用JSON格式存储，包含characterID、expiredTime、currentSkinInfo、ExpLevel等属性
	OwnedCharacters    string `gorm:"type:json"` // 格式为 [{"characterID":100,"expiredTime":0,"currentSkinInfo":{...},"ExpLevel":999999999,...}, ...]
	// 卡牌皮肤数据，使用JSON格式存储，包含cardOwnSkin和cardSkinExpiredTime两个属性
	CardSkins          string `gorm:"type:json"` // 格式为 [{"cardOwnSkin":601826,"cardSkinExpiredTime":0}, ...]
	// 卡牌样式数据，使用JSON格式存储，包含cardOwnStyle和cardStyleExpiredTime两个属性
	CardStyles         string `gorm:"type:json"` // 格式为 [{"cardOwnStyle":650041,"cardStyleExpiredTime":0}, ...]
	// 玩家雷达数据，使用键值对格式存储，包含radarThief、radarPolice、radarRemainRoundPolice和radarRemainRoundThief
	PlayerRadar        string `gorm:"type:text"` // 格式为 radarThief=[64,64,36,4,4]\n\nradarPolice=[]\n\nradarRemainRoundPolice=1\n\nradarRemainRoundThief=0
	// 表情数据，使用JSON格式存储，包含ownedIngameEmotion和ingameEmotionConfigs
	EmotionData        string `gorm:"type:json"` // 格式为 {"ownedIngameEmotion":{"id":[950001,960001,...],"expiredTime":[0,0,...]},"ingameEmotionConfigs":[{"character":100,"config":["950001","960701",...]},...]}
	// 资产数据，使用JSON格式存储，包含ownedAssets
	AssetsData         string `gorm:"type:json"` // 格式为 {"ownedAssets":[{"itemID":1,"itemCount":10},...]}
	// 装饰框数据，使用JSON格式存储，包含ownedHeadBoxes和ownedBubbleBoxes
	BoxesData          string `gorm:"type:json"` // 格式为 {"ownedHeadBoxes":{"headBoxID":[900001,...],"expiredTime":[0,...]},"ownedBubbleBoxes":{"bubbleBoxID":[910001,...],"expiredTime":[0,...]}}
	// 炫光数据，使用JSON格式存储，包含id、expiredTime和config
	LightnessData      string `gorm:"type:json"` // 格式为 {"id":[1001,1002,...],"expiredTime":[0,0,...],"config":1}
	// 游玩时长数据，使用JSON格式存储
	PlaytimeData       string `gorm:"type:json"` // 格式为 {"playedTime":0,"isVIP":false,"dailyPlayTime":5400,"todayExtraTime":0,"deviceID":"xxx","realDeviceID":"xxx","ip":"xxx"}
	// 其他未被现有工具管理的数据，使用JSON格式存储
//	Others             string `gorm:"type:json"` // 格式为 {"field1":value1,"field2":value2,...}
	AuthKey            string
	AuthKeyExpire      int64
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (PlayerData) TableName() string {
	return "dmm_playerdata"
}
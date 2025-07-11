// internal/bootstrap/bootstrap.go
package bootstrap

import (
	"dmmserver/conf"
	"dmmserver/db"
	_ "dmmserver/handler" // 【关键】匿名导入handler包以触发其下所有文件的init()函数
	"dmmserver/server"
	"dmmserver/services/banning"
	"dmmserver/services/playtime"
	"dmmserver/services/serversettings"
)

// Run 启动服务器的完整流程
func Run() {
	// 1. 加载配置
	conf.Init()

	// 2. 初始化数据库连接并自动建表
	db.InitDB()

	// 3. 初始化后台服务模块（加载封禁列表并启动智能刷新协程）
	banning.Init()

	// 初始化服务器设置模块（加载设置并启动智能刷新协程）
	serversettings.Init()

	// 初始化游戏时长控制模块（加载游戏时长数据并启动智能刷新和每日重置协程）
	playtime.Init()

	// 4. 所有准备工作完成，最后启动Web服务器。
	//    handler的注册已通过上面的匿名导入自动完成。
	server.Run()
}

---

# "躲猫猫大作战" 游戏大厅服务器后端 (Go)

这是一个基于 Go 语言构建的高性能、高扩展性的游戏大厅服务器后端。项目深度借鉴了优秀开源项目 [Alist](https://github.com/alist-org/alist) 的模块化设计思想，旨在提供一个稳健、可维护且易于二次开发的服务器框架。

## ✨ 项目特色

- **高度模块化**: 采用基于`msg_id`的自注册处理器（Handler）系统，新增业务逻辑就像添加一个新文件一样简单，无需修改任何核心代码。
- **配置驱动**: 核心行为由配置文件驱动，如服务器端口、数据库连接，以及前置安全策略的响应格式。
- **关注点分离 (SoC)**: 严格分层设计，将网络I/O、业务逻辑、数据访问、后台服务等完全解耦。
    - **业务逻辑与传输解耦**: `Handler` 只负责处理纯业务逻辑并返回数据，由`Dispatcher`统一负责响应的格式化与发送。
    - **安全策略与业务解耦**: 封禁（Banning）等通用安全检查作为前置中间件实现，业务代码无需关心。
- **智能后台服务**: 实现了按需刷新的智能封禁列表管理器，在保证数据安全的同时，避免了服务器空闲时不必要的数据库轮询。
- **统一错误处理**: 建立了全局的`game_error`中心，业务代码只需返回预定义的错误，即可由框架生成统一格式的错误响应。
- **数据库自迁移**: 使用 GORM 的`AutoMigrate`功能，在服务器启动时自动检查并创建不存在的数据表，简化部署流程。
- **可扩展的数据模型**: 广泛使用`JSON`数据类型来存储复杂的游戏数据（如玩家资产、卡牌配置等），在简化数据表结构的同时，保证了数据读取的性能。

## 🏗️ 项目架构

项目遵循了 Go 社区推荐的“标准项目布局”，并进行了定制化调整，以保证其清晰性和可维护性。

```
unity-lobby-server/
├── main.go                      # 程序的唯一入口
├── internal/
│   ├── bootstrap/               # 启动协调器，负责按顺序初始化所有模块
│   │   └── bootstrap.go
│   ├── conf/                    # 配置中心
│   │   ├── conf.go              # 加载 config.json 文件
│   │   └── response_format.go   # 【代码配置】定义不同msg_id的封禁响应格式
│   ├── db/                      # 数据库模块，负责初始化GORM和自动迁移
│   │   └── db.go
│   ├── game_error/              # 全局游戏错误码定义中心
│   │   └── errors.go
│   ├── handler/                 # 核心业务逻辑处理器
│   │   ├── dispatcher.go        # Handler的注册与分发器
│   │   ├── all.go               # 用于激活所有handler的"all"包（确保init被调用）
│   │   └── 30008.go             # 业务逻辑处理器示例
│   ├── model/                   # GORM数据模型定义
│   │   └── ban.go
│   │   └── player.go
│   ├── server/                  # 网络层，负责HTTP服务器、路由和中间件
│   │   └── server.go
│   ├── services/                # 后台服务模块（非直接响应请求）
│   │   └── banning/             # 账号封禁列表管理服务
│   └── utils/                   # 通用工具函数
└── configs/
    └── config.json              # 运行时配置文件
```

## 🚀 核心工作流

理解一个请求的完整生命周期是掌握此架构的关键。

#### 1. 启动流程 (`bootstrap`)
1.  加载 `configs/config.json`。
2.  初始化数据库连接 (`db`)，并自动迁移`model`中定义的表。
3.  **匿名导入 `handler` 包**，触发该包下所有 `*.go` 文件的 `init()` 函数，完成所有`msg_id`处理器的“自注册”。
4.  初始化后台服务，如 `banning` 服务，加载初始封禁列表并启动智能刷新协程。
5.  启动 `Gin` Web 服务器 (`server`)，开始监听端口。

#### 2. 请求处理流程 (`server`)
*（一个典型的中间件流程图可以很好地诠释本项目的请求处理方式）*

一个来自Unity客户端的`POST`请求到达服务器：
1.  **`activityMiddleware`**: 第一个中间件被触发。它通知`banning`服务“有活动发生”，并检查封禁列表是否“过时”，如果过时则强制刷新。
2.  **`banningEnforcementMiddleware`**: 第二个中间件接管。
    -   它解析出`msg_id`、`IP`、`DeviceID`等信息。
    -   依次检查这些信息是否命中封禁列表。
    -   **如果命中**：它会查询`conf/response_format.go`中的配置，判断应返回`JSON`还是`Text`，然后直接构造响应并**中止（Abort）**请求。请求不会再向后传递。
    -   **如果未命中**：调用`c.Next()`，将请求放行。
3.  **`dispatchHandler`**: 最终的处理器被执行。
    -   它从中间件设置的上下文中获取`msg_id`。
    -   通过`handler.GetHandler(msgID)`找到已注册的、对应的业务处理器函数。
    -   调用该函数，并接收返回的 `(数据, 错误)`。
    -   **如果返回错误**: 它会根据错误类型构造失败的JSON响应（如`{"errorCode": -19}`）。
    -   **如果返回成功**: 它会将返回的数据与`{"errorCode": 0}`合并，构造成功的JSON响应。
    -   **发送响应**: 将最终构造好的JSON发送给客户端。

## ⚙️ 快速开始

#### 1. 先决条件

-   [Go](https://golang.org/) (版本 >= 1.18)
-   一个正在运行的 [MySQL](https://www.mysql.com/) 数据库实例

#### 2. 安装与配置

1.  克隆本仓库到您的本地：
    ```bash
    git clone <your-repo-url>
    cd unity-lobby-server
    ```

2.  安装Go依赖项：
    ```bash
    go mod tidy
    ```

3.  配置服务器。复制或重命名`configs/config.json.example`为`configs/config.json`，并修改其中的内容，特别是数据库连接信息：
    ```json
    {
      "server": {
        "port": "8080"
      },
      "database": {
        "type": "mysql",
        "host": "127.0.0.1",
        "port": 3306,
        "user": "your_db_user",
        "password": "your_db_password",
        "name": "dmmtestdata"
      }
    }
    ```

#### 3. 运行服务器

直接在项目根目录运行`main.go`即可启动服务器：
```bash
go run main.go
```
您将在控制台看到类似以下的启动日志：
```
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.
...
[GIN-debug] POST   /                         --> unity-lobby-server/internal/server.dispatchHandler (3 handlers)
...
[GIN-debug] Listening and serving HTTP on :8080
```

## 📝 如何添加新的业务逻辑 (例如 `msg_id=30009`)

得益于模块化的设计，添加新的业务接口变得极其简单。

**第1步：在`internal/handler/`目录下新建一个文件 `30009.go`。**

**第2步：在该文件中编写您的业务逻辑，并使用`init()`函数进行自注册。**

```go
// internal/handler/30009.go
package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"unity-lobby-server/internal/game_error"
)

// init函数会在服务器启动时自动执行
func init() {
	// 将我们的处理函数注册到 msg_id "30009"
	Register("30009", handle30009)
}

// handle30009 是处理购买物品的纯业务逻辑函数
func handle30009(c *gin.Context, msgData map[string]interface{}) (map[string]interface{}, error) {
	// 从 msgData 中获取客户端传来的参数
	itemID, ok := msgData["itemID"].(float64) // JSON数字默认解析为float64
	if !ok {
		// 如果参数不正确，返回一个预定义的错误
		return nil, game_error.New(-13) // 非法参数
	}

	log.Printf("Player is trying to buy item: %d", int(itemID))
	
	// --- 在这里编写您的购买逻辑 ---
	// 1. 检查玩家金币是否足够
	// 2. 从数据库中扣除金币
	// 3. 将物品添加到玩家的资产中
	// 4. ...

	// 业务处理成功，返回需要额外发给客户端的数据
	resultData := map[string]interface{}{
		"newItemCount": 5, // 假设玩家现在有5个这个物品
		"goldRemained": 9500,  // 剩余金币
	}

	return resultData, nil
}
```

**第3步：重新启动服务器。**

完成！您不需要修改任何其他文件。服务器现在已经可以处理`msg_id=30009`的请求了。

## 🎨 贡献与未来方向

欢迎对本项目进行贡献。一些可以探索的未来方向包括：
-   实现配置文件的热加载。
-   引入Redis来管理更实时的在线状态和会话数据。
-   编写更全面的单元测试和集成测试。
-   优化数据库模型，将大型JSON字段根据查询需求拆分为更小的表。

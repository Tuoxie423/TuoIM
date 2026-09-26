# GinChat

基于 Gin + WebSocket + MySQL + Redis 实现的即时通讯（IM）系统，支持单聊、群聊、好友管理、群组管理、文件上传。

## 技术栈

| 层 | 技术 |
|---|---|
| Web 框架 | Gin |
| 实时通信 | gorilla/websocket + UDP |
| ORM | GORM (MySQL) |
| 缓存/消息 | go-redis |
| 配置 | Viper |
| 文件上传 | 本地存储 / 阿里云 OSS |

## 目录结构

```
ginchat/
├── main.go                 # 入口：初始化 + 启动服务
├── config.yml              # 配置文件（数据库、Redis、OSS、心跳、端口、日志）
├── config/                 # 配置层：结构体 + 全局实例 + 读配置
│   └── config.go
├── initial/                # 初始化层：建立连接（MySQL/Redis/日志）
│   ├── mysql.go
│   ├── redis.go
│   └── logger.go
├── router/                 # 路由注册
│   └── router.go
├── handler/                # HTTP 接入层：参数解析、调 service、返回响应
│   ├── user.go
│   ├── contact.go
│   ├── community.go
│   └── attach.go
├── service/                # 业务逻辑层：加好友校验、建群事务、登录鉴权等
│   ├── user.go
│   ├── contact.go
│   └── community.go
├── model/                  # 数据模型层：struct + TableName + 基础查询
│   ├── user.go
│   ├── contact.go
│   ├── community.go
│   └── message.go
├── ws/                     # WebSocket 接入层：长连接管理（与 HTTP 隔离）
│   ├── server.go           #   握手升级 + 连接注册
│   ├── node.go             #   连接对象 + 心跳 + 超时清理
│   ├── dispatch.go         #   消息分发（私聊/群聊路由）
│   └── udp.go              #   UDP 广播
├── pkg/                    # 可复用工具（无业务逻辑）
│   ├── resp/               #   统一响应封装
│   └── util/               #   md5、随机数等
├── views/                  # HTML 模板
├── asset/                  # 前端静态资源
└── docs/                   # 接口文档（Swagger）
```

## 分层原则

依赖方向是**单向**的，上层可以调用下层，下层禁止反向依赖上层：

```
handler  ──▶  service  ──▶  model
   │            │
   └────────────┴──▶  ws（独立，只管长连接）
```

| 层 | 职责 | 不应该做 |
|---|---|---|
| `handler` | 解析参数、调用 service、组装响应 | 写业务逻辑、直接操作数据库 |
| `service` | 业务规则、事务、鉴权校验 | 直接碰 HTTP 上下文 |
| `model` | 数据结构定义、基础增删改查 | 写业务判断、管连接 |
| `ws` | 长连接、心跳、消息路由 | 处理好友/群等业务 |
| `config` | 配置加载、持有全局实例 | 写业务 |
| `initial` | 建立 DB/Redis/日志连接 | 写业务 |

## 关键设计点

- **配置与连接分离**：`config` 只管「读配置」，`initial` 只管「建连接」，运行时连接挂在 `config.Global` 上（`config.Global.DB`、`config.Global.Red`）。
- **WebSocket 独立成 `ws` 包**：长连接和 HTTP 请求/响应是两种完全不同的东西，物理隔离，避免和业务搅在一起。
- **统一响应**：所有接口用 `pkg/resp` 统一返回格式，字段用 `Code`/`Msg`/`Data`。
- **消息流转**：客户端 → `ws` 接入 → `dispatch` 分发 → 目标连接；历史消息走 Redis ZSet。

## 快速开始

```bash
# 1. 建库建表（sql/init_ginchat.sql）
# 2. 修改 config.yml 的数据库/Redis 地址
# 3. 运行
go build -o ginchat.exe .
./ginchat.exe
```

默认监听 `:8082`，Swagger 文档见 `/swagger/index.html`。

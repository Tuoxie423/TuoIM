#  TuoIM —— 高并发即时通讯后端

<div align="center">

一个基于 **Gin + WebSocket + MySQL + Redis** 的即时通讯（IM）后端，支持单聊、群聊、好友管理、实时消息推送。

**⚡ 核心亮点：单体架构下的高并发抗压设计**

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-v1.12-009688?style=flat-square&logo=go&logoColor=white)
![WebSocket](https://img.shields.io/badge/WebSocket-gorilla%2Fwebsocket-8A2BE2?style=flat-square)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat-square&logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-5.0%2B-DC382D?style=flat-square&logo=redis&logoColor=white)
![JWT](https://img.shields.io/badge/Auth-JWT-000000?style=flat-square&logo=jsonwebtokens&logoColor=white)
![架构](https://img.shields.io/badge/架构-单体-ff69b4?style=flat-square)

**前端仓库**　[![GitHub](https://img.shields.io/badge/GitHub-TuoIM--Frontend-181717?style=flat-square&logo=github&logoColor=white)](https://github.com/Tuoxie423/TuoIM-Frontend)　[![Gitee](https://img.shields.io/badge/Gitee-tuo--im--frontend-C71D23?style=flat-square&logo=gitee&logoColor=white)](https://gitee.com/tuoxie423/tuo-im-frontend)

</div>

---

## 📌 项目定位

这是一个 **单体项目**：**没有引入 MQ 中间件、没有微服务拆分**。消息的「削峰、解耦、攒批、可靠投递」全部靠**进程内异步管道 + Redis 做共享状态**来实现，保持架构简单可控的同时，扛住高频消息写入压力。

> 核心思路就是：**把慢操作（落库）从请求主链路剥离，用异步化 + 攒批 + 缓存 + 背压换取吞吐。**

---

## ✨ 功能特性

| 模块 | 能力 |
|---|---|
| 👤 用户 | 注册、登录、JWT 鉴权、资料修改、头像上传、按手机号搜索 |
| 🤝 好友 | 申请、同意、拒绝、删除、好友列表（含在线状态）、未读申请数 |
| 💬 单聊 | 文本 / 图片消息、历史消息、实时推送、游标分页 |
| 👥 群聊 | 建群、发消息、加人、踢人、退群、群信息修改、群成员/群列表 |
| 📡 实时通信 | WebSocket 推送、应用层心跳、离线消息补推、多端在线 |
| 📁 文件 | 图片上传（本地磁盘存储，返回可访问 URL） |

---

## ⚡ 高并发抗压设计（核心）

核心就是把消息链路做成一条**「异步流水线」**，并用 Redis 承担高频读、MySQL 只做落库。下面逐层拆解。

### 整体数据流

```
HTTP 请求（/chat/send）── 校验好友/群成员 ──> msgCh ── 立即返回，请求不等待落库
                                                    │
                         ┌──────────────────────────┼──────────────────────────┐
                         ▼                          ▼                          ▼
                   🚀 dispatcher             💾 persister               🔁 retrier
                   分发器（推送）             落库器（攒批）             重试器（死信）
                         │                          │                          │
              在线 → WebSocket 推          攒 100 条/100ms       每 5s 拉死信队列重试
              离线 → Redis 离线队列         CreateInBatches 批量写   成功 → LREM 移除
```

三个 goroutine 各司其职、互不阻塞，形成一条 **「生产者-消费者」** 流水线。

### ① 异步消息管道（削峰解耦）

发消息**不阻塞 HTTP 请求**——消息进 channel 后立即返回，真正耗时的「推送 + 落库」交给后台 goroutine：

```go
msgCh     = make(chan messageTask, 1024)   // 消息主管道
persistCh = make(chan models.Message, 1024) // 落库管道
wg.Add(3)
go dispatcher()  // 分发：在线推 WS / 离线存 Redis，再丢进落库管道
go persister()   // 落库：攒批批量写 MySQL
go retrier()     // 重试：定时重放死信队列
```

**效果**：请求主链路只剩「校验 + 入管道」两个快操作，落库延迟与请求延迟解耦。

### ② 攒批落库（降低 DB 压力）

消息**不逐条 INSERT**，而是攒批批量写入，把 N 次往返合并成 1 次：

```go
if len(batch) >= 100 {          // 攒满 100 条，立即刷
    flushBatch(batch)
} else if 到点 100ms {          // 没攒满但超时，也刷（低延迟兜底）
    flushBatch(batch)
}

config.Global.DB.CreateInBatches(batch, 100)  // 一次写入
```

**效果**：批量插入吞吐比逐条高一个数量级，大幅降低 MySQL 的 QPS 与锁竞争。

### ③ 背压机制（宁可慢，不可丢）

消息管道是**阻塞发送**，channel 满时让上游请求**等待**，而不是悄悄丢消息：

```go
msgCh <- task   // 阻塞：满了就等，保证消息一条不丢
```

与「异步日志」形成对比——日志允许丢弃（见 ⑧），**消息不允许**，两者的取舍是有意为之。

### ④ 死信队列（可靠保障）

落库失败的消息**不丢弃**，进 Redis 死信队列，由 `retrier` 定时重试：

```
落库失败 → RPUSH fail_msg → retrier 每 5s LRANGE 拉取 → 逐条重试
                             成功 → LREM 删除 / 失败 → 留着下次再试
```

**效果**：即使 MySQL 短暂抖动，消息也只是「延迟送达」，不会永久丢失。

### ⑤ Redis 缓存分层（减轻 MySQL 压力）

高频读操作全部走 Redis，MySQL 只做权威存储和兜底：

| 用途 | Key | 类型 | 说明 |
|---|---|---|---|
| 在线状态 | `online:{uid}` | string | 高频查询，TTL 4h，避免打 DB |
| 好友关系 | `friend:{uid}` | set | 发消息校验好友，`SIsMember` O(1) 查 |
| 离线消息 | `offline:{uid}` | list | 不在线暂存，上线 `LRANGE` + `DEL` 补推 |
| 死信队列 | `fail_msg` | list | 落库失败的消息，定时重试 |

好友校验采用 **cache-aside** 模式：先查 Redis，miss 则回查 MySQL 并回填缓存，Redis 重启后能自愈。

### ⑥ WebSocket 高并发（goroutine 天然优势）

每个连接**读写分离两个 goroutine**，用「写缓冲 channel + 互斥锁」避免并发写冲突：

```go
type Client struct {
    Send chan []byte        // 写缓冲（256），writePump 独占消费，避免并发写
    Done chan struct{}      // 连接关闭信号
}

type Hub struct {
    clients map[int64][]*Client  // userID → 连接列表，支持一个用户多端在线
    mu      sync.RWMutex
}
```

- **读写分离**：`readPump` 独占读 / `writePump` 只写，单连接无并发写竞争。
- **多端在线**：Hub 用 `map[uid][]*Client` 管理，Push 时遍历该用户所有连接。
- **非阻塞推送**：`select { case c.Send <- data: case <-c.Done: }`，连接已死就跳过，不拖累其他连接。
- **应用层心跳**：读超时 60s，收到任何消息（含心跳）即续命，超时判死并清理 Hub / 在线状态。

### ⑦ 游标分页

历史消息用**游标（id）分页**，不用 `OFFSET`，避免大数据量下深分页的性能崩塌：

```sql
WHERE room_id = ? AND id < ? ORDER BY id DESC LIMIT 20
```

多查一条判断 `has_more`，前端据此决定是否还能上拉加载。

### ⑧ 异步日志（不拖累业务）

日志走**独立异步通道**，绝不阻塞业务请求：

```go
ch: make(chan logEntry, 1024)   // 缓冲 channel
out: bufio.NewWriter(w)          // 缓冲写，攒批刷盘
```

- **非阻塞投递**：channel 满时 `select default` **丢弃并计数**，日志丢了可以，请求慢了不行。
- **定时刷盘**：每 3s 强制 `Flush`，低流量时日志不会一直堆在内存缓冲。
- **结构化 JSON**：字段顺序稳定（用 struct 而非 map），便于日志系统索引。

### ⑨ 连接池 & 端口隔离

| 配置 | 值 | 作用 |
|---|---|---|
| MySQL `max_con` | 100 | 限制 DB 连接数，防止打爆数据库 |
| Redis `poolSize` / `minIdleConn` | 30 / 30 | 复用连接，预热空闲连接降低冷启动延迟 |
| HTTP / WebSocket | 8080 / 8081 | 两个端口**物理隔离**，WS 长连接与 HTTP 请求互不争抢 |

---

## 🧠 设计思路

| 问题 | 单体方案的答案 |
|---|---|
| 消息量突增怎么办？ | 异步管道削峰，请求不等待落库 |
| 落库慢怎么办？ | 攒批批量写，100 条/100ms 一次 |
| 消息不能丢怎么办？ | 背压 + 死信队列重试 |
| MySQL 扛不住高频读怎么办？ | Redis 缓存分层，读走缓存 |
| 长连接怎么写才不冲突？ | 读写分离 goroutine + 写缓冲 channel |
| 日志拖慢请求怎么办？ | 异步旁路日志，可丢弃可计数 |

> ⚠️ 适用边界：单体方案适合**中小规模**场景（单实例、单库）。当用户量/消息量达到需要**水平扩容**量级时，`msgCh` 进程内 channel 会成为单点，届时需引入真正的 MQ（Kafka/RabbitMQ）并把攒批落库拆成独立消费服务——这也是本项目刻意保留的「演进留白」。

---

## 🛠 技术栈

| 层 | 技术 | 说明 |
|---|---|---|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) | 路由 + 中间件 |
| 实时通信 | [gorilla/websocket](https://github.com/gorilla/websocket) | WebSocket 长连接 |
| ORM | [GORM](https://gorm.io/) + MySQL | 数据访问 + AutoMigrate 自动建表 |
| 缓存 / 在线状态 | [go-redis](https://github.com/redis/go-redis) | 在线状态、好友缓存、离线队列、死信队列 |
| 鉴权 | [golang-jwt/v5](https://github.com/golang-jwt/jwt) | 无状态 JWT 鉴权 |
| 密码加密 | bcrypt | 不可逆哈希存储 |
| 配置 | [Viper](https://github.com/spf13/viper) | YAML 配置加载 |
| 文档 | [Swagger](https://github.com/swaggo/swag) | 接口文档自动生成 |
| 文件存储 | 本地磁盘 | 图片存 `upload/`，静态托管 |

---

## 📂 目录结构

```
.
├── main.go             # 入口：初始化 + 启动消息队列 + WebSocket + HTTP 服务
├── config.yml          # 配置（含密钥，已被 .gitignore 忽略）
├── config.example.yml  # 配置样例
├── config/             # 配置结构 + Viper 加载
├── initial/            # 初始化：MySQL / Redis / 日志
├── router/             # 路由注册
├── api/                # HTTP 处理器（handler 层）
├── service/            # 业务逻辑层（含消息队列：dispatcher/persister/retrier）
├── models/             # 数据模型（GORM）
├── middleware/         # 中间件：JWT / CORS / 异步日志
├── cache/              # Redis 缓存封装（在线/好友/离线/死信）
├── ws/                 # WebSocket（Hub / Client / 读写分离）
├── utils/              # 工具：JWT / bcrypt / 响应 / 错误码
├── docs/               # Swagger 文档
└── upload/             # 上传的图片（静态托管）
```

---

## 🏗 架构设计

### 分层

```
api（HTTP 处理器） → service（业务逻辑） → models（数据模型）
                          ↓
                    cache（Redis 缓存）
                    ws（WebSocket 推送）
```

### 数据模型

| 表 | 作用 | 关键设计 |
|---|---|---|
| `user_basic` | 用户 | 手机号登录，密码 bcrypt |
| `room` | 会话（单聊/群聊统一抽象） | `type` 1 群聊 2 单聊 |
| `room_friend` | 单聊会话 | `room_key` 唯一索引（`uid1_uid2` 排序拼接）防重复建房 |
| `room_group` | 群聊详情 | 群名、群头像 |
| `group_member` | 群成员 | 角色：1 群主 / 2 管理员 / 3 成员 |
| `message` | 消息 | `room_id` 索引 + 自增 id 支撑游标分页 |
| `user_friend` | 好友关系 | **双向存储** A↔B 两条记录，联合唯一索引防重复 |
| `user_apply` | 好友申请 | 状态机 + 已读未读标记 |

### 核心设计点

- 🏠 **Room 抽象**：单聊、群聊统一成「会话」，消息只认 `room_id`，读写逻辑复用。
- ⏳ **懒建房间**：第一次聊天才建 room，加好友不建。
- 🔗 **好友双向存储**：A↔B 建两条记录 + 联合唯一索引，防重复也方便查「我的好友」。
- 🧯 **错误分层**：业务错误直接返回；缓存错误降级记日志不影响主流程；落库错误进死信队列重试。
- 🔐 **登录防枚举**：用户不存在与密码错误统一提示，避免手机号枚举。

---

## 📡 接口列表

### 公开接口（无需鉴权）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/register` | 用户注册 |
| POST | `/api/login` | 用户登录（返回 token） |
| GET | `/api/index` | 首页 |
| GET | `/api/swagger/*any` | Swagger 文档 |

### 用户（需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/users/info` | 当前用户信息 |
| PUT | `/api/users/info` | 更新资料（昵称/头像/性别） |
| GET | `/api/users/search` | 按手机号搜索用户 |

### 好友（需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/friend/apply` | 发起好友申请 |
| POST | `/api/friend/agree` | 同意申请 |
| POST | `/api/friend/reject` | 拒绝申请 |
| DELETE | `/api/friend/delete` | 删除好友 |
| GET | `/api/friend/applyList` | 申请列表 |
| GET | `/api/friend/list` | 好友列表（含在线状态） |
| GET | `/api/friend/unreadNum` | 未读申请数 |

### 消息 / 房间（需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/chat/send` | 发单聊消息 |
| GET | `/api/chat/history` | 历史消息（游标分页） |
| POST | `/api/room/get_or_create` | 获取/创建单聊房间 |

### 群聊（需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/group/create` | 建群 |
| POST | `/api/group/send` | 发群消息 |
| POST | `/api/group/add` | 加人 |
| POST | `/api/group/remove` | 踢人 |
| POST | `/api/group/quit` | 退群 |
| POST | `/api/group/update` | 更新群信息（群名/头像） |
| GET | `/api/group/members` | 群成员列表 |
| GET | `/api/group/list` | 我的群列表 |

### 文件（需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/file/upload` | 上传图片（存本地 `upload/`） |

### WebSocket

```
ws://localhost:8081/ws?token=<JWT>
```

WebSocket 无法携带 Header，token 通过 URL query 传递；应用层心跳：客户端定期发消息（如 `{"type":"heartbeat"}`），60s 无消息则判死下线。

---

## 🚀 快速开始

### 1. 环境要求

- Go 1.26+
- MySQL 8.0（自动建库建表）
- Redis 5.0+

### 2. 配置

```bash
cp config.example.yml config.yml
# 修改 config.yml 里的 MySQL / Redis / JWT 密钥
```

### 3. 启动

```bash
go run main.go
```

启动后自动建库建表，服务监听：

| 服务 | 端口 | 说明 |
|---|---|---|
| HTTP API | 8080 | Gin 服务 |
| WebSocket | 8081 | 独立端口，物理隔离 |

### 4. Swagger 文档

```bash
swag init -g main.go
# 访问 http://localhost:8080/api/swagger/index.html
```

---

## 🖥 前端

前端是独立的纯静态应用（HTML/CSS/JS），无构建依赖，与后端分仓库维护：

| 平台 | 仓库地址 |
|---|---|
| 🔗 GitHub | https://github.com/Tuoxie423/TuoIM-Frontend |
| 🔗 Gitee | https://gitee.com/tuoxie423/tuo-im-frontend |

本地开发时（前端代码目录）：

```bash
cd web && python serve.py   # no-cache 静态服务器
```

前端默认连接 `localhost:8080`（API）和 `localhost:8081`（WebSocket），部署时修改 `app.js` 顶部的地址。

---

## 🔭 后续规划（分布式演进）

本项目当前以单体形态交付，已经用「进程内异步管道」验证了**异步化 + 攒批 + 缓存 + 背压**这套高并发思路。后续有计划向**分布式架构**演进：

| 阶段 | 目标 | 主要改动 |
|---|---|---|
| 1️⃣ 消息中间件化 | 打破进程内单点 | 用 Kafka / RabbitMQ 替代 `msgCh`，消息队列独立成服务 |
| 2️⃣ 服务拆分 | 落库 / 推送可独立扩容 | 攒批落库拆成独立消费者，水平扩容 |
| 3️⃣ 水平扩容 | 支撑更大规模 | 多实例 + 网关负载均衡 + Redis Pub/Sub 跨实例推送 |

> 演进的关键在于：当前的核心设计（异步化、攒批、缓存分层、背压、死信队列）与分布式方案是**同构**的——迁移时只需把「进程内 channel」替换为「分布式消息队列」，业务代码改动最小。

---

<div align="center">

**TuoIM** · 单体即服务的 IM 高并发实践 🚀

</div>

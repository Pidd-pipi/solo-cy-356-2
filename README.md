# 城市共享菜园管理平台（communitygarden）

面向城市居民和农场主的社区农业平台：提供**菜园地块认养与 GIS 展示、种植计划与作物推荐、种植日记图文记录、收成预警与采摘提醒、农友社区交流**的完整全栈 Web 应用。

## 🚀 快速启动（Docker Compose，首选）

```bash
cd <项目目录>            # 支持任意目录名，包括中文目录名
docker compose up -d --build
```

启动完成后访问：

| 服务 | 地址 |
| --- | --- |
| 前端页面 | http://localhost:28516 |
| 后端健康检查 | http://localhost:29516/healthz |
| 后端 API | http://localhost:29516/api/v1 |
| API 文档（OpenAPI） | `backend/api/openapi.yaml` |

内置演示账号（密码 = 用户名 + `123`）：

| 用户名 | 密码 | 角色 |
| --- | --- | --- |
| `admin` | `admin123` | 管理员 |
| `farmer` | `farmer123` | 农场主 |
| `citizen` | `citizen123` | 城市居民 |

## ✨ 主要功能

1. **地块认养与 GIS 展示**：地图展示地块分布，标注空闲/已认养/待释放状态，展示面积、土壤类型、日照条件，在线认养。
2. **种植计划与作物推荐**：认养后制定种植计划，按季节推荐适宜作物，生成预期收获时间线（蔬菜 45 天/水果 90 天/香草 35 天）。
3. **种植日记图文记录**：按播种/浇水/施肥/除虫/收成记录种植过程，支持点赞与评论。
4. **收成预警与采摘提醒**：近 7 天成熟作物自动提醒，记录采摘重量与品质，生成年度收成统计报表。
5. **农友社区交流**：种植经验 / 病虫害防治 / 食谱创意 / 线下农耕活动四类帖子，实时动态 WebSocket 广播。

## 🛠 技术栈

| 层 | 技术栈 |
| --- | --- |
| 前端 | Vue 3 + TypeScript |
| UI 库 | Element Plus |
| 构建工具 | Vite |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 16 |
| 缓存/限流 | Redis 7 |
| 实时通信 | gorilla/websocket |
| 认证 | JWT + RBAC |
| 日志 | log/slog 结构化日志 |
| 参数校验 | go-playground/validator/v10 |

## 📁 项目目录结构

```
backend/
├── cmd/server/main.go          # 入口：加载配置、装配依赖、启动服务
├── internal/
│   ├── config/config.go        # 环境变量配置
│   ├── database/database.go    # 连接 + AutoMigrate + 种子数据
│   ├── model/                  # 每个实体一个 model 文件
│   ├── dto/                    # 每个实体一个 DTO 文件
│   ├── repository/             # 每个实体一个 repository 文件
│   ├── service/                # 每个实体一个 service 文件（事务/状态机）
│   ├── handler/                # 每个实体一个 handler 文件
│   ├── router/                 # 每个实体一个路由注册文件
│   ├── middleware/             # auth / rbac / request_id / error_handler / rate_limit / audit / request_log / cors
│   ├── constants/              # enums / error_codes / log_templates / messages
│   ├── util/                   # logger / jwt / password / response / pagination / formatters / app_error / context
│   └── ws/hub.go               # WebSocket 社区实时广播
├── migrations/                 # 迁移脚本
├── api/openapi.yaml            # OpenAPI 接口文档
├── Dockerfile                  # Go 多阶段构建
├── go.mod / go.sum
frontend/
├── src/
│   ├── api/                    # 每个实体一个 API 文件
│   ├── components/             # StatusBadge / EmptyState / DataTable / ConfirmDialog
│   ├── pages/                  # 每个模块一个页面
│   ├── stores/                 # 按实体拆分 Pinia store
│   ├── hooks/                  # useAuth / usePagination
│   ├── utils/                  # request.ts（拦截器）/ format.ts
│   ├── constants/              # 与后端对应的枚举
│   └── layouts/MainLayout.vue
├── Dockerfile                  # Node 构建 + Nginx 托管
└── nginx.conf                  # 路由回退 + /api 反向代理 + WebSocket
database/
└── init.sql                    # 数据库初始化基线脚本（幂等）
docker-compose.yml
.env / .env.example
README.md
```

**严禁合并职责到单一文件**：不允许把多个实体的 model/repository/service/handler 写进同一个文件，也不允许前端把所有页面写在 App.vue / App.tsx 中。

## 🔑 核心实体（≥4 个，贯穿全栈）

| 实体 | 模型/表 | 后端分层文件 | 前端 api / store / 页面 |
| --- | --- | --- | --- |
| 用户 User | `users` | `model/user.go`、`repository/user_repository.go`、`service/user_service.go`、`handler/user_handler.go`、`router/user.go` | `api/auth.ts`、`stores/auth.ts`、`pages/Login.vue`、`pages/Register.vue` |
| 地块 Plot | `plots` | `model/plot.go`、`repository/plot_repository.go`、`service/plot_service.go`、`handler/plot_handler.go`、`router/plot.go` | `api/plot.ts`、`stores/plot.ts`、`pages/PlotMap.vue` |
| 种植计划 PlantingPlan | `planting_plans` | `model/planting_plan.go`、`repository/planting_plan_repository.go`、`service/planting_plan_service.go`、`handler/planting_plan_handler.go`、`router/planting_plan.go` | `api/plantingPlan.ts`、`stores/plantingPlan.ts`、`pages/PlantingPlan.vue` |
| 收成记录 HarvestRecord | `harvest_records` | `model/harvest_record.go`、`repository/harvest_record_repository.go`、`service/harvest_record_service.go`、`handler/harvest_handler.go`、`router/harvest.go` | `api/harvest.ts`、`pages/Harvest.vue` |
| 种植日记 DiaryEntry | `diary_entries` / `diary_comments` | `model/diary_entry.go`、`repository/diary_entry_repository.go`、`service/diary_service.go`、`handler/diary_handler.go`、`router/diary.go` | `api/diary.ts`、`stores/diary.ts`、`pages/Diary.vue` |
| 社区帖子 CommunityPost | `community_posts` / `community_comments` | `model/community_post.go`、`repository/community_post_repository.go`、`service/community_service.go`、`handler/community_handler.go`、`router/community.go` | `api/community.ts`、`stores/community.ts`、`pages/Community.vue` |

> 审计日志 AuditLog 作为横切关注点实体：`model/audit_log.go` → `repository/audit_repository.go` → `service/audit_service.go` → `handler/audit_handler.go` → `router/audit.go` → 前端 `api/audit.ts`、`pages/Audit.vue`。

## 🧭 横切关注点（触达文件层）

1. **JWT 认证 + RBAC 权限**：数据库 `users.role` 角色字段 → `internal/middleware/auth.go`、`internal/middleware/rbac.go`、`internal/util/jwt.go` → 前端 `utils/request.ts`（自动携带 token、401 跳转）、`router/index.ts`（路由守卫）、`hooks/useAuth.ts`（按钮显隐）。
2. **操作审计日志**：数据库 `audit_logs` 表 → `internal/middleware/audit.go`（写操作自动落库）→ service/handler 埋点（认养/释放/收成/角色变更）→ 前端 `pages/Audit.vue`（管理员查看）。
3. **全局错误处理与请求追踪**：`internal/middleware/request_id.go`、`internal/middleware/error_handler.go`、`internal/util/app_error.go`、`internal/constants/error_codes.go` → 前端 `utils/request.ts` 拦截器统一提示。

## 🧩 共享枚举/常量（后端 `internal/constants/enums.go`，前端 `src/constants/index.ts`）

| 枚举 | 取值 | 后端出现位置 |
| --- | --- | --- |
| RoleType 角色 | admin / farmer / citizen | `constants/enums.go`、`model/user.go`、`dto/user_dto.go`、`service/user_service.go`（ChangeRole 校验）、`middleware/rbac.go`、`middleware/audit.go`、`handler/planting_plan_handler.go`、`handler/harvest_handler.go`、`handler/diary_handler.go`、`util/formatters.go`、`log_templates.go`、`error_codes.go`、`database/database.go`（种子数据） |
| PlotStatus 地块状态 | available / adopted / harvested | `constants/enums.go`、`model/plot.go`、`dto/plot_dto.go`、`service/plot_service.go`（认养/释放状态机）、`repository/plot_repository.go`（过滤）、`util/formatters.go`、`log_templates.go`、`database/database.go`（种子数据）、`api/openapi.yaml` |
| PlanStatus 种植计划状态 | planned / planting / growing / harvesting / completed | `constants/enums.go`、`model/planting_plan.go`、`dto/planting_plan_dto.go`（oneof 校验）、`service/planting_plan_service.go`（PlanStatusTransitions 状态机）、`handler/planting_plan_handler.go`、`util/formatters.go`、`log_templates.go`、`error_codes.go`（CodePlanStateNotAllowed）、`database/database.go`（种子数据）、前端 `constants/index.ts`（PlanStatusMeta / PlanStatusNext 按钮显隐） |
| CropType 作物类型 | vegetable / fruit / herb | `constants/enums.go`、`model/planting_plan.go`、`dto/planting_plan_dto.go`、`service/planting_plan_service.go`（成熟时间估算）、`util/formatters.go`、`repository/harvest_record_repository.go`（分组统计）、`database/database.go` |
| Season 季节 | spring / summer / autumn / winter | `constants/enums.go`、`dto/planting_plan_dto.go`、`service/planting_plan_service.go`（SeasonCrops 推荐表）、`util/formatters.go`、`database/database.go`、前端 `pages/PlantingPlan.vue`、`pages/Dashboard.vue` |
| DiaryAction 日记动作 | sowing / watering / fertilizing / pest_control / harvest / other | `constants/enums.go`、`model/diary_entry.go`、`dto/diary_entry_dto.go`、`util/formatters.go`、`database/database.go`、`log_templates.go` |
| HarvestQuality 收成品质 | excellent / good / fair | `constants/enums.go`、`model/harvest_record.go`、`dto/harvest_record_dto.go`、`util/formatters.go`、`repository/harvest_record_repository.go`（分组统计）、`database/database.go` |
| PostType 帖子类型 | experience / pest / recipe / activity | `constants/enums.go`、`model/community_post.go`、`dto/community_post_dto.go`、`repository/community_repository.go`（过滤/统计）、`util/formatters.go`、`log_templates.go`、`database/database.go`、`api/openapi.yaml` |

> 状态机必须跨多处定义：核心状态流转规则在 `service/planting_plan_service.go` 状态机、前端 `constants/index.ts` 按钮显隐、`util/formatters.go`、`log_templates.go`、`error_codes.go` 中同时存在。新增一个状态值需要修改至少 10 处文件（牵一发动全身）。

## 🌐 API 清单（统一前缀 `/api/v1`，分页参数 `page` / `page_size`，响应 `{code, message, data}`）

### 认证
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| POST | `/auth/register` | 注册（默认 citizen） | 公开 |
| POST | `/auth/login` | 登录获取 JWT | 公开 |

### 用户
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/me` | 当前用户信息 | 登录 |
| PUT | `/me` | 更新当前用户资料 | 登录 |
| GET | `/users` | 用户列表 | 管理员 |
| GET | `/users/:id` | 用户详情 | 登录 |
| PUT | `/users/:id/role` | 变更角色 | 管理员 |
| PUT | `/users/:id/status` | 启用/禁用 | 管理员 |

### 地块
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/plots` | 地块列表（`?status=`） | 公开 |
| GET | `/plots/:id` | 地块详情 | 公开 |
| POST | `/plots` | 创建地块 | 管理员 |
| PUT | `/plots/:id` | 更新地块 | 管理员 |
| POST | `/plots/:id/adopt` | 认养地块（事务 + FOR UPDATE） | 登录 |
| POST | `/plots/:id/release` | 释放地块 | 认养人/管理员 |

### 种植计划
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/planting-plans` | 计划列表 | 登录 |
| POST | `/planting-plans` | 创建计划（季节校验 + 收获时间线） | 登录 |
| GET | `/planting-plans/:id` | 计划详情 | 登录 |
| PUT | `/planting-plans/:id` | 更新计划（仅 planned） | 登录 |
| POST | `/planting-plans/:id/status` | 状态流转（状态机） | 登录 |
| GET | `/crops/recommendations?season=` | 季节作物推荐 | 登录 |
| GET | `/reminders/harvest` | 采摘提醒（近 7 天成熟） | 登录 |

### 收成记录
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/harvests` | 收成列表 | 登录 |
| POST | `/harvests` | 记录收成（成熟校验） | 登录 |
| PUT | `/harvests/:id` | 更新收成 | 本人/管理员 |
| DELETE | `/harvests/:id` | 删除收成 | 本人/管理员 |
| GET | `/stats/annual?year=` | 年度收成统计 | 登录 |

### 种植日记
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/diaries` | 日记列表 | 登录 |
| POST | `/diaries` | 发布日记 | 登录 |
| GET | `/diaries/:id` | 日记详情（含评论） | 登录 |
| PUT | `/diaries/:id` | 更新日记 | 本人/管理员 |
| DELETE | `/diaries/:id` | 删除日记 | 本人/管理员 |
| POST | `/diaries/:id/like` | 点赞 | 登录 |
| POST | `/diaries/:id/comments` | 评论 | 登录 |

### 农友社区
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/community-posts` | 帖子列表（`?post_type=`） | 登录 |
| POST | `/community-posts` | 发布帖子 | 登录 |
| GET | `/community-posts/:id` | 帖子详情（含评论） | 登录 |
| PUT | `/community-posts/:id` | 更新帖子 | 本人/管理员 |
| DELETE | `/community-posts/:id` | 删除帖子（软删除） | 本人/管理员 |
| POST | `/community-posts/:id/like` | 点赞 | 登录 |
| POST | `/community-posts/:id/comments` | 评论 | 登录 |

### 其他
| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/dashboard/stats` | 平台概览统计 | 登录 |
| GET | `/audit-logs` | 审计日志列表 | 管理员 |
| GET | `/ws/community?token=` | 社区实时消息 WebSocket | 登录 |

### 接口复用说明
- `GET /stats/annual`（收成统计接口）与 `GET /planting-plans/stats`（种植计划统计）**复用同一个 service 方法** `HarvestRecordService.AnnualStats`。
- `GET /plots/:id`（地块详情接口）与创建种植计划 `POST /planting-plans` **复用同一个 service 方法** `PlotService.GetByID`。
- 分页 `util.Paginate` 被全部 repository 复用；`ListByUser` 系列仓储方法被计划/收成列表接口复用。

## 🔌 API 调用示例（curl，含 JWT 请求头）

```bash
# 1. 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:29516/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')

# 2. 健康检查
curl -s http://localhost:29516/healthz

# 3. 地块列表
curl -s http://localhost:29516/api/v1/plots

# 4. 认养地块（登录用户）
curl -s -X POST http://localhost:29516/api/v1/plots/1/adopt \
  -H "Authorization: Bearer $TOKEN"

# 5. 创建种植计划
curl -s -X POST http://localhost:29516/api/v1/planting-plans \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"plot_id":1,"crop_name":"菠菜","crop_type":"vegetable","season":"spring"}'

# 6. 季节作物推荐
curl -s "http://localhost:29516/api/v1/crops/recommendations?season=summer" \
  -H "Authorization: Bearer $TOKEN"

# 7. 年度收成统计
curl -s "http://localhost:29516/api/v1/stats/annual?year=2026" \
  -H "Authorization: Bearer $TOKEN"

# 8. 审计日志（管理员）
curl -s "http://localhost:29516/api/v1/audit-logs" \
  -H "Authorization: Bearer $TOKEN"
```

## 🐳 Docker 部署说明

- **端口映射**：前端 `28516:80`、后端 `29516:8080`、数据库 `44001:5432`、Redis `46301:6379`（容器内固定端口不变，宿主机端口可通过 `.env` 修改）。
- **数据卷**：`db_data`（PostgreSQL）、`redis_data`（Redis）使用命名卷持久化，删除容器数据不丢失；执行 `docker compose down -v` 会清空数据卷。
- **常见问题**：
  - 后端未启动：`docker compose logs backend` 查看数据库连接是否成功。
  - 端口冲突：修改 `.env` 中 `FRONTEND_PORT` / `BACKEND_PORT` / `DB_PORT` / `REDIS_PORT` 后重新 `docker compose up -d`。
  - 中文目录名：本项目所有路径均在容器内固定为 `/app`，与宿主机目录名无关，可在任意目录名（含中文）下启动。

## 💻 本地开发（备选）

```bash
# 后端
cd backend
go mod tidy
go run ./cmd/server        # 或 go build ./...

# 前端
cd frontend
npm install
npm run dev                # http://localhost:5173，/api 代理到 29516
```

## 🧪 测试与质量

```bash
cd backend
go build ./...
go vet ./...
go test ./...              # service 与 repository 表驱动单元测试
```

## 📄 License

MIT License

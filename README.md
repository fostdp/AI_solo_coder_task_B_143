# 🏹 古代连弩动力学仿真与射速优化系统

基于多刚体动力学仿真与强化学习的古代连弩射速优化系统。通过传感器数据采集、多刚体动力学计算、强化学习优化，实现连弩射速与弓弦寿命的最佳平衡。

---

## 📐 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              前端 (React + Three.js)                         │
│  ┌─────────────────────┐  ┌────────────────────┐  ┌────────────────────┐ │
│  │ repeating_crossbow  │  │  fire_rate_panel   │  │  WebSocket Client  │ │
│  │  _3d.js (3D渲染)    │  │  (控制面板)        │  │  (实时数据)        │ │
│  └─────────┬───────────┘  └─────────┬──────────┘  └─────────┬──────────┘ │
└────────────┼────────────────────────┼────────────────────────┼────────────┘
             │                        │                        │
             │  HTTP REST API        │  WebSocket             │
             ▼                        ▼                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Go 后端服务 (模块化架构)                               │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐    │
│  │                         Coordinator (协调器)                          │    │
│  │                  模块生命周期 / Channel路由 / 数据分发                │    │
│  └───┬──────────┬───────────────┬──────────────────┬───────────────────┘    │
│      │          │               │                  │                        │
│      ▼          ▼               ▼                  ▼                        │
│  ┌───────┐ ┌──────────┐ ┌──────────────┐ ┌───────────────┐               │
│  │ DTU   │ │Mechanism │ │ FireRate     │ │   Alarm WS    │               │
│  │Receiver│ │Simulator │ │  Optimizer   │ │  (告警+推送)   │               │
│  │(采集校验)│ │(动力学)  │ │  (强化学习)  │ │               │               │
│  └───┬───┘ └────┬─────┘ └──────┬───────┘ └───────┬───────┘               │
│      │         │               │                 │                       │
└──────┼─────────┼───────────────┼─────────────────┼───────────────────────┘
       │         │               │                 │
       ▼         ▼               ▼                 ▼
  ┌──────────┐ ┌──────────┐ ┌──────────┐    ┌──────────┐
  │Timescale │ │  Redis   │ │Prometheus│    │  MQTT    │
  │   DB     │ │ (缓存)   │ │ (指标)    │    │ Broker   │
  │ (时序数据)│ └──────────┘ └──────────┘    └──────────┘
  └──────────┘
       ▲
       │ 传感器数据
  ┌───────────┐
  │  传感器   │
  │  模拟器   │─── MQTT ───► Mosquitto
  └───────────┘
```

### 后端模块架构

```
                        ┌─────────────────────────┐
                        │      Coordinator        │
                        │  (Channel 通信枢纽)    │
                        └─────────────────────────┘
                                  ▲
                                  │
                   ┌──────────────┼──────────────┐
                   │              │              │
          ┌────────┴───┐   ┌─────┴──────┐  ┌───┴──────────┐
          │ dtu_receiver│   │ mechanism_ │  │ fire_rate_   │
          │  (采集校验) │   │ simulator  │  │  optimizer   │
          └────────────┘   │  (动力学)  │  │ (强化学习)   │
                           └────────────┘  └──────────────┘
                                  ▲
                                  │
                           ┌──────┴───────┐
                           │   alarm_ws   │
                           │  (告警+WS)   │
                           └──────────────┘
```

### 数据流向

```
传感器/模拟器 → dtu_receiver (校验) → Coordinator
                                            ├→ mechanism_simulator (仿真对比)
                                            ├→ fire_rate_optimizer (RL状态输入)
                                            └→ alarm_ws (告警检测 + WebSocket推送)
```

---

## 📁 目录结构

```
.
├── backend/                     # Go 后端服务
│   ├── cmd/server/              # 主程序入口
│   │   └── main.go
│   ├── config/                  # 配置文件
│   │   ├── config.go
│   │   ├── config.yaml
│   │   ├── mechanism_params.json    # 机构参数（外置）
│   │   └── rl_params.json           # RL参数（外置）
│   └── internal/
│       ├── api/                  # API层 (Gin)
│       ├── coordinator/          # 协调器 (Channel枢纽)
│       ├── dtu_receiver/         # 模块1：传感器采集校验
│       ├── mechanism_simulator/  # 模块2：多刚体动力学
│       ├── fire_rate_optimizer/  # 模块3：强化学习优化
│       ├── alarm_ws/             # 模块4：告警与WebSocket
│       ├── middleware/           # Prometheus指标中间件
│       ├── model/                # 数据模型
│       ├── repository/           # 数据访问层
│       ├── simulation/           # 底层：仿真算法库
│       └── rl/                   # 底层：强化学习算法库
│
├── frontend/                    # 前端 (React + Three.js)
│   ├── src/
│   │   └── components/
│   │       ├── repeating_crossbow_3d.js    # 3D连弩渲染
│   │       └── fire_rate_panel.js          # 射速控制面板
│   ├── Dockerfile
│   └── nginx.conf               # Nginx + Gzip配置
│
├── sensor-simulator/            # 连弩传感器模拟器
│   ├── Dockerfile
│   ├── simulator.py             # 主程序
│   ├── config.yaml              # 模拟器配置
│   └── requirements.txt
│
├── database/                    # 数据库
│   └── init.sql                 # TimescaleDB初始化 + 降采样配置
│
├── mqtt/                        # MQTT Broker配置
│   └── mosquitto.conf
│
├── monitoring/                  # 监控
│   └── prometheus/
│       ├── prometheus.yml
│       └── rules/
│           └── crossbow_alerts.yml
│
├── docker-compose.yml           # Docker Compose 编排
├── .env                         # 环境变量
└── README.md                    # 本文档
```

---

## 🚀 快速部署

### 前置要求

- Docker 20.10+
- Docker Compose v2+
- 至少 4GB 可用内存

### 一键启动（核心服务）

```bash
# 1. 克隆项目
git clone <repo-url>
cd crossbow-simulation

# 2. 复制环境变量（可选，默认已配置）
cp .env.example .env  # 如果有示例文件的话

# 3. 启动核心服务（数据库、后端、前端、Redis、MQTT）
docker-compose up -d

# 4. 查看服务状态
docker-compose ps
```

### 启动完整监控栈

```bash
# 启动Prometheus + Grafana
docker-compose up -d prometheus grafana
```

### 启动传感器模拟器

```bash
# 启动模拟器1号（正常模式）
docker-compose --profile simulator up -d sensor-simulator

# 启动模拟器2号（疲劳模式 + 小箭匣）
docker-compose --profile simulator2 up -d sensor-simulator-2

# 或同时启动两台
docker-compose --profile simulator --profile simulator2 up -d
```

### 停止服务

```bash
# 停止所有
docker-compose down

# 保留数据
docker-compose stop

# 停止并删除数据（危险！）
docker-compose down -v
```

---

## 🌐 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端 UI | http://localhost:3000 | 主界面，3D仿真 + 控制面板 |
| 后端 API | http://localhost:8080/api/v1 | RESTful API |
| WebSocket | ws://localhost:8080/ws/crossbow/{id} | 实时数据推送 |
| Prometheus | http://localhost:9090 | 指标采集与查询 |
| Grafana | http://localhost:3001 | 可视化面板（admin/admin） |
| pprof | http://localhost:6060/debug/pprof | Go性能分析 |
| MQTT Broker | localhost:1883 | 传感器数据MQTT上报 |
| MQTT WebSocket | localhost:9001 | MQTT over WebSocket |
| Redis | localhost:6379 | 缓存服务 |

---

## 🧪 传感器模拟器用法

### 命令行参数

```bash
python sensor-simulator/simulator.py [选项]
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--crossbow-id` | string | crossbow-001 | 连弩ID |
| `--backend-url` | string | http://localhost:8080/api/v1/sensor/data | 后端API地址 |
| `--report-interval` | float | 1 | 上报间隔（秒） |
| `--mode` | string | normal | 模拟模式 |
| `--config` | string | config.yaml | 配置文件路径 |
| `--magazine-size` | int | 10 | 箭匣容量（发） |
| `--initial-fatigue` | float | 0.0 | 初始弓弦疲劳值 0.0~1.0 |
| `--fatigue-rate` | float | - | 每发射疲劳增量（覆盖配置） |
| `--jam-after` | int | -1 | 发射N发后自动卡死 |
| `--disable-mqtt` | flag | false | 禁用MQTT上报 |

### 模拟模式

| 模式 | 说明 |
|------|------|
| `normal` | 正常工作模式，疲劳缓慢增长 |
| `fatigue` | 疲劳加速模式（3倍速率） |
| `accelerated` | 极快疲劳模式（10倍速率） |
| `jammed` | 直接进入卡弹状态 |
| `string_break` | 直接进入断弦状态 |

### 使用示例

```bash
# 1. 正常模式
python sensor-simulator/simulator.py

# 2. 小箭匣（5发）+ 高初始疲劳（0.5）
python sensor-simulator/simulator.py \
  --magazine-size 5 \
  --initial-fatigue 0.5

# 3. 加速疲劳测试（10倍速率）
python sensor-simulator/simulator.py --mode accelerated

# 4. 故障注入：50发后卡死
python sensor-simulator/simulator.py --jam-after 50

# 5. 调整上报频率（0.5秒一次，高频）
python sensor-simulator/simulator.py --report-interval 0.5

# 6. 只使用HTTP，禁用MQTT
python sensor-simulator/simulator.py --disable-mqtt

# 7. 指定不同的连弩ID（多台模拟）
python sensor-simulator/simulator.py --crossbow-id crossbow-003
```

### Docker 方式运行

```bash
# 用环境变量控制模拟器参数
docker run --rm \
  --network crossbow-crossbow-net \
  -v $(pwd)/sensor-simulator/config.yaml:/app/config.yaml:ro \
  crossbow-sensor-simulator:latest \
  --crossbow-id crossbow-001 \
  --mode normal \
  --magazine-size 10 \
  --initial-fatigue 0.0
```

### MQTT 订阅示例

```bash
# 订阅传感器数据
mosquitto_sub -h localhost -t "crossbow/crossbow-001/sensor" -v

# 使用MQTT客户端工具
mqtt sub -t "crossbow/+/sensor" -h localhost -p 1883
```

MQTT消息格式与HTTP上报的JSON格式一致。

---

## 📊 TimescaleDB 数据策略

### 降采样（连续聚合）

系统配置了三级连续聚合，自动从高频原始数据聚合出低精度数据：

| 聚合粒度 | 刷新间隔 | 保留时长 | 用途 |
|----------|----------|----------|------|
| 1分钟 | 每分钟 | 90天 | 实时监控、短期趋势 |
| 1小时 | 每15分钟 | 1年 | 日报、历史趋势 |
| 1天 | 每小时 | 永久 | 周报、月报、长期分析 |

### 数据保留策略

| 数据类型 | 保留时长 | 说明 |
|----------|----------|------|
| 原始传感器读数 | 30天 | 高频、高存储占用 |
| 高频动力学状态 | 7天 | 最占空间的原始数据 |
| 1分钟聚合 | 90天 | 用于短期趋势分析 |
| 1小时聚合 | 1年 | 用于日报和历史回顾 |
| 1天聚合 | 永久 | 用于长期研究分析 |

### 压缩策略

| 表 | 压缩延迟 | 压缩方式 |
|----|----------|----------|
| sensor_readings | 3天 | 列式压缩 + crossbow_id分段 |
| dynamics_states | 1天 | 列式压缩 + crossbow_id分段 |

压缩后存储空间可节省 80%-95%。

### 常用查询示例

```sql
-- 查询最近1小时原始数据（用于详细分析）
SELECT time, string_tension, fire_rate, string_fatigue
FROM sensor_readings
WHERE crossbow_id = '...'
  AND time > now() - INTERVAL '1 hour'
ORDER BY time DESC;

-- 查询最近24小时的小时级统计（用于趋势图）
SELECT bucket, avg_fire_rate, max_string_tension, max_fatigue
FROM sensor_readings_1h
WHERE crossbow_id = '...'
  AND bucket > now() - INTERVAL '24 hours'
ORDER BY bucket;

-- 查询最近7天的日统计（用于周报）
SELECT bucket, total_shots, final_fatigue, fatigue_delta
FROM sensor_readings_1d
WHERE crossbow_id = '...'
  AND bucket > now() - INTERVAL '7 days'
ORDER BY bucket;
```

---

## 📈 Prometheus 指标

### HTTP 指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `crossbow_http_requests_total` | Counter | HTTP请求总数（按method/path/status） |
| `crossbow_http_request_duration_seconds` | Histogram | HTTP请求耗时直方图 |

### 业务指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `crossbow_sensor_data_total` | Counter | 传感器数据接收总量 |
| `crossbow_alerts_total` | Counter | 告警产生数量（按级别） |
| `crossbow_simulations_running` | Gauge | 运行中的仿真实例数 |
| `crossbow_fire_rate` | Gauge | 当前射速（发/分钟） |
| `crossbow_string_tension_newtons` | Gauge | 当前弓弦张力（N） |
| `crossbow_string_fatigue_ratio` | Gauge | 当前弓弦疲劳比（0-1） |
| `crossbow_bow_arm_deformation_mm` | Gauge | 当前弩臂变形（mm） |
| `crossbow_ws_active_connections` | Gauge | 活跃WebSocket连接数 |

### 强化学习指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `crossbow_rl_training_active` | Gauge | RL训练是否活跃 |
| `crossbow_rl_episodes_total` | Counter | RL训练总回合数 |
| `crossbow_rl_reward` | Gauge | 最新RL奖励值 |
| `crossbow_rl_epsilon` | Gauge | 当前探索率ε |

### 内置告警规则

Prometheus 已配置 6 条告警规则：

| 告警名 | 触发条件 | 级别 |
|--------|----------|------|
| CrossbowServiceDown | 后端服务宕机 > 1分钟 | critical |
| HighStringTension | 张力 > 1150N 持续 30s | warning |
| HighFatigue | 疲劳 > 0.8 持续 1分钟 | warning |
| CriticalFatigue | 疲劳 > 0.9 持续 30s | critical |
| CriticalAlertGenerated | 产生critical级别告警 | critical |
| RLTakingTooLong | RL训练15分钟 < 5个episode | warning |

### pprof 性能分析

Go服务已启用 net/http/pprof：

```bash
# 查看所有profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存profile
go tool pprof http://localhost:6060/debug/pprof/heap

# goroutine分析
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 火焰图（需graphviz）
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## 🎯 前端优化

### Gzip 压缩

Nginx 已配置完整的 Gzip 压缩策略：

- 压缩级别：6（平衡压缩率与CPU消耗）
- 最小压缩文件：256字节
- 支持类型：文本、JS、CSS、JSON、XML、字体、SVG等
- gzip_vary：开启，支持CDN缓存

### 静态资源缓存

| 资源类型 | 缓存时长 | 策略 |
|----------|----------|------|
| 带hash的JS/CSS | 1年 | public + immutable |
| 图片/字体/SVG | 30天 | public + immutable |
| index.html | 不缓存 | 每次重新验证 |

### 安全响应头

- X-Frame-Options: SAMEORIGIN
- X-Content-Type-Options: nosniff
- X-XSS-Protection: 1; mode=block
- Referrer-Policy: strict-origin-when-cross-origin

---

## ⚙️ 配置文件

### 机构参数外置

机构动力学参数存储在 [backend/config/mechanism_params.json](backend/config/mechanism_params.json)：

- 弩臂参数（长度、刚度、质量、转动惯量）
- 弓弦参数（长度、刚度、阻尼、线密度）
- 凸轮参数（基圆半径、升程、转速、8段运动规律）
- 缓冲弹簧参数（刚度、预紧力、阻尼、最大压缩）
- 棘爪棘轮参数
- 箭矢参数（质量、直径、阻力系数）
- 箭匣参数（容量、弹簧刚度、摩擦）
- 设计规格（最高射速、设计寿命）
- 仿真参数（时间步长、积分器类型）

修改配置后重启服务生效：
```bash
docker-compose restart backend
```

### 强化学习参数外置

RL参数存储在 [backend/config/rl_params.json](backend/config/rl_params.json)：

- Agent参数（学习率、折扣因子、目标网络更新率）
- 训练参数（总回合、每回合步数、回放池大小）
- 预训练参数（是否启用、演示回合数、预训练轮数）
- 专家策略参数（启发式规则权重）
- 探索参数（ε初始/最终/衰减率）
- 日志参数（记录间隔、保存间隔）

---

## 🛠️ 开发与调试

### 后端开发

```bash
cd backend

# 安装依赖
go mod download

# 本地运行
go run cmd/server/main.go

# 构建静态二进制
CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-w -s -extldflags '-static'" \
  -tags netgo,osusergo \
  -o server ./cmd/server
```

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 生产构建
npm run build
```

### 数据库迁移

初始化脚本已自动执行。如需手动执行：

```bash
docker exec -i crossbow-timescale psql -U postgres -d crossbow_db < database/init.sql
```

---

## 📚 API 概览

### 连弩管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/crossbows` | 连弩列表 |
| GET | `/api/v1/crossbows/:id` | 连弩详情 |
| POST | `/api/v1/crossbows` | 创建连弩 |
| PUT | `/api/v1/crossbows/:id` | 更新连弩 |

### 仿真控制

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/crossbows/:id/start` | 启动仿真 |
| POST | `/api/v1/crossbows/:id/stop` | 停止仿真 |
| POST | `/api/v1/crossbows/:id/reset` | 重置仿真 |

### 传感器数据

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/sensor/data` | 上报传感器数据 |
| POST | `/api/v1/data/query` | 查询历史数据 |

### 强化学习

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/rl/train/:id` | 启动RL训练 |
| GET | `/api/v1/rl/status/:id` | RL训练状态 |
| GET | `/api/v1/rl/result/:id` | RL优化结果 |
| POST | `/api/v1/rl/pause/:id` | 暂停训练 |
| POST | `/api/v1/rl/resume/:id` | 恢复训练 |

### 告警

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/alerts` | 告警列表 |
| POST | `/api/v1/alerts/:id/ack` | 确认告警 |

### 系统

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| GET | `/api/v1/system-stats` | 系统运行状态 |
| GET | `/metrics` | Prometheus指标 |
| GET | `/debug/pprof/` | pprof性能分析（端口6060） |

### WebSocket

| 路径 | 说明 |
|------|------|
| `/ws/crossbow/:id` | 实时数据推送（传感器状态、动力学、告警） |

---

## 🔍 故障排查

### 后端无法连接数据库

```bash
# 查看数据库状态
docker-compose ps timescale

# 查看数据库日志
docker-compose logs timescale

# 测试连接
docker exec -it crossbow-timescale psql -U postgres -d crossbow_db -c "SELECT 1;"
```

### 传感器数据未上报

```bash
# 查看模拟器日志
docker-compose --profile simulator logs -f sensor-simulator

# 查看后端日志
docker-compose logs -f backend

# 手动测试API
curl -X POST http://localhost:8080/api/v1/sensor/data \
  -H "Content-Type: application/json" \
  -d '{"crossbowId":"test","timestamp":"2024-01-01T00:00:00Z","stringTension":500,"bowArmDeformation":10,"magazinePosition":0.5,"fireRate":10,"arrowVelocity":70,"camAngle":180,"stringFatigue":0.1,"temperature":25}'
```

### WebSocket 连接失败

```bash
# 直接测试WebSocket连接
wscat -c ws://localhost:8080/ws/crossbow/crossbow-001

# 检查后端日志中有无WebSocket相关错误
docker-compose logs backend | grep -i websocket
```

### Prometheus 无数据

```bash
# 检查target状态
# 访问 http://localhost:9090/targets

# 手动验证metrics端点
curl http://localhost:8080/metrics
```

---

## 📄 License

本项目为研究用途，仅供学习和仿真参考。

# Push Notification Feature — 消息推送功能设计 Spec

**日期:** 2026-06-05
**状态:** 已确认
**架构方案:** B — 模块化推送系统

---

## 概述

通过 PushPlus 微信公众号推送房租催收提醒给房东本人。支持两种调度模式：
按缴租日提前推送、或固定每周/每月推送。消息模板可自定义编辑，内含变量占位符自动替换。

### 通知对象

**房东本人。**PushPlus 默认推送到 Token 对应的微信公众号粉丝（即房东自己的微信），和每笔收款的租客无关。

---

## PushPlus API 参考

| 项目 | 值 |
|------|-----|
| 端点 | `POST http://www.pushplus.plus/send` |
| Content-Type | `application/json` |
| 默认 channel | `wechat`（微信公众号，免费） |
| 模板格式 | `html`（支持 HTML 标签） |
| 认证 | `token` 参数 |
| 超时 | 10s |
| 响应 | `{code: 200, msg: "请求成功", data: "<流水号>"}` |

Token 可通过两种方式提供：数据库设置页面 或 `.env` 文件。数据库值优先，空值回退到 `.env`。

---

## 数据模型

### PushConfig（push_configs 表）

单例配置，ID=1。首次运行时自动创建默认值。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| PushPlusToken | string | `""` | 空=使用 .env 回退；均空=禁用推送 |
| PushMode | string | `"summary"` | `summary` 汇总 / `individual` 逐条 |
| PushTime | string | `"09:00"` | HH:MM 格式 |
| ScheduleMode | string | `"advance"` | `advance` 按缴租日提前 / `fixed` 固定周期 |
| AdvanceDays | int | `3` | 提前 N 天；0=当天 |
| FixedPeriod | string | `"weekly"` | `weekly` 每周 / `monthly` 每月 |
| FixedWeekday | int | `2` | 1=Mon..7=Sun |
| FixedMonthDay | int | `1` | 1-31；月末边界自动处理 |
| Enabled | bool | `true` | 总开关 |

### PushTemplate（push_templates 表）

消息模板。当前仅一条"房租催收"模板，预留扩展。

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | 模板名称，如 `"房租催收"` |
| TitleTpl | string | 标题模板 |
| BodyTpl | text | 内容模板，含 `{{变量}}` 占位符 |

### PushLog（push_logs 表）

推送日志。不做去重检查，每次 cron 触发都发送。

| 字段 | 类型 | 说明 |
|------|------|------|
| PushType | string | `scheduled` 定时 / `test` 测试 |
| Mode | string | `summary` / `individual` |
| Title | string | 实际发送的标题 |
| Content | text | 替换变量后的实际内容 |
| Recipient | string | 固定 `"landlord"` |
| Status | string | `success` / `failed` / `skipped` |
| PushPlusID | string | PushPlus 返回的流水号 |
| ErrorMsg | string | 失败原因 |
| PaymentIDs | string | JSON 数组：[1, 2, 3] |
| CreatedAt | time | 索引 |

---

## 模板变量

### 单条变量（逐条/汇总均可用）

| 占位符 | 数据来源 | 示例 |
|--------|----------|------|
| `{{姓名}}` | Tenant.Name | 张三 |
| `{{手机号}}` | Tenant.Phone | 138xxxx8888 |
| `{{房源号}}` | Room.RoomNo | A101 |
| `{{房源标题}}` | Room.Title | 阳光大单间 |
| `{{金额}}` | Payment.Amount（分→元，¥ 前缀） | ¥1,500.00 |
| `{{费用类型}}` | Payment.Type 中文映射 | 租金 |
| `{{应缴日期}}` | Payment.PayDate 格式化 | 2026年6月5日 |
| `{{月租金}}` | Tenant.RentPrice（分→元） | 1500.00 |
| `{{入住日期}}` | Tenant.CheckinDate 格式化 | 2025年1月1日 |
| `{{逾期天数}}` | today - PayDate（非负） | 3 |

### 全局变量（一次性替换）

| 占位符 | 数据来源 |
|--------|----------|
| `{{房东姓名}}` | Settings / Config |
| `{{房东电话}}` | Settings / Config |

### 汇总变量（仅汇总模式）

| 占位符 | 数据来源 |
|--------|----------|
| `{{待收总额}}` | SUM(Payment.Amount) → 元 |
| `{{待收笔数}}` | COUNT(payments) |

### 模板循环语法（汇总模式）

汇总模式使用 `{% for item in items %}...{% endfor %}` 标记循环块，每条记录填充一次。

### 默认模板内容

```
房东您好，当前有以下房租待收：

{% for item in items %}
<b>{{姓名}}</b> — {{房源号}} {{房源标题}}
📞 {{手机号}}
💰 应缴金额：<b>¥{{金额}}</b>（{{费用类型}}）
📅 应缴日期：{{应缴日期}}

{% endfor %}
<hr>
💰 待收总计：<b>¥{{待收总额}}</b>（{{待收笔数}}笔）

—— {{房东姓名}} · {{房东电话}}
```

---

## 推送模式

### 汇总模式 (summary)

1. 扫描所有到期/逾期未付记录
2. 替换全局变量（房东信息）
3. 遍历每条记录填充 `{% for %}` 块
4. 替换汇总变量（待收总额、笔数）
5. 拼接成一条消息，发送一次 PushPlus 请求

### 逐条模式 (individual)

1. 扫描所有到期/逾期未付记录
2. 每条记录单独替换变量
3. 每条记录发送一次 PushPlus 请求（N 条记录 = N 次请求）
4. 不含 `{% for %}` 循环；汇总变量忽略

---

## 调度模式

### 模式 A：advance（按缴租日提前推送）

- **每天**在 PushTime 执行
- 查询窗口：`payDate - advanceDays <= today`
- 满足条件的未付记录纳入推送

### 模式 B：fixed（固定周期推送）

- 每天在 PushTime 执行
- 先判断今天是否匹配固定日：
  - weekly: `today.weekday == FixedWeekday`
  - monthly: `today.day == FixedMonthDay`（若当月无此日则取月末最后一天）
- 匹配后扫描所有到期/逾期未付记录纳入推送

### 查询条件（两种模式共用）

| 条件 | 值 |
|------|-----|
| Paid | `false` |
| Excluded | `false` |
| Tenant.Status | `active` |
| Preload | Tenant, Room |

---

## 架构

### 新增文件

| 文件 | 职责 | 预估行数 |
|------|------|----------|
| `internal/model/push.go` | PushConfig, PushTemplate, PushLog GORM 模型 | ~60 |
| `internal/repository/push_repo.go` | 配置 CRUD、日志写入 | ~80 |
| `internal/service/push_service.go` | 模板替换、PushPlus API 调用 | ~200 |
| `internal/service/push_scheduler.go` | robfig/cron 定时调度 | ~120 |
| `internal/handler/admin_push.go` | 设置保存、测试推送 | ~100 |
| `templates/admin/push_config.html` | 推送配置 + 模板编辑 UI | ~120 |
| `templates/admin/push_test_result.html` | 测试推送结果 feedback | ~30 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `config/config.go` | +PUSHPLUS_TOKEN 字段（~5 行） |
| `.env.example` | +PUSHPLUS_TOKEN 行（~2 行） |
| `internal/server/` | 注册路由、注入依赖（~15 行） |
| `cmd/server/main.go` | AutoMigrate 新表、启动 scheduler（~10 行） |
| `templates/admin/settings.html` | 引入 push_config.html partial（~5 行） |
| `internal/handler/admin_settings.go` | Page 方法注入 PushConfig + PushTemplate 数据（~8 行） |
| `internal/service/settings_service.go` | 新增 GetPushSettings 方法（~3 行） |

### 依赖注入链

```
PushRepo → PushService → PushHandler + PushScheduler
PushScheduler.Start(ctx) → goroutine in main.go
```

### 路由

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/settings/push` | 保存推送配置 + 模板 |
| POST | `/admin/settings/push/test` | 发送测试推送到微信 |

推送配置通过 AdminSettingsHandler.Page 加载 settings 页面时注入，不额外添加 GET 路由。

---

## 错误处理

| 场景 | 行为 |
|------|------|
| Token 未配置（DB + .env 均空） | 跳过，PushLog status=skipped |
| PushPlus API 超时（>10s） | 记录 failed + error_msg，不重试 |
| PushPlus 返回非 200 | 记录 failed + 返回msg，不重试 |
| 无到期/逾期记录 | 静默跳过，不推送也不记日志 |
| 模板内容为空 | 跳过，记录 skipped |
| Enabled=false | Scheduler cron 运行但跳过逻辑 |

---

## Scheduler 生命周期

- 在 `main.go` 中作为一个 goroutine 启动
- 监听 parent context 的 `Done()` 信号优雅退出
- 每次触发时重新读取配置（支持运行时修改即时生效，不需要重启 scheduler）
- 使用 `robfig/cron/v3` 库管理 cron 表达式

---

## .env 变更

```env
# 消息推送（可选，也可在设置页面配置）
PUSHPLUS_TOKEN=
```

---

## 测试策略

1. **单元测试** — PushService 模板替换逻辑（各种变量组合）
2. **集成测试** — PushRepo CRUD 操作
3. **手动测试** — 设置页面表单验证 + 测试推送按钮验证真实 API 连通性
4. **Scheduler 测试** — 独立测试 shouldRunToday 和 findDuePayments 逻辑

---

## 不做的事（YAGNI）

- ~~手动推送到单个租客~~ — 暂时不需要，以后可扩展
- ~~消息去重~~ — 房东在 PushPlus 管理页自行管理
- ~~失败重试~~ — 不重试，通过日志排查
- ~~推送历史列表页~~ — 日志只记不查，用 sqlite3 命令行查看
- ~~多模板支持~~ — 当前一条模板够用，预留扩展能力

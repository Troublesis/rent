package model

import "time"

const (
	PushModeSummary    = "summary"
	PushModeIndividual = "individual"
)

const (
	ScheduleModeAdvance = "advance"
	ScheduleModeFixed   = "fixed"
)

const (
	FixedPeriodWeekly  = "weekly"
	FixedPeriodMonthly = "monthly"
)

const (
	PushTypeScheduled = "scheduled"
	PushTypeTest      = "test"
)

const (
	PushStatusSuccess = "success"
	PushStatusFailed  = "failed"
	PushStatusSkipped = "skipped"
)

type PushConfig struct {
	ID            uint   `gorm:"primarykey"`
	PushPlusToken string `gorm:"default:''"`
	PushMode      string `gorm:"default:summary"`
	PushTime      string `gorm:"default:09:00"`
	ScheduleMode  string `gorm:"default:advance"`
	AdvanceDays   int    `gorm:"default:3"`
	FixedPeriod   string `gorm:"default:weekly"`
	FixedWeekday  int    `gorm:"default:2"`
	FixedMonthDay int    `gorm:"default:1"`
	Enabled       bool   `gorm:"default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (PushConfig) TableName() string { return "push_configs" }

type PushTemplate struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"not null"`
	TitleTpl  string
	BodyTpl   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PushTemplate) TableName() string { return "push_templates" }

type PushLog struct {
	ID         uint   `gorm:"primarykey"`
	PushType   string `gorm:"index"`
	Mode       string
	Title      string
	Content    string `gorm:"type:text"`
	Recipient  string
	Status     string `gorm:"index"`
	PushPlusID string
	ErrorMsg   string
	PaymentIDs string
	CreatedAt  time.Time `gorm:"index"`
}

func (PushLog) TableName() string { return "push_logs" }

func DefaultPushTemplate() PushTemplate {
	return PushTemplate{
		Name:     "房租催收",
		TitleTpl: "📋 房租催收提醒",
		BodyTpl: `房东您好，当前有以下房租待收：

{% for item in items %}
<b>{{姓名}}</b> — {{房源号}} {{房源标题}}
📞 {{手机号}}
💰 应缴金额：<b>¥{{金额}}</b>（{{费用类型}}）
📅 应缴日期：{{应缴日期}}

{% endfor %}
<hr>
💰 待收总计：<b>¥{{待收总额}}</b>（{{待收笔数}}笔）

—— {{房东姓名}} · {{房东电话}}`,
	}
}

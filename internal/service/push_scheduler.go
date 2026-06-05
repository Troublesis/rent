package service

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/troublesis/rent/internal/model"
	"github.com/troublesis/rent/internal/repository"
)

type PushScheduler struct {
	pushRepo    *repository.PushRepository
	pushService *PushService
	cron        *cron.Cron
}

func NewPushScheduler(pushRepo *repository.PushRepository, pushService *PushService) *PushScheduler {
	return &PushScheduler{
		pushRepo:    pushRepo,
		pushService: pushService,
	}
}

func (ps *PushScheduler) Start(ctx context.Context) {
	ps.cron = cron.New(cron.WithSeconds())
	// Tick every minute, but gate on HH:MM match inside tick()
	ps.cron.AddFunc("0 * * * * *", func() { ps.tick() })
	ps.cron.Start()

	go func() {
		<-ctx.Done()
		ps.cron.Stop()
	}()
}

func (ps *PushScheduler) tick() {
	config, err := ps.pushRepo.GetConfigOrCreate()
	if err != nil {
		log.Printf("[push-scheduler] get config: %v", err)
		return
	}
	if !config.Enabled {
		return
	}

	now := time.Now()
	if !ps.matchesPushTime(config.PushTime, now) {
		return
	}
	if !ps.shouldRunToday(config, now) {
		return
	}

	if err := ps.pushService.SendReminder(); err != nil {
		log.Printf("[push-scheduler] send reminder: %v", err)
	}
}

func (ps *PushScheduler) matchesPushTime(pushTime string, now time.Time) bool {
	parts := strings.Split(strings.TrimSpace(pushTime), ":")
	if len(parts) != 2 {
		return false
	}
	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return now.Hour() == hour && now.Minute() == minute
}

func (ps *PushScheduler) shouldRunToday(config *model.PushConfig, now time.Time) bool {
	if config.ScheduleMode == model.ScheduleModeAdvance {
		return true
	}
	if config.FixedPeriod == model.FixedPeriodWeekly {
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return weekday == config.FixedWeekday
	}
	if config.FixedPeriod == model.FixedPeriodMonthly {
		targetDay := config.FixedMonthDay
		lastDay := lastDayOfMonth(now.Year(), now.Month())
		if targetDay > lastDay {
			targetDay = lastDay
		}
		return now.Day() == targetDay
	}
	return false
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

package repository

import (
	"testing"

	"github.com/troublesis/rent/internal/model"
)

func newPushTestDB(t *testing.T) *PushRepository {
	db := newTestDB(t)
	if err := db.AutoMigrate(&model.PushConfig{}, &model.PushTemplate{}, &model.PushLog{}); err != nil {
		t.Fatalf("migrate push models: %v", err)
	}
	return NewPushRepository(db)
}

func TestGetConfigOrCreate(t *testing.T) {
	repo := newPushTestDB(t)
	config, err := repo.GetConfigOrCreate()
	if err != nil {
		t.Fatalf("GetConfigOrCreate: %v", err)
	}
	if config.ID != 1 {
		t.Errorf("expected ID=1, got %d", config.ID)
	}
	if config.PushMode != model.PushModeSummary {
		t.Errorf("expected default PushMode=summary, got %s", config.PushMode)
	}
	if config.PushTime != "09:00" {
		t.Errorf("expected default PushTime=09:00, got %s", config.PushTime)
	}
}

func TestGetConfigOrCreate_Existing(t *testing.T) {
	repo := newPushTestDB(t)
	config1, _ := repo.GetConfigOrCreate()
	config1.PushMode = model.PushModeIndividual
	if err := repo.SaveConfig(config1); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	config2, err := repo.GetConfigOrCreate()
	if err != nil {
		t.Fatalf("GetConfigOrCreate: %v", err)
	}
	if config2.PushMode != model.PushModeIndividual {
		t.Errorf("expected PushMode=individual, got %s", config2.PushMode)
	}
}

func TestSaveConfig(t *testing.T) {
	repo := newPushTestDB(t)
	config, _ := repo.GetConfigOrCreate()
	config.AdvanceDays = 7
	config.PushTime = "10:30"
	if err := repo.SaveConfig(config); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	reloaded, err := repo.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if reloaded.AdvanceDays != 7 {
		t.Errorf("expected AdvanceDays=7, got %d", reloaded.AdvanceDays)
	}
	if reloaded.PushTime != "10:30" {
		t.Errorf("expected PushTime=10:30, got %s", reloaded.PushTime)
	}
}

func TestGetTemplateOrCreate(t *testing.T) {
	repo := newPushTestDB(t)
	tmpl, err := repo.GetTemplateOrCreate()
	if err != nil {
		t.Fatalf("GetTemplateOrCreate: %v", err)
	}
	if tmpl.Name != "房租催收" {
		t.Errorf("expected template name, got %s", tmpl.Name)
	}
	if tmpl.BodyTpl == "" {
		t.Error("expected non-empty body template")
	}
}

func TestSaveTemplate(t *testing.T) {
	repo := newPushTestDB(t)
	tmpl, _ := repo.GetTemplateOrCreate()
	tmpl.TitleTpl = "新标题"
	if err := repo.SaveTemplate(tmpl); err != nil {
		t.Fatalf("SaveTemplate: %v", err)
	}
	reloaded, err := repo.GetTemplate()
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if reloaded.TitleTpl != "新标题" {
		t.Errorf("expected TitleTpl, got %s", reloaded.TitleTpl)
	}
}

func TestCreateLog(t *testing.T) {
	repo := newPushTestDB(t)
	log := &model.PushLog{
		PushType:  model.PushTypeTest,
		Mode:      model.PushModeSummary,
		Status:    model.PushStatusSuccess,
		Recipient: "landlord",
		Title:     "测试",
		Content:   "内容",
	}
	if err := repo.CreateLog(log); err != nil {
		t.Fatalf("CreateLog: %v", err)
	}
	if log.ID == 0 {
		t.Error("expected ID to be assigned")
	}
}

package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/Yoshikrit/observability-test/internal/model"
)

func InitDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(&model.Task{})
}

func SeedDatabase(db *gorm.DB) {
	var count int64
	db.Model(&model.Task{}).Count(&count)
	if count > 0 {
		return
	}

	seed := []model.Task{
		{Title: "Set up observability stack", Description: "Wire up OpenTelemetry + Grafana LGTM", Done: false},
		{Title: "Learn Fiber basics", Description: "Routing, middleware, error handling", Done: true},
	}
	db.Create(&seed)
}

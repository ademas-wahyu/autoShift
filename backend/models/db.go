package models

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dialector gorm.Dialector) {
	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	log.Println("database connected")
}

func Migrate() {
	err := DB.AutoMigrate(
		&Tenant{},
		&User{},
		&EmployeeRole{},
		&Employee{},
		&EmployeeLeaveQuota{},
		&ShiftTemplate{},
		&RoleRequirement{},
		&Schedule{},
		&ScheduleFixedLeave{},
		&ScheduleRandomLeave{},
		&ScheduleShift{},
		&ScheduleEmployeeLeave{},
		&Holiday{},
		&GenerationLog{},
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("database migrated")
}

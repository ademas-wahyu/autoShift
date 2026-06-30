package models

import "time"

type ShiftTemplate struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TenantID   uint   `gorm:"not null;uniqueIndex:idx_st_tenant" json:"tenant_id"`
	Name       string `gorm:"size:100;not null;uniqueIndex:idx_st_tenant" json:"name"`
	StartTime  string `gorm:"size:5;not null" json:"start_time"`
	EndTime    string `gorm:"size:5;not null" json:"end_time"`
	IsCrossDay bool   `gorm:"default:false" json:"is_cross_day"`
	ColorHex   string `gorm:"size:7;default:'#3B82F6'" json:"color_hex"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

func (ShiftTemplate) TableName() string { return "shift_templates" }

type ShiftTemplateConfig struct {
	ID       uint `json:"id"`
	Capacity int  `json:"capacity"`
}

type RoleRequirement struct {
	ID              uint `gorm:"primaryKey" json:"id"`
	ShiftTemplateID uint `gorm:"not null;uniqueIndex:idx_rr_st_role" json:"shift_template_id"`
	RoleID          uint `gorm:"not null;uniqueIndex:idx_rr_st_role" json:"role_id"`
	MinCount        int  `gorm:"default:1" json:"min_count"`

	ShiftTemplate ShiftTemplate `gorm:"foreignKey:ShiftTemplateID" json:"-"`
	Role          EmployeeRole  `gorm:"foreignKey:RoleID" json:"-"`
}

func (RoleRequirement) TableName() string { return "shift_role_requirements" }

// ── Schedule ────────────────────────────────────────────────

type LeaveMode string

const (
	LeaveModeFixed  LeaveMode = "fixed"
	LeaveModeRandom LeaveMode = "random"
)

type ScheduleStatus string

const (
	StatusDraft     ScheduleStatus = "draft"
	StatusPublished ScheduleStatus = "published"
)

type Schedule struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;uniqueIndex:idx_sch_tenant_month" json:"tenant_id"`
	Month       int            `gorm:"not null;uniqueIndex:idx_sch_tenant_month" json:"month"`
	Year        int            `gorm:"not null;uniqueIndex:idx_sch_tenant_month" json:"year"`
	LeaveMode   LeaveMode      `gorm:"type:varchar(10);not null" json:"leave_mode"`
	Status      ScheduleStatus `gorm:"type:varchar(10);default:'draft'" json:"status"`
	Notes       *string        `gorm:"type:text" json:"notes"`
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	PublishedAt *time.Time     `json:"published_at"`
}

func (Schedule) TableName() string { return "schedules" }

// ── Leave Configs ───────────────────────────────────────────

type ScheduleFixedLeave struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	ScheduleID uint `gorm:"not null;uniqueIndex:idx_sfl_sch_day" json:"schedule_id"`
	DayOfWeek  int  `gorm:"not null;uniqueIndex:idx_sfl_sch_day" json:"day_of_week"`
}

func (ScheduleFixedLeave) TableName() string { return "schedule_fixed_leaves" }

type ScheduleRandomLeave struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	ScheduleID   uint `gorm:"not null;uniqueIndex" json:"schedule_id"`
	DaysPerWeek  int  `gorm:"not null" json:"days_per_week"`
}

func (ScheduleRandomLeave) TableName() string { return "schedule_random_leaves" }

// ── Assignments ─────────────────────────────────────────────

type ScheduleShift struct {
	ID              uint  `gorm:"primaryKey" json:"id"`
	ScheduleID      uint  `gorm:"not null;index" json:"schedule_id"`
	EmployeeID      uint  `gorm:"not null;index" json:"employee_id"`
	ShiftTemplateID uint  `gorm:"not null" json:"shift_template_id"`
	Date            string `gorm:"size:10;not null" json:"date"`
	IsOverride      bool  `gorm:"default:false" json:"is_override"`
}

func (ScheduleShift) TableName() string { return "schedule_shifts" }

type ScheduleEmployeeLeave struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	ScheduleID uint `gorm:"not null;index" json:"schedule_id"`
	EmployeeID uint `gorm:"not null;index" json:"employee_id"`
	Date       string `gorm:"size:10;not null" json:"date"`
	IsOverride bool  `gorm:"default:false" json:"is_override"`
}

func (ScheduleEmployeeLeave) TableName() string { return "schedule_employee_leaves" }

// ── Generation Log ──────────────────────────────────────────

type GenerationLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ScheduleID      uint      `gorm:"not null;index" json:"schedule_id"`
	BatchIndex      *int      `json:"batch_index"`
	Status          string    `gorm:"size:50" json:"status"`
	PromptTokens    *int      `json:"prompt_tokens"`
	ResponseTimeMs  *int      `json:"response_time_ms"`
	ErrorMessage    *string   `gorm:"type:text" json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
}

func (GenerationLog) TableName() string { return "generation_logs" }

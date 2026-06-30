package models

type Holiday struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Date       string `gorm:"size:10;uniqueIndex;not null" json:"date"`
	Name       string `gorm:"size:255;not null" json:"name"`
	IsNational bool   `gorm:"default:true" json:"is_national"`
}

func (Holiday) TableName() string { return "holidays" }

// ── API Response Types ──────────────────────────────────────

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ScheduleDetail struct {
	Schedule          Schedule               `json:"schedule"`
	ShiftTemplates    []ShiftTemplate        `json:"shift_templates"`
	Shifts            []ScheduleShift        `json:"shifts"`
	Leaves            []ScheduleEmployeeLeave `json:"leaves"`
	Holidays          []Holiday              `json:"holidays"`
	GenerationErrors  []ValidationViolation  `json:"generation_errors"`
	FairnessMetrics   *FairnessMetrics       `json:"fairness_metrics"`
	GenerationSummary *GenerationSummary     `json:"generation_summary"`
}

type ValidationViolation struct {
	Type            string `json:"type"`
	Severity        string `json:"severity"`
	Message         string `json:"message"`
	EmployeeID      uint   `json:"employee_id"`
	EmployeeName    string `json:"employee_name,omitempty"`
	Date            string `json:"date,omitempty"`
	ShiftTemplateID uint   `json:"shift_template_id,omitempty"`
	Quota           int    `json:"quota,omitempty"`
	Actual          int    `json:"actual,omitempty"`
}

type FairnessMetrics struct {
	TotalEmployees          int     `json:"total_employees"`
	AvgShiftHours           float64 `json:"avg_shift_hours"`
	StdDevHours             float64 `json:"std_dev_hours"`
	StdDevWeekendShifts     float64 `json:"std_dev_weekend_shifts"`
	WeekendShiftsPerEmployee WeekendShiftStats `json:"weekend_shifts_per_employee"`
}

type WeekendShiftStats struct {
	Min int     `json:"min"`
	Max int     `json:"max"`
	Avg float64 `json:"avg"`
}

type GenerationSummary struct {
	TotalBatches         int `json:"total_batches"`
	CompletedBatches     int `json:"completed_batches"`
	FailedBatches        int `json:"failed_batches"`
	TotalShiftsGenerated int `json:"total_shifts_generated"`
	TotalLeavesGenerated int `json:"total_leaves_generated"`
	TotalErrors          int `json:"total_errors"`
}

type ShiftChange struct {
	Action          string `json:"action"`
	ScheduleShiftID *uint  `json:"schedule_shift_id,omitempty"`
	ScheduleLeaveID *uint  `json:"schedule_leave_id,omitempty"`
	EmployeeID      uint   `json:"employee_id"`
	ShiftTemplateID *uint  `json:"shift_template_id,omitempty"`
	Date            string `json:"date"`
	Reason          string `json:"reason,omitempty"`
}

type UpdateShiftsRequest struct {
	ShiftChanges []ShiftChange `json:"shift_changes"`
	LeaveChanges []ShiftChange `json:"leave_changes"`
}

type CreateScheduleRequest struct {
	TenantID         uint                   `json:"tenant_id" validate:"required"`
	Month            int                    `json:"month" validate:"required,min=1,max=12"`
	Year             int                    `json:"year" validate:"required"`
	LeaveMode        string                 `json:"leave_mode" validate:"required"`
	FixedLeaves      []int                  `json:"fixed_leaves"`
	RandomDaysPerWeek *int                  `json:"random_days_per_week"`
	ShiftTemplates   []ShiftTemplateConfig  `json:"shift_templates" validate:"required"`
	RoleRequirements []RoleRequirementInput `json:"role_requirements"`
	EmployeeIDs      []uint                 `json:"employee_ids" validate:"required"`
}

type RoleRequirementInput struct {
	ShiftTemplateID uint `json:"shift_template_id"`
	RoleID          uint `json:"role_id"`
	MinCount        int  `json:"min_count"`
}

type ValidationResult struct {
	IsValid    bool                 `json:"is_valid"`
	Violations []ValidationViolation `json:"violations"`
}

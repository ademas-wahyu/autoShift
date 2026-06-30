package services

import (
	"fmt"
	"log"
	"time"

	"github.com/ademaswahyu/autoshift-backend/ai"
	"github.com/ademaswahyu/autoshift-backend/models"
)

type SchedulerEngine struct {
	validator *Validator
	generator *ai.Generator
}

func NewSchedulerEngine(validator *Validator, generator *ai.Generator) *SchedulerEngine {
	return &SchedulerEngine{
		validator: validator,
		generator: generator,
	}
}

// GenerateSchedule runs the full AI generation pipeline.
func (e *SchedulerEngine) GenerateSchedule(scheduleID uint, req models.CreateScheduleRequest) (*models.ScheduleDetail, []models.ValidationViolation, error) {
	// 1. Load data
	var templates []models.ShiftTemplate
	models.DB.Where("tenant_id = ?", req.TenantID).Find(&templates)

	var roleReqs []models.RoleRequirement
	models.DB.Where("shift_template_id IN ?", templateIDs(templates)).Find(&roleReqs)

	var employees []models.Employee
	models.DB.Preload("Role").Where("id IN ? AND tenant_id = ?", req.EmployeeIDs, req.TenantID).Find(&employees)

	var quotas []models.EmployeeLeaveQuota
	models.DB.Where("employee_id IN ? AND month = ? AND year = ?", req.EmployeeIDs, req.Month, req.Year).Find(&quotas)

	holidays := getHolidaysForMonth(req.Month, req.Year)

	// 2. Build period
	periodStart := fmt.Sprintf("%d-%02d-01", req.Year, req.Month)
	periodEnd := lastDayOfMonth(req.Year, req.Month)

	// 3. Determine leave mode details
	var fixedLeaves []int
	var randomDaysPerWeek *int
	if req.LeaveMode == "fixed" {
		fixedLeaves = req.FixedLeaves
	} else {
		randomDaysPerWeek = req.RandomDaysPerWeek
	}

	// 4. Build batch input
	input := ai.BatchInput{
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		ShiftTemplates:   templates,
		LeaveMode:        req.LeaveMode,
		RandomDaysPerWeek: randomDaysPerWeek,
		FixedLeaves:      fixedLeaves,
		Holidays:         holidays,
		RoleRequirements: roleReqs,
		MinRestHours:     e.validator.MinRestHours,
	}

	// 5. Generate batches
	log.Printf("generating schedule %d for period %s - %s", scheduleID, periodStart, periodEnd)
	shifts, leaves, genErrors, err := e.generator.Generate(input, employees, quotas)
	if err != nil {
		return nil, nil, fmt.Errorf("generation failed: %w", err)
	}

	// 6. Save results
	tx := models.DB.Begin()

	for i := range shifts {
		shifts[i].ScheduleID = scheduleID
	}
	if len(shifts) > 0 {
		tx.CreateInBatches(&shifts, 500)
	}

	for i := range leaves {
		leaves[i].ScheduleID = scheduleID
	}
	if len(leaves) > 0 {
		tx.CreateInBatches(&leaves, 500)
	}

	tx.Commit()

	// Reload from DB to get IDs
	var savedShifts []models.ScheduleShift
	models.DB.Where("schedule_id = ?", scheduleID).Find(&savedShifts)
	var savedLeaves []models.ScheduleEmployeeLeave
	models.DB.Where("schedule_id = ?", scheduleID).Find(&savedLeaves)

	// 7. Validate
	result := e.validator.ValidateSchedule(scheduleID, savedShifts, savedLeaves, templates, employees, roleReqs, quotas)
	allViolations := append(genErrors, result.Violations...)

	// 8. Calculate fairness
	metrics := CalculateFairnessMetrics(savedShifts, templates, employees)

	// 9. Build detail
	detail := &models.ScheduleDetail{
		ShiftTemplates:   templates,
		Shifts:           savedShifts,
		Leaves:           savedLeaves,
		Holidays:         holidays,
		GenerationErrors: allViolations,
		FairnessMetrics:  metrics,
		GenerationSummary: &models.GenerationSummary{
			TotalBatches:         ((len(employees) + 19) / 20),
			CompletedBatches:     ((len(employees) + 19) / 20),
			TotalShiftsGenerated: len(savedShifts),
			TotalLeavesGenerated: len(savedLeaves),
			TotalErrors:          len(allViolations),
		},
	}

	return detail, allViolations, nil
}

func templateIDs(templates []models.ShiftTemplate) []uint {
	ids := make([]uint, len(templates))
	for i, t := range templates {
		ids[i] = t.ID
	}
	return ids
}

func getHolidaysForMonth(month, year int) []models.Holiday {
	var holidays []models.Holiday
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	endDate := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	models.DB.Where("date >= ? AND date < ?", startDate, endDate).Find(&holidays)
	return holidays
}

func lastDayOfMonth(year, month int) string {
	t := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC)
	return t.Format("2006-01-02")
}

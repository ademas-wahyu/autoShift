package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/ademaswahyu/autoshift-backend/models"
	"github.com/ademaswahyu/autoshift-backend/services"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type ScheduleHandler struct {
	engine *services.SchedulerEngine
}

func NewScheduleHandler(engine *services.SchedulerEngine) *ScheduleHandler {
	return &ScheduleHandler{engine: engine}
}

// ── POST /schedules ─────────────────────────────────────────

func (h *ScheduleHandler) Create(c *fiber.Ctx) error {
	var req models.CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
	}

	userID := extractUserID(c)
	tenantID := extractTenantID(c)

	schedule := models.Schedule{
		TenantID:  tenantID,
		Month:     req.Month,
		Year:      req.Year,
		LeaveMode: models.LeaveMode(req.LeaveMode),
		Status:    models.StatusDraft,
		CreatedBy: userID,
	}

	tx := models.DB.Begin()

	if err := tx.Create(&schedule).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "failed to create schedule: " + err.Error(),
		})
	}

	// Save leave config
	if req.LeaveMode == "fixed" {
		for _, dow := range req.FixedLeaves {
			tx.Create(&models.ScheduleFixedLeave{
				ScheduleID: schedule.ID,
				DayOfWeek:  dow,
			})
		}
	} else if req.RandomDaysPerWeek != nil {
		tx.Create(&models.ScheduleRandomLeave{
			ScheduleID:  schedule.ID,
			DaysPerWeek: *req.RandomDaysPerWeek,
		})
	}

	tx.Commit()

	// Populate employee IDs from tenant if not provided
	if len(req.EmployeeIDs) == 0 {
		var emps []models.Employee
		models.DB.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&emps)
		for _, e := range emps {
			req.EmployeeIDs = append(req.EmployeeIDs, e.ID)
		}
	}
	req.TenantID = tenantID

	// Trigger AI generation async
	go func() {
		start := time.Now()
		detail, violations, err := h.engine.GenerateSchedule(schedule.ID, req)
		elapsed := time.Since(start)

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		models.DB.Create(&models.GenerationLog{
			ScheduleID:      schedule.ID,
			Status:          "completed",
			ResponseTimeMs:  intPtr(int(elapsed.Milliseconds())),
			ErrorMessage:    strPtr(errMsg),
		})

		if err != nil {
			log.Printf("generation failed for schedule %d: %v", schedule.ID, err)
			return
		}

		models.DB.Model(&schedule).Updates(map[string]interface{}{
			"notes": formatErrors(violations),
		})

		for _, v := range violations {
			if v.Type == "rest_time_violation" && v.Date != "" {
				log.Printf("Conflict: %s on %s", v.Message, v.Date)
			}
		}

		log.Printf("schedule %d generated in %v: %d shifts, %d leaves, %d errors",
			schedule.ID, elapsed, detail.GenerationSummary.TotalShiftsGenerated,
			detail.GenerationSummary.TotalLeavesGenerated, len(violations))
	}()

	return c.Status(201).JSON(models.APIResponse{
		Success: true,
		Data: fiber.Map{
			"schedule_id": schedule.ID,
			"status":      "generating",
			"message":     "Jadwal sedang di-generate.",
		},
	})
}

// ── GET /schedules/:id ───────────────────────────────────────

func (h *ScheduleHandler) Get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{Success: false, Error: "invalid id"})
	}

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{Success: false, Error: "schedule not found"})
	}

	// Load all data
	var templates []models.ShiftTemplate
	var shifts []models.ScheduleShift
	var leaves []models.ScheduleEmployeeLeave
	var holidays []models.Holiday
	var errors []models.ValidationViolation
	var logs []models.GenerationLog

	models.DB.Where("tenant_id = ?", schedule.TenantID).Find(&templates)
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&shifts)
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&leaves)
	startDate := fmt.Sprintf("%d-%02d-01", schedule.Year, schedule.Month)
	endDate := time.Date(schedule.Year, time.Month(schedule.Month)+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	models.DB.Where("date >= ? AND date < ?", startDate, endDate).Find(&holidays)

	var employees []models.Employee
	models.DB.Preload("Role").Where("tenant_id = ?", schedule.TenantID).Find(&employees)

	models.DB.Where("schedule_id = ?", schedule.ID).Order("created_at desc").First(&logs)

	metrics := services.CalculateFairnessMetrics(shifts, templates, employees)

	summary := &models.GenerationSummary{
		TotalBatches:         1,
		TotalShiftsGenerated: len(shifts),
		TotalLeavesGenerated: len(leaves),
		TotalErrors:          len(errors),
	}
	if len(logs) > 0 && logs[0].ErrorMessage != nil {
		summary.TotalErrors++
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data: models.ScheduleDetail{
			Schedule:          schedule,
			ShiftTemplates:    templates,
			Shifts:            shifts,
			Leaves:            leaves,
			Holidays:          holidays,
			GenerationErrors:  errors,
			FairnessMetrics:   metrics,
			GenerationSummary: summary,
		},
	})
}

// ── GET /schedules/:id/validate ─────────────────────────────

func (h *ScheduleHandler) Validate(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{Success: false, Error: "invalid id"})
	}

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{Success: false, Error: "schedule not found"})
	}

	var templates []models.ShiftTemplate
	var shifts []models.ScheduleShift
	var leaves []models.ScheduleEmployeeLeave
	var employees []models.Employee
	var roleReqs []models.RoleRequirement
	var quotas []models.EmployeeLeaveQuota

	models.DB.Where("tenant_id = ?", schedule.TenantID).Find(&templates)
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&shifts)
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&leaves)
	models.DB.Preload("Role").Where("tenant_id = ?", schedule.TenantID).Find(&employees)
	models.DB.Where("shift_template_id IN ?", templateIDs(templates)).Find(&roleReqs)
	models.DB.Where("employee_id IN ? AND month = ? AND year = ?",
		employeeIDs(employees), schedule.Month, schedule.Year).Find(&quotas)

	validator := services.NewValidator(12)
	result := validator.ValidateSchedule(schedule.ID, shifts, leaves, templates, employees, roleReqs, quotas)

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// ── PUT /schedules/:id/shifts ────────────────────────────────

func (h *ScheduleHandler) UpdateShifts(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{Success: false, Error: "invalid id"})
	}

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{Success: false, Error: "schedule not found"})
	}

	if schedule.Status == models.StatusPublished {
		return c.Status(409).JSON(models.APIResponse{
			Success: false,
			Error:   "SCHEDULE_ALREADY_PUBLISHED",
			Message: "Tidak bisa edit jadwal yang sudah publish",
		})
	}

	var req models.UpdateShiftsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{Success: false, Error: "invalid body"})
	}

	tx := models.DB.Begin()

	for _, ch := range req.ShiftChanges {
		switch ch.Action {
		case "update":
			if ch.ScheduleShiftID != nil {
				tx.Model(&models.ScheduleShift{}).Where("id = ?", *ch.ScheduleShiftID).Updates(map[string]interface{}{
					"employee_id":      ch.EmployeeID,
					"shift_template_id": ch.ShiftTemplateID,
					"date":             ch.Date,
					"is_override":      true,
				})
			}
		case "delete":
			if ch.ScheduleShiftID != nil {
				tx.Delete(&models.ScheduleShift{}, *ch.ScheduleShiftID)
			}
		case "create":
			tx.Create(&models.ScheduleShift{
				ScheduleID:      uint(id),
				EmployeeID:      ch.EmployeeID,
				ShiftTemplateID: *ch.ShiftTemplateID,
				Date:            ch.Date,
				IsOverride:      true,
			})
		}
	}

	for _, ch := range req.LeaveChanges {
		switch ch.Action {
		case "create":
			tx.Create(&models.ScheduleEmployeeLeave{
				ScheduleID: uint(id),
				EmployeeID: ch.EmployeeID,
				Date:       ch.Date,
				IsOverride: true,
			})
		case "delete":
			if ch.ScheduleLeaveID != nil {
				tx.Delete(&models.ScheduleEmployeeLeave{}, *ch.ScheduleLeaveID)
			}
		}
	}

	tx.Commit()

	// Re-validate
	var shifts []models.ScheduleShift
	var leaves []models.ScheduleEmployeeLeave
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&shifts)
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&leaves)

	var templates []models.ShiftTemplate
	var employees []models.Employee
	var roleReqs []models.RoleRequirement
	var quotas []models.EmployeeLeaveQuota
	models.DB.Where("tenant_id = ?", schedule.TenantID).Find(&templates)
	models.DB.Preload("Role").Where("tenant_id = ?", schedule.TenantID).Find(&employees)

	validator := services.NewValidator(12)
	result := validator.ValidateSchedule(schedule.ID, shifts, leaves, templates, employees, roleReqs, quotas)

	if !result.IsValid {
		return c.Status(422).JSON(models.APIResponse{
			Success: false,
			Error:   "VALIDATION_FAILED",
			Message: "Perubahan menyebabkan pelanggaran constraint",
			Data: fiber.Map{
				"validation_result": result,
			},
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data: fiber.Map{
			"validation_result": result,
		},
	})
}

// ── PUT /schedules/:id/publish ────────────────────────────

func (h *ScheduleHandler) Publish(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{Success: false, Error: "invalid id"})
	}

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{Success: false, Error: "schedule not found"})
	}

	now := time.Now()
	models.DB.Model(&schedule).Updates(map[string]interface{}{
		"status":       models.StatusPublished,
		"published_at": &now,
	})

	// Parse optional notes
	var body struct {
		Notes string `json:"notes"`
	}
	c.BodyParser(&body)
	if body.Notes != "" {
		models.DB.Model(&schedule).Update("notes", body.Notes)
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data: fiber.Map{
			"schedule_id":  schedule.ID,
			"status":       "published",
			"published_at": now,
			"message":      "Jadwal berhasil dipublikasikan",
		},
	})
}

// ── POST /schedules/:id/regenerate ─────────────────────────

func (h *ScheduleHandler) Regenerate(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{Success: false, Error: "invalid id"})
	}

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{Success: false, Error: "schedule not found"})
	}

	var req struct {
		EmployeeIDs []uint `json:"employee_ids"`
		DateFrom    string `json:"date_from"`
		DateTo      string `json:"date_to"`
	}
	c.BodyParser(&req)

	// Delete existing shifts in range
	tx := models.DB.Begin()
	query := tx.Where("schedule_id = ?", schedule.ID)
	if req.DateFrom != "" {
		query = query.Where("date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("date <= ?", req.DateTo)
	}
	if len(req.EmployeeIDs) > 0 {
		query = query.Where("employee_id IN ?", req.EmployeeIDs)
	}
	query.Delete(&models.ScheduleShift{})
	query.Delete(&models.ScheduleEmployeeLeave{})
	tx.Commit()

	createReq := models.CreateScheduleRequest{
		TenantID:  schedule.TenantID,
		Month:     schedule.Month,
		Year:      schedule.Year,
		LeaveMode: string(schedule.LeaveMode),
	}

	var employees []models.Employee
	models.DB.Where("tenant_id = ?", schedule.TenantID).Find(&employees)
	for _, e := range employees {
		createReq.EmployeeIDs = append(createReq.EmployeeIDs, e.ID)
	}

	go func() {
		h.engine.GenerateSchedule(schedule.ID, createReq)
	}()

	return c.JSON(models.APIResponse{
		Success: true,
		Data: fiber.Map{
			"schedule_id": schedule.ID,
			"status":      "generating",
			"message":     "Generate ulang dimulai.",
		},
	})
}

// ── Helpers ────────────────────────────────────────────────

func intPtr(i int) *int { return &i }
func strPtr(s string) *string { return &s }

func formatErrors(violations []models.ValidationViolation) string {
	msg := ""
	for _, v := range violations {
		msg += v.Message + "; "
	}
	if len(msg) > 500 {
		msg = msg[:500]
	}
	return msg
}

func templateIDs(templates []models.ShiftTemplate) []uint {
	ids := make([]uint, len(templates))
	for i, t := range templates {
		ids[i] = t.ID
	}
	return ids
}

func employeeIDs(employees []models.Employee) []uint {
	ids := make([]uint, len(employees))
	for i, e := range employees {
		ids[i] = e.ID
	}
	return ids
}

func extractUserID(c *fiber.Ctx) uint {
	return extractClaimUint(c, "sub")
}

func extractTenantID(c *fiber.Ctx) uint {
	return extractClaimUint(c, "tenant_id")
}

func extractClaimUint(c *fiber.Ctx, claim string) uint {
	token, ok := c.Locals("jwt").(*jwt.Token)
	if !ok {
		token, ok = c.Locals("user").(*jwt.Token)
		if !ok {
			return 0
		}
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0
	}
	v, ok := claims[claim].(float64)
	if !ok {
		return 0
	}
	return uint(v)
}

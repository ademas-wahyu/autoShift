package handlers

import (
	"fmt"

	"github.com/ademaswahyu/autoshift-backend/models"
	"github.com/gofiber/fiber/v2"
)

type ExportHandler struct{}

func NewExportHandler() *ExportHandler {
	return &ExportHandler{}
}

// GET /schedules/:id/export?format=pdf
func (h *ExportHandler) Export(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid id"})
	}

	format := c.Query("format", "xlsx")

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "schedule not found"})
	}

	var shifts []models.ScheduleShift
	var leaves []models.ScheduleEmployeeLeave
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&shifts)
	models.DB.Where("schedule_id = ?", schedule.ID).Find(&leaves)

	var employees []models.Employee
	models.DB.Preload("Role").Where("tenant_id = ?", schedule.TenantID).Find(&employees)

	var templates []models.ShiftTemplate
	models.DB.Where("tenant_id = ?", schedule.TenantID).Find(&templates)

	if format == "xlsx" || format == "excel" {
		return h.exportExcel(c, schedule, shifts, leaves, employees, templates)
	}

	// Default: return JSON data for frontend to generate PDF
	return c.JSON(models.APIResponse{
		Success: true,
		Data: fiber.Map{
			"schedule":   schedule,
			"shifts":     shifts,
			"leaves":     leaves,
			"employees":  employees,
			"templates":  templates,
		},
	})
}

// GET /schedules/:id/share
func (h *ExportHandler) Share(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid id"})
	}

	var schedule models.Schedule
	if result := models.DB.First(&schedule, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "error": "schedule not found"})
	}

	// In real app: generate a signed token
	token := generateShareToken(schedule.ID)

	return c.JSON(models.APIResponse{
		Success: true,
		Data: fiber.Map{
			"share_url":  "https://autoshift.app/s/" + token,
			"is_active":  true,
		},
	})
}

func (h *ExportHandler) exportExcel(c *fiber.Ctx, schedule models.Schedule, shifts []models.ScheduleShift, leaves []models.ScheduleEmployeeLeave, employees []models.Employee, templates []models.ShiftTemplate) error {
	// For MVP: return CSV data
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=schedule.csv")

	c.WriteString("Karyawan,Role,Tanggal,Shift\n")
	leaveMap := make(map[string]bool)
	for _, l := range leaves {
		leaveMap[l.Date] = true
	}

	for _, s := range shifts {
		empName := ""
		empRole := ""
		for _, e := range employees {
			if e.ID == s.EmployeeID {
				empName = e.Name
				empRole = e.Role.Name
				break
			}
		}
		shiftName := ""
		for _, t := range templates {
			if t.ID == s.ShiftTemplateID {
				shiftName = t.Name
				break
			}
		}
		c.WriteString(empName + "," + empRole + "," + s.Date + "," + shiftName + "\n")
	}

	return nil
}

func generateShareToken(scheduleID uint) string {
	// Simplified: use base64 of schedule ID
	return "s" + fmt.Sprintf("%d", scheduleID)
}

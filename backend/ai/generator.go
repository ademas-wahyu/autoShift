package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/ademaswahyu/autoshift-backend/models"
)

type Generator struct {
	provider    string
	apiURL      string
	apiKey      string
	model       string
	batchSize    int
	minRestHours float64
	maxRetries   int
	client       *http.Client
}

type BatchInput struct {
	BatchIndex      int                        `json:"batch_index"`
	TotalBatches    int                        `json:"total_batches"`
	PeriodStart     string                     `json:"period_start"`
	PeriodEnd       string                     `json:"period_end"`
	ShiftTemplates  []models.ShiftTemplate    `json:"shift_templates"`
	LeaveMode       string                     `json:"leave_mode"`
	RandomDaysPerWeek *int                     `json:"random_days_per_week"`
	FixedLeaves     []int                      `json:"fixed_leaves"`
	Holidays        []models.Holiday           `json:"holidays"`
	RoleRequirements []models.RoleRequirement  `json:"role_requirements"`
	Employees       []BatchEmployee            `json:"employees"`
	MinRestHours    float64                    `json:"min_rest_hours"`
	PreviousErrors  []string                   `json:"previous_errors"`
}

type BatchEmployee struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	RoleID     uint   `json:"role_id"`
	RoleName   string `json:"role_name"`
	LeaveQuota int    `json:"leave_quota"`
}

type BatchResponse struct {
	Shifts []BatchShift `json:"shifts"`
	Leaves []BatchLeave `json:"leaves"`
	BatchInfo struct {
		BatchIndex int `json:"batch_index"`
		ShiftCount int `json:"shift_count"`
		LeaveCount int `json:"leave_count"`
	} `json:"batch_info"`
}

type BatchShift struct {
	EmployeeID      uint   `json:"employee_id"`
	Date            string `json:"date"`
	ShiftTemplateID uint   `json:"shift_template_id"`
}

type BatchLeave struct {
	EmployeeID uint   `json:"employee_id"`
	Date       string `json:"date"`
}

func NewGenerator(provider, apiURL, apiKey, model string, batchSize int, minRestHours float64, maxRetries int) *Generator {
	return &Generator{
		provider:     provider,
		apiURL:       apiURL,
		apiKey:       apiKey,
		model:        model,
		batchSize:    batchSize,
		minRestHours: minRestHours,
		maxRetries:   maxRetries,
		client:       &http.Client{Timeout: 60 * time.Second},
	}
}

// Generate runs the full batch generation process.
func (g *Generator) Generate(input BatchInput, allEmployees []models.Employee, quotas []models.EmployeeLeaveQuota) ([]models.ScheduleShift, []models.ScheduleEmployeeLeave, []models.ValidationViolation, error) {
	var allShifts []models.ScheduleShift
	var allLeaves []models.ScheduleEmployeeLeave
	var allErrors []models.ValidationViolation

	quotaMap := make(map[uint]int)
	for _, q := range quotas {
		quotaMap[q.EmployeeID] = q.QuotaDays
	}

	// Split into batches
	var batches []BatchInput
	for i := 0; i < len(allEmployees); i += g.batchSize {
		end := i + g.batchSize
		if end > len(allEmployees) {
			end = len(allEmployees)
		}
		batch := input
		batch.BatchIndex = len(batches)
		batch.TotalBatches = ((len(allEmployees) + g.batchSize - 1) / g.batchSize)

		batch.Employees = make([]BatchEmployee, 0, end-i)
		for _, emp := range allEmployees[i:end] {
			batch.Employees = append(batch.Employees, BatchEmployee{
				ID:         emp.ID,
				Name:       emp.Name,
				RoleID:     emp.RoleID,
				RoleName:   emp.Role.Name,
				LeaveQuota: quotaMap[emp.ID],
			})
		}
		batches = append(batches, batch)
	}

	if len(batches) == 0 {
		// Single batch
		batch := input
		batch.BatchIndex = 0
		batch.TotalBatches = 1
		batch.Employees = make([]BatchEmployee, 0, len(allEmployees))
		for _, emp := range allEmployees {
			batch.Employees = append(batch.Employees, BatchEmployee{
				ID:         emp.ID,
				Name:       emp.Name,
				RoleID:     emp.RoleID,
				RoleName:   emp.Role.Name,
				LeaveQuota: quotaMap[emp.ID],
			})
		}
		batches = append(batches, batch)
	}

	for _, batch := range batches {
		resp, err := g.generateBatch(batch)
		if err != nil {
			allErrors = append(allErrors, models.ValidationViolation{
				Type:     "ai_generation_error",
				Severity: "error",
				Message:  fmt.Sprintf("Batch %d gagal: %v", batch.BatchIndex, err),
			})
			continue
		}

		for _, s := range resp.Shifts {
			allShifts = append(allShifts, models.ScheduleShift{
				EmployeeID:      s.EmployeeID,
				ShiftTemplateID: s.ShiftTemplateID,
				Date:            s.Date,
			})
		}
		for _, l := range resp.Leaves {
			allLeaves = append(allLeaves, models.ScheduleEmployeeLeave{
				EmployeeID: l.EmployeeID,
				Date:       l.Date,
			})
		}
	}

	return allShifts, allLeaves, allErrors, nil
}

func (g *Generator) generateBatch(input BatchInput) (*BatchResponse, error) {
	switch g.provider {
	case "openai":
		return g.callOpenAI(input)
	default:
		return g.mockGenerate(input)
	}
}

// ── Mock Generator ─────────────────────────────────────────
func (g *Generator) mockGenerate(input BatchInput) (*BatchResponse, error) {
	log.Printf("mock generating batch %d/%d with %d employees", input.BatchIndex+1, input.TotalBatches, len(input.Employees))

	start, _ := time.Parse("2006-01-02", input.PeriodStart)
	end, _ := time.Parse("2006-01-02", input.PeriodEnd)

	var resp BatchResponse
	resp.BatchInfo.BatchIndex = input.BatchIndex

	holidaySet := make(map[string]bool)
	for _, h := range input.Holidays {
		holidaySet[h.Date] = true
	}

	leaveDaysPerWeek := 2
	if input.RandomDaysPerWeek != nil {
		leaveDaysPerWeek = *input.RandomDaysPerWeek
	}

	shiftTemplates := input.ShiftTemplates
	capacityIndex := 0

	for _, emp := range input.Employees {
		weekOffset := 0
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			dow := int(d.Weekday())

			isOff := false
			if input.LeaveMode == "fixed" {
				for _, fd := range input.FixedLeaves {
					if dow == fd {
						isOff = true
						break
					}
				}
				if !isOff && holidaySet[dateStr] {
					isOff = true
				}
			} else {
				// Random: deterministik dari hash employee+week
				r := rand.New(rand.NewSource(int64(emp.ID)*1000 + int64(d.Year())*100 + int64(d.Month()) + int64(weekOffset)))
				offDays := make(map[int]bool)
				for len(offDays) < leaveDaysPerWeek {
					offDays[r.Intn(7)] = true
				}
				if offDays[dow] {
					isOff = true
				}
			}

			if isOff {
				if d.Weekday() != time.Sunday || input.LeaveMode == "random" {
					resp.Leaves = append(resp.Leaves, BatchLeave{
						EmployeeID: emp.ID,
						Date:       dateStr,
					})
				}
				continue
			}

			t := shiftTemplates[capacityIndex%len(shiftTemplates)]
			capacityIndex++

			resp.Shifts = append(resp.Shifts, BatchShift{
				EmployeeID:      emp.ID,
				Date:            dateStr,
				ShiftTemplateID: t.ID,
			})
		}
		weekOffset++
	}

	resp.BatchInfo.ShiftCount = len(resp.Shifts)
	resp.BatchInfo.LeaveCount = len(resp.Leaves)

	return &resp, nil
}

// ── OpenAI Generator ────────────────────────────────────────
func (g *Generator) callOpenAI(input BatchInput) (*BatchResponse, error) {
	// Build prompt
	prompt := g.buildPrompt(input)

	payload := map[string]interface{}{
		"model": g.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Kamu adalah asisten penjadwal shift. Output hanya JSON array, tanpa teks lain.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.3,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", g.apiURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if g.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	var batchResp BatchResponse
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &batchResp); err != nil {
		return nil, fmt.Errorf("failed to parse openai JSON: %w", err)
	}

	return &batchResp, nil
}

func (g *Generator) buildPrompt(input BatchInput) string {
	prompt := fmt.Sprintf(`---
batch_info:
  batch_index: %d
  total_batches: %d
  period_start: '%s'
  period_end: '%s'

shift_templates:
`, input.BatchIndex+1, input.TotalBatches, input.PeriodStart, input.PeriodEnd)

	for _, t := range input.ShiftTemplates {
		prompt += fmt.Sprintf(`  - id: %d
    name: %s
    start_time: '%s'
    end_time: '%s'
    is_cross_day: %v
`, t.ID, t.Name, t.StartTime, t.EndTime, t.IsCrossDay)
	}

	prompt += fmt.Sprintf(`
leave_mode: %s
min_rest_hours: %.0f
`, input.LeaveMode, input.MinRestHours)

	if input.LeaveMode == "fixed" {
		prompt += "fixed_leaves:\n"
		for _, fd := range input.FixedLeaves {
			prompt += fmt.Sprintf("  - %d\n", fd)
		}
	}

	if input.RandomDaysPerWeek != nil {
		prompt += fmt.Sprintf("random_days_per_week: %d\n", *input.RandomDaysPerWeek)
	}

	if len(input.Holidays) > 0 {
		prompt += "\nholidays:\n"
		for _, h := range input.Holidays {
			prompt += fmt.Sprintf("  - date: %s, name: %s\n", h.Date, h.Name)
		}
	}

	prompt += "\nrole_requirements:\n"
	for _, r := range input.RoleRequirements {
		prompt += fmt.Sprintf("  - shift_template_id: %d, role_id: %d, min_count: %d\n", r.ShiftTemplateID, r.RoleID, r.MinCount)
	}

	prompt += "\nemployees:\n"
	for _, e := range input.Employees {
		prompt += fmt.Sprintf("  - id: %d, name: %s, role_id: %d, role_name: %s, leave_quota: %d\n", e.ID, e.Name, e.RoleID, e.RoleName, e.LeaveQuota)
	}

	if len(input.PreviousErrors) > 0 {
		prompt += "\nprevious_errors:\n"
		for _, err := range input.PreviousErrors {
			prompt += fmt.Sprintf("  - \"%s\"\n", err)
		}
	}

	prompt += `
OUTPUT FORMAT (JSON):
{
  "shifts": [
    {"employee_id": 1, "date": "2026-07-01", "shift_template_id": 1}
  ],
  "leaves": [
    {"employee_id": 1, "date": "2026-07-02"}
  ],
  "batch_info": {
    "batch_index": 1,
    "shift_count": 25,
    "leave_count": 10
  }
}
`

	return prompt
}

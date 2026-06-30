package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ademaswahyu/autoshift-backend/models"
)

type Validator struct {
	MinRestHours float64
}

func NewValidator(minRestHours float64) *Validator {
	return &Validator{MinRestHours: minRestHours}
}

func parseTime(date, timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", date, timeStr))
}

// ValidateSchedule runs all constraint checks.
func (v *Validator) ValidateSchedule(scheduleID uint, shifts []models.ScheduleShift, leaves []models.ScheduleEmployeeLeave, shiftTemplates []models.ShiftTemplate, employees []models.Employee, roleReqs []models.RoleRequirement, leaveQuotas []models.EmployeeLeaveQuota) models.ValidationResult {
	var violations []models.ValidationViolation

	violations = append(violations, v.CheckRestTime(shifts, shiftTemplates)...)
	violations = append(violations, v.CheckRoleComposition(shifts, employees, shiftTemplates, roleReqs)...)
	violations = append(violations, v.CheckDoubleBooking(scheduleID, shifts, leaves)...)
	violations = append(violations, v.CheckLeaveQuota(shifts, leaves, leaveQuotas)...)

	isValid := true
	for _, vi := range violations {
		if vi.Severity == "error" {
			isValid = false
			break
		}
	}

	return models.ValidationResult{IsValid: isValid, Violations: violations}
}

// CheckRestTime: minimal jeda antar shift di hari berurutan.
func (v *Validator) CheckRestTime(shifts []models.ScheduleShift, templates []models.ShiftTemplate) []models.ValidationViolation {
	var violations []models.ValidationViolation
	tmplMap := make(map[uint]models.ShiftTemplate)
	for _, t := range templates {
		tmplMap[t.ID] = t
	}

	empShiftMap := make(map[uint]map[string]models.ScheduleShift)
	for _, s := range shifts {
		if empShiftMap[s.EmployeeID] == nil {
			empShiftMap[s.EmployeeID] = make(map[string]models.ScheduleShift)
		}
		empShiftMap[s.EmployeeID][s.Date] = s
	}

	for empID, days := range empShiftMap {
		var dates []string
		for d := range days {
			dates = append(dates, d)
		}
		sort.Strings(dates)

		for i := 1; i < len(dates); i++ {
			prev := days[dates[i-1]]
			curr := days[dates[i]]

			prevTmpl, ok1 := tmplMap[prev.ShiftTemplateID]
			currTmpl, ok2 := tmplMap[curr.ShiftTemplateID]
			if !ok1 || !ok2 {
				continue
			}

			prevEnd, err1 := parseTime(prev.Date, prevTmpl.EndTime)
			currStart, err2 := parseTime(curr.Date, currTmpl.StartTime)
			if err1 != nil || err2 != nil {
				continue
			}

			if prevTmpl.IsCrossDay {
				prevEnd = prevEnd.Add(24 * time.Hour)
			}

			rest := currStart.Sub(prevEnd)
			if rest < time.Duration(v.MinRestHours)*time.Hour {
				violations = append(violations, models.ValidationViolation{
					Type:       "rest_time_violation",
					Severity:   "error",
					Message:    fmt.Sprintf("Jeda hanya %.0f jam (min %.0f jam)", rest.Hours(), v.MinRestHours),
					EmployeeID: empID,
					Date:       curr.Date,
				})
			}
		}
	}
	return violations
}

// CheckRoleComposition: pastikan setiap shift punya role requirement terpenuhi.
func (v *Validator) CheckRoleComposition(shifts []models.ScheduleShift, employees []models.Employee, templates []models.ShiftTemplate, roleReqs []models.RoleRequirement) []models.ValidationViolation {
	var violations []models.ValidationViolation

	empMap := make(map[uint]models.Employee)
	for _, e := range employees {
		empMap[e.ID] = e
	}

	reqMap := make(map[uint]map[uint]int) // shiftTemplateID -> roleID -> minCount
	for _, r := range roleReqs {
		if reqMap[r.ShiftTemplateID] == nil {
			reqMap[r.ShiftTemplateID] = make(map[uint]int)
		}
		reqMap[r.ShiftTemplateID][r.RoleID] = r.MinCount
	}

	type shiftKey struct {
		Date            string
		ShiftTemplateID uint
	}

	composition := make(map[shiftKey]map[uint]int)
	for _, s := range shifts {
		key := shiftKey{Date: s.Date, ShiftTemplateID: s.ShiftTemplateID}
		if composition[key] == nil {
			composition[key] = make(map[uint]int)
		}
		if emp, ok := empMap[s.EmployeeID]; ok {
			composition[key][emp.RoleID]++
		}
	}

	for key, roleCounts := range composition {
		if reqs, ok := reqMap[key.ShiftTemplateID]; ok {
			for roleID, minCount := range reqs {
				if roleCounts[roleID] < minCount {
					violations = append(violations, models.ValidationViolation{
						Type:            "role_composition",
						Severity:        "error",
						Message:         fmt.Sprintf("Kekurangan role (butuh %d, ada %d)", minCount, roleCounts[roleID]),
						Date:            key.Date,
						ShiftTemplateID: key.ShiftTemplateID,
					})
				}
			}
		}
	}
	return violations
}

// CheckDoubleBooking: pastikan tidak ada karyawan yang shift dan libur di hari sama.
func (v *Validator) CheckDoubleBooking(scheduleID uint, shifts []models.ScheduleShift, leaves []models.ScheduleEmployeeLeave) []models.ValidationViolation {
	var violations []models.ValidationViolation
	leaveSet := make(map[string]bool)
	for _, l := range leaves {
		leaveSet[fmt.Sprintf("%d-%s", l.EmployeeID, l.Date)] = true
	}
	for _, s := range shifts {
		key := fmt.Sprintf("%d-%s", s.EmployeeID, s.Date)
		if leaveSet[key] {
			violations = append(violations, models.ValidationViolation{
				Type:       "double_booking",
				Severity:   "error",
				Message:    "Terdaftar shift dan libur di hari yang sama",
				EmployeeID: s.EmployeeID,
				Date:       s.Date,
			})
		}
	}
	return violations
}

// CheckLeaveQuota: pastikan kuota libur tidak terlampaui.
func (v *Validator) CheckLeaveQuota(shifts []models.ScheduleShift, leaves []models.ScheduleEmployeeLeave, quotas []models.EmployeeLeaveQuota) []models.ValidationViolation {
	var violations []models.ValidationViolation
	quotaMap := make(map[uint]int)
	for _, q := range quotas {
		quotaMap[q.EmployeeID] = q.QuotaDays
	}

	leaveCount := make(map[uint]int)
	for _, l := range leaves {
		leaveCount[l.EmployeeID]++
	}

	for empID, count := range leaveCount {
		if quota, ok := quotaMap[empID]; ok && count > quota {
			violations = append(violations, models.ValidationViolation{
				Type:       "leave_quota_exceeded",
				Severity:   "warning",
				Message:    fmt.Sprintf("Kuota libur %d hari, sekarang %d hari", quota, count),
				EmployeeID: empID,
				Quota:      quota,
				Actual:     count,
			})
		}
	}
	return violations
}

// CalculateFairnessMetrics: hitung SD jam kerja dan shift weekend.
func CalculateFairnessMetrics(shifts []models.ScheduleShift, templates []models.ShiftTemplate, employees []models.Employee) *models.FairnessMetrics {
	if len(employees) == 0 {
		return nil
	}

	tmplMap := make(map[uint]models.ShiftTemplate)
	for _, t := range templates {
		tmplMap[t.ID] = t
	}

	totalHours := make(map[uint]float64)
	weekendShifts := make(map[uint]int)

	for _, s := range shifts {
		t, ok := tmplMap[s.ShiftTemplateID]
		if !ok {
			continue
		}
		start, _ := time.Parse("15:04", t.StartTime)
		end, _ := time.Parse("15:04", t.EndTime)
		hours := end.Sub(start).Hours()
		if hours <= 0 {
			hours += 24
		}
		totalHours[s.EmployeeID] += hours

		date, err := time.Parse("2006-01-02", s.Date)
		if err == nil && (date.Weekday() == time.Saturday || date.Weekday() == time.Sunday) {
			weekendShifts[s.EmployeeID]++
		}
	}

	var hoursList []float64
	var weekendList []float64
	var minW, maxW int
	firstW := true

	for _, emp := range employees {
		if emp.IsActive {
			h := totalHours[emp.ID]
			hoursList = append(hoursList, h)

			w := float64(weekendShifts[emp.ID])
			weekendList = append(weekendList, w)
			if firstW {
				minW, maxW = weekendShifts[emp.ID], weekendShifts[emp.ID]
				firstW = false
			} else {
				if weekendShifts[emp.ID] < minW {
					minW = weekendShifts[emp.ID]
				}
				if weekendShifts[emp.ID] > maxW {
					maxW = weekendShifts[emp.ID]
				}
			}
		}
	}

	if len(hoursList) == 0 {
		return nil
	}

	avg := mean(hoursList)
	var sumSq float64
	for _, h := range hoursList {
		sumSq += (h - avg) * (h - avg)
	}
	sdHours := math.Sqrt(sumSq / float64(len(hoursList)))

	avgW := mean(weekendList)
	var sumSqW float64
	for _, w := range weekendList {
		sumSqW += (w - avgW) * (w - avgW)
	}
	sdW := math.Sqrt(sumSqW / float64(len(weekendList)))

	return &models.FairnessMetrics{
		TotalEmployees:      len(employees),
		AvgShiftHours:       math.Round(avg*10) / 10,
		StdDevHours:         math.Round(sdHours*10) / 10,
		StdDevWeekendShifts: math.Round(sdW*10) / 10,
		WeekendShiftsPerEmployee: models.WeekendShiftStats{
			Min: minW,
			Max: maxW,
			Avg: math.Round(avgW*10) / 10,
		},
	}
}

func mean(vals []float64) float64 {
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

export interface ShiftTemplate {
  id: number
  name: string
  start_time: string
  end_time: string
  is_cross_day: boolean
  color_hex: string
  capacity: number
}

export interface EmployeeRole {
  id: number
  name: string
  level: number
}

export interface Employee {
  id: number
  tenant_id: number
  name: string
  role_id: number
  role_name: string
  email: string
  phone: string
  is_active: boolean
  created_at: string
  role?: EmployeeRole
}

export interface CreateEmployeeRequest {
  name: string
  role_id: number
  email?: string
  phone?: string
}

export interface UpdateEmployeeRequest {
  name?: string
  role_id?: number
  email?: string
  phone?: string
  is_active?: boolean
}

export interface RoleRequirement {
  shift_template_id: number
  role_id: number
  min_count: number
}

export interface FixedLeave {
  day_of_week: number
}

export interface RandomLeaveConfig {
  days_per_week: number
}

export interface ScheduleConfig {
  tenant_id: number
  month: number
  year: number
  leave_mode: 'fixed' | 'random'
  fixed_leaves: number[] | null
  random_days_per_week: number | null
  shift_templates: { id: number; capacity: number }[]
  role_requirements: RoleRequirement[]
  employee_ids: number[]
}

export interface ScheduleShift {
  id: number
  employee_id: number
  employee_name: string
  employee_role: string
  shift_template_id: number
  shift_name: string
  date: string
  is_override: boolean
}

export interface ScheduleLeave {
  id: number
  employee_id: number
  employee_name: string
  date: string
  is_override: boolean
}

export interface Holiday {
  date: string
  name: string
  is_national: boolean
}

export interface ValidationViolation {
  type: 'rest_time_violation' | 'role_composition' | 'leave_quota_exceeded' | 'double_booking'
  severity: 'error' | 'warning'
  message: string
  employee_id: number
  employee_name?: string
  date?: string
  shift_template_id?: number
}

export interface FairnessMetrics {
  total_employees: number
  avg_shift_hours: number
  std_dev_hours: number
  std_dev_weekend_shifts: number
  weekend_shifts_per_employee: {
    min: number
    max: number
    avg: number
  }
}

export interface GenerationSummary {
  total_batches: number
  completed_batches: number
  failed_batches: number
  total_shifts_generated: number
  total_leaves_generated: number
  total_errors: number
}

export interface ScheduleDetail {
  schedule: {
    id: number
    month: number
    year: number
    leave_mode: 'fixed' | 'random'
    status: 'draft' | 'published'
    notes: string | null
    created_at: string
    published_at: string | null
  }
  shift_templates: ShiftTemplate[]
  shifts: ScheduleShift[]
  leaves: ScheduleLeave[]
  holidays: Holiday[]
  generation_errors: ValidationViolation[]
  fairness_metrics: FairnessMetrics
  generation_summary: GenerationSummary
}

export type DragItem = {
  type: 'shift' | 'leave'
  schedule_shift_id?: number
  schedule_leave_id?: number
  employee_id: number
  shift_template_id?: number
  date: string
}

export type CalendarCell = {
  date: string
  dayOfWeek: number
  isHoliday: boolean
  holidayName?: string
  isToday: boolean
  isWeekend: boolean
}

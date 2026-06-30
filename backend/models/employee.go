package models

import "time"

type EmployeeRole struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TenantID uint   `gorm:"not null;uniqueIndex:idx_role_tenant" json:"tenant_id"`
	Name     string `gorm:"size:100;not null;uniqueIndex:idx_role_tenant" json:"name"`
	Level    int    `gorm:"default:0" json:"level"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

func (EmployeeRole) TableName() string { return "employee_roles" }

type Employee struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"not null;index" json:"tenant_id"`
	RoleID    uint      `gorm:"not null" json:"role_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255" json:"email"`
	Phone     string    `gorm:"size:50" json:"phone"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`

	Role EmployeeRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (Employee) TableName() string { return "employees" }

type EmployeeLeaveQuota struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	TenantID   uint `gorm:"not null;index" json:"tenant_id"`
	EmployeeID uint `gorm:"not null;uniqueIndex:idx_quota_employee_month" json:"employee_id"`
	Month      int  `gorm:"not null;uniqueIndex:idx_quota_employee_month" json:"month"`
	Year       int  `gorm:"not null;uniqueIndex:idx_quota_employee_month" json:"year"`
	QuotaDays  int  `gorm:"default:0" json:"quota_days"`
}

func (EmployeeLeaveQuota) TableName() string { return "employee_leave_quotas" }

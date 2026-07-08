package handlers

import (
	"github.com/ademaswahyu/autoshift-backend/models"
	"github.com/gofiber/fiber/v2"
)

type EmployeeHandler struct{}

func NewEmployeeHandler() *EmployeeHandler {
	return &EmployeeHandler{}
}

type CreateEmployeeRequest struct {
	Name   string `json:"name"`
	RoleID uint   `json:"role_id"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
}

type UpdateEmployeeRequest struct {
	Name     string `json:"name,omitempty"`
	RoleID   *uint  `json:"role_id,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// ── GET /employees ─────────────────────────────────────────

func (h *EmployeeHandler) List(c *fiber.Ctx) error {
	tenantID := extractTenantID(c)
	if tenantID == 0 {
		tenantID = 1 // default tenant when no auth
	}

	query := models.DB.Preload("Role").Where("tenant_id = ?", tenantID)

	if isActive := c.Query("is_active"); isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}
	if roleID := c.Query("role_id"); roleID != "" {
		query = query.Where("role_id = ?", roleID)
	}

	var employees []models.Employee
	if err := query.Order("name asc").Find(&employees).Error; err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "failed to fetch employees",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    employees,
	})
}

// ── GET /employees/:id ──────────────────────────────────────

func (h *EmployeeHandler) Get(c *fiber.Ctx) error {
	tenantID := extractTenantID(c)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid id",
		})
	}

	var employee models.Employee
	if result := models.DB.Preload("Role").Where("tenant_id = ? AND id = ?", tenantID, id).First(&employee); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "employee not found",
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    employee,
	})
}

// ── POST /employees ─────────────────────────────────────────

func (h *EmployeeHandler) Create(c *fiber.Ctx) error {
	tenantID := extractTenantID(c)
	if tenantID == 0 {
		tenantID = 1 // default tenant when no auth
	}

	var req CreateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(422).JSON(models.APIResponse{
			Success: false,
			Error:   "name is required",
		})
	}
	if req.RoleID == 0 {
		return c.Status(422).JSON(models.APIResponse{
			Success: false,
			Error:   "role_id is required",
		})
	}

	// Validate role exists
	var role models.EmployeeRole
	if result := models.DB.Where("tenant_id = ? AND id = ?", tenantID, req.RoleID).First(&role); result.Error != nil {
		return c.Status(422).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid role_id",
		})
	}

	employee := models.Employee{
		TenantID: tenantID,
		RoleID:   req.RoleID,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		IsActive: true,
	}

	if err := models.DB.Create(&employee).Error; err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "failed to create employee: " + err.Error(),
		})
	}

	// Reload with role
	models.DB.Preload("Role").First(&employee, employee.ID)

	return c.Status(201).JSON(models.APIResponse{
		Success: true,
		Data:    employee,
	})
}

// ── PUT /employees/:id ─────────────────────────────────────

func (h *EmployeeHandler) Update(c *fiber.Ctx) error {
	tenantID := extractTenantID(c)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid id",
		})
	}

	var employee models.Employee
	if result := models.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&employee); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "employee not found",
		})
	}

	var req UpdateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.RoleID != nil {
		// Validate role exists
		var role models.EmployeeRole
		if result := models.DB.Where("tenant_id = ? AND id = ?", tenantID, *req.RoleID).First(&role); result.Error != nil {
			return c.Status(422).JSON(models.APIResponse{
				Success: false,
				Error:   "invalid role_id",
			})
		}
		updates["role_id"] = *req.RoleID
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := models.DB.Model(&employee).Updates(updates).Error; err != nil {
			return c.Status(500).JSON(models.APIResponse{
				Success: false,
				Error:   "failed to update employee: " + err.Error(),
			})
		}
	}

	// Reload with role
	models.DB.Preload("Role").First(&employee, employee.ID)

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    employee,
	})
}

// ── DELETE /employees/:id ───────────────────────────────────

func (h *EmployeeHandler) Delete(c *fiber.Ctx) error {
	tenantID := extractTenantID(c)
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Success: false,
			Error:   "invalid id",
		})
	}

	var employee models.Employee
	if result := models.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&employee); result.Error != nil {
		return c.Status(404).JSON(models.APIResponse{
			Success: false,
			Error:   "employee not found",
		})
	}

	// Soft delete - set is_active to false
	if err := models.DB.Model(&employee).Update("is_active", false).Error; err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Error:   "failed to delete employee: " + err.Error(),
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "employee deactivated",
	})
}

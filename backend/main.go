package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ademaswahyu/autoshift-backend/ai"
	"github.com/ademaswahyu/autoshift-backend/config"
	"github.com/ademaswahyu/autoshift-backend/handlers"
	"github.com/ademaswahyu/autoshift-backend/middleware"
	"github.com/ademaswahyu/autoshift-backend/models"
	"github.com/ademaswahyu/autoshift-backend/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.Load()

	// ── Database ────────────────────────────────────────────
	models.Connect(cfg.Dialector())
	models.Migrate()

	// Seed default data if empty
	seedDefaultData()

	// ── Services ─────────────────────────────────────────────
	validator := services.NewValidator(cfg.MinRestHours)
	generator := ai.NewGenerator(
		cfg.AIProvider,
		cfg.AIAPIURL,
		cfg.AIAPIKey,
		cfg.AIModel,
		cfg.BatchSize,
		cfg.MinRestHours,
		cfg.MaxRetries,
	)
	schedulerEngine := services.NewSchedulerEngine(validator, generator)
	holidayFetcher := services.NewHolidayFetcher(cfg.HolidayAPIURL)

	// ── Handlers ────────────────────────────────────────────
	scheduleHandler := handlers.NewScheduleHandler(schedulerEngine)
	exportHandler := handlers.NewExportHandler()

	// ── Fiber App ────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName: "autoShift API",
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${method} ${path} ${status} ${latency}\n",
	}))

	// ── Public Routes ────────────────────────────────────────
	api := app.Group("/api/v1")

	// Health check (MUST be first, before any auth middleware)
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	api.Post("/login", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "login endpoint"})
	})

	api.Get("/holidays", func(c *fiber.Ctx) error {
		year := c.QueryInt("year", 2026)
		holidays := holidayFetcher.GetHolidays(year)
		return c.JSON(models.APIResponse{
			Success: true,
			Data:    holidays,
		})
	})

	api.Post("/holidays/fetch", func(c *fiber.Ctx) error {
		year := c.QueryInt("year", 2026)
		country := c.Query("country", "ID")
		if err := holidayFetcher.FetchAndStore(year, country); err != nil {
			return c.Status(500).JSON(models.APIResponse{
				Success: false,
				Error:   err.Error(),
			})
		}
		return c.JSON(models.APIResponse{
			Success: true,
			Message: "Holidays fetched",
		})
	})

	// ── Protected Routes ────────────────────────────────────
	protected := api.Group("/schedules", middleware.Auth())

	protected.Post("/", scheduleHandler.Create)
	protected.Get("/:id", scheduleHandler.Get)
	protected.Get("/:id/validate", scheduleHandler.Validate)
	protected.Put("/:id/shifts", scheduleHandler.UpdateShifts)
	protected.Put("/:id/publish", scheduleHandler.Publish)
	protected.Post("/:id/regenerate", scheduleHandler.Regenerate)
	protected.Get("/:id/export", exportHandler.Export)
	protected.Get("/:id/share", exportHandler.Share)

	// ── Start ────────────────────────────────────────────────
	port := cfg.ServerPort
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// ── Graceful Shutdown ────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("autoShift API starting on :%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	sig := <-quit
	log.Printf("received signal %v, shutting down...", sig)
	if err := app.Shutdown(); err != nil {
		log.Printf("forced shutdown error: %v", err)
	}
	log.Println("server stopped gracefully")
}

func findOrCreateRole(tenantID uint, name string, level int) models.EmployeeRole {
	var r models.EmployeeRole
	models.DB.Where("tenant_id = ? AND name = ?", tenantID, name).First(&r)
	if r.ID == 0 {
		r = models.EmployeeRole{Name: name, Level: level, TenantID: tenantID}
		models.DB.Create(&r)
	}
	return r
}

func findOrCreateShift(tenantID uint, name, start, end, color string, cross bool) models.ShiftTemplate {
	var t models.ShiftTemplate
	models.DB.Where("tenant_id = ? AND name = ?", tenantID, name).First(&t)
	if t.ID == 0 {
		t = models.ShiftTemplate{
			Name: name, StartTime: start, EndTime: end,
			IsCrossDay: cross, ColorHex: color, TenantID: tenantID,
		}
		models.DB.Create(&t)
	}
	return t
}

func findOrCreateEmployee(tenantID, roleID uint, name string) models.Employee {
	var e models.Employee
	models.DB.Where("tenant_id = ? AND name = ?", tenantID, name).First(&e)
	if e.ID == 0 {
		e = models.Employee{Name: name, RoleID: roleID, TenantID: tenantID}
		models.DB.Create(&e)
	}
	return e
}

func seedDefaultData() {
	var tenant models.Tenant
	models.DB.Where("name = ?", "Default Company").First(&tenant)
	if tenant.ID == 0 {
		tenant = models.Tenant{Name: "Default Company"}
		models.DB.Create(&tenant)
	}

	supervisor := findOrCreateRole(tenant.ID, "Supervisor", 2)
	staff := findOrCreateRole(tenant.ID, "Staff", 1)
	findOrCreateRole(tenant.ID, "Intern", 0)

	pagi := findOrCreateShift(tenant.ID, "Pagi", "08:00", "16:00", "#3B82F6", false)
	siang := findOrCreateShift(tenant.ID, "Siang", "16:00", "00:00", "#F59E0B", true)
	malam := findOrCreateShift(tenant.ID, "Malam", "22:00", "06:00", "#6366F1", true)

	for _, t := range []models.ShiftTemplate{pagi, siang, malam} {
		var cnt int64
		models.DB.Model(&models.RoleRequirement{}).
			Where("shift_template_id = ? AND role_id = ?", t.ID, supervisor.ID).
			Count(&cnt)
		if cnt == 0 {
			models.DB.Create(&models.RoleRequirement{
				ShiftTemplateID: t.ID, RoleID: supervisor.ID, MinCount: 1,
			})
		}
	}

	findOrCreateEmployee(tenant.ID, supervisor.ID, "Budi")
	findOrCreateEmployee(tenant.ID, staff.ID, "Siti")
	findOrCreateEmployee(tenant.ID, staff.ID, "Andi")
	findOrCreateEmployee(tenant.ID, staff.ID, "Dewi")
	findOrCreateEmployee(tenant.ID, supervisor.ID, "Rudi")

	log.Println("default data seeded")
}

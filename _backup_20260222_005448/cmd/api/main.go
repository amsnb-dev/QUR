package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quran-school/api/internal/auth"
	"github.com/quran-school/api/internal/config"
	"github.com/quran-school/api/internal/db"
	"github.com/quran-school/api/internal/handler"
	"github.com/quran-school/api/internal/middleware"
	"github.com/quran-school/api/internal/migrate"
	"github.com/quran-school/api/internal/model"
)

func main() {
	// ── Config ────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Database ──────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()
	log.Println("✅ database connected")

	// ── Migrations (opt-in: RUN_MIGRATIONS=true) ──────────────
	if cfg.RunMigrations {
		log.Printf("running migrations from %s ...", cfg.MigrationsDir)
		if err := migrate.Run(ctx, pool.Pool, cfg.MigrationsDir); err != nil {
			log.Fatalf("migrations: %v", err)
		}
		log.Println("✅ migrations applied")
	} else {
		log.Println("⏭  migrations skipped (set RUN_MIGRATIONS=true to enable)")
	}

	// ── Services + Handlers ───────────────────────────────────
	authSvc := auth.NewService(cfg)
	authH := handler.NewAuthHandler(cfg, pool, authSvc)
	groupH := handler.NewGroupHandler()
	studentH := handler.NewStudentHandler()
	attendH := handler.NewAttendanceHandler()
	memH := handler.NewMemorizationHandler()
        teacherH := handler.NewTeacherHandler()

	// ── Router ────────────────────────────────────────────────
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
// CORS
r.Use(func(c *gin.Context) {
c.Header("Access-Control-Allow-Origin", "*")
c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
if c.Request.Method == "OPTIONS" {
c.AbortWithStatus(204)
return
}
c.Next()
})

	// ── Health ────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "db_down",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ── Public auth ───────────────────────────────────────────
	public := r.Group("/auth")
	{
		public.POST("/login", authH.Login)
		public.POST("/refresh", authH.Refresh)
	}

	// All protected routes: RequireAuth opens tenant-scoped tx per request.
	authed := r.Group("/", middleware.RequireAuth(authSvc, pool))
	{
		authed.POST("/auth/logout", authH.Logout)
		authed.GET("/me", authH.Me)

		// ── Groups ────────────────────────────────────────────
		// RBAC: accountant → 403 everywhere
		//       teacher   → GET only
		//       admin     → full CRUD
		grp := authed.Group("/groups", middleware.RequireNotAccountant())
		{
			grp.GET("", groupH.List)
			grp.GET("/:id", groupH.Get)

			// Attendance sub-resource lives under groups
			// supervisor/school_admin/super_admin: write
			// teacher: write only on own group (checked inside handler)
			grp.POST("/:id/attendance", attendH.BulkUpsert)
			grp.GET("/:id/attendance", attendH.List)

			adminWrite := grp.Group("", middleware.RequireWrite())
			adminWrite.POST("", groupH.Create)
			adminWrite.PUT("/:id", groupH.Update)
			adminWrite.PATCH("/:id/archive", groupH.Archive)
		}

		// ── Students ──────────────────────────────────────────
		stu := authed.Group("/students", middleware.RequireNotAccountant())
		{
			stu.GET("", studentH.List)
			stu.GET("/:id", studentH.Get)
			stu.GET("/:id/groups", studentH.ListGroups)
			stu.GET("/:id/memorization", memH.List)

			// teacher + admin can record memorization
			stu.POST("/:id/memorization",
				middleware.RequireRole(
					model.RoleSuperAdmin, model.RoleSchoolAdmin,
					model.RoleSupervisor, model.RoleTeacher,
				),
				memH.Create,
			)

			adminWrite := stu.Group("", middleware.RequireWrite())
			adminWrite.POST("", studentH.Create)
			adminWrite.PUT("/:id", studentH.Update)
			adminWrite.PATCH("/:id/archive", studentH.Archive)
			adminWrite.POST("/:id/groups", studentH.AddGroup)
		}

		// ── Student-group memberships ─────────────────────────
		sg := authed.Group("/student-groups",
			middleware.RequireNotAccountant(),
			middleware.RequireRole(model.RoleSuperAdmin, model.RoleSchoolAdmin),
		)
		{
			sg.PATCH("/:id/close", studentH.CloseGroup)
		}

		// ── Attendance patch (standalone) ─────────────────────
		// supervisor/school_admin/super_admin/teacher can patch
		att := authed.Group("/attendance",
			middleware.RequireNotAccountant(),
			middleware.RequireRole(
				model.RoleSuperAdmin, model.RoleSchoolAdmin,
				model.RoleSupervisor, model.RoleTeacher,
			),
		)
		{
			att.PATCH("/:id", attendH.Update)
		}

		// ── Memorization patch (standalone) ───────────────────
		mem := authed.Group("/memorization",
			middleware.RequireNotAccountant(),
			middleware.RequireRole(
				model.RoleSuperAdmin, model.RoleSchoolAdmin,
				model.RoleSupervisor, model.RoleTeacher,
			),
		)
		{
			mem.PATCH("/:id", memH.Update)
		}
	}

	// ── Graceful shutdown ─────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bye")
}


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
// -- Config ------------------------------------------------
cfg, err := config.Load()
if err != nil {
log.Fatalf("config: %v", err)
}

baseCtx := context.Background()

// -- Database (runtime) ------------------------------------
dbCtx, dbCancel := context.WithTimeout(baseCtx, 15*time.Second)
defer dbCancel()

pool, err := db.Connect(dbCtx, cfg.DatabaseURL)
if err != nil {
log.Fatalf("db: %v", err)
}
defer pool.Close()
log.Println("? database connected")

// -- Migrations (opt-in: RUN_MIGRATIONS=true) --------------
if cfg.RunMigrations {
migCtx, migCancel := context.WithTimeout(baseCtx, 2*time.Minute)
defer migCancel()

migPool, err := db.Connect(migCtx, cfg.MigrateDatabaseURL)
if err != nil {
log.Fatalf("migrations db: %v", err)
}
defer migPool.Close()

log.Printf("running migrations from %s ...", cfg.MigrationsDir)
if err := migrate.Run(migCtx, migPool.Pool, cfg.MigrationsDir); err != nil {
log.Fatalf("migrations: %v", err)
}
log.Println("? migrations applied")
} else {
log.Println("?  migrations skipped (set RUN_MIGRATIONS=true to enable)")
}

// -- Services + Handlers -----------------------------------
authSvc := auth.NewService(cfg)
authH := handler.NewAuthHandler(cfg, pool, authSvc)
groupH := handler.NewGroupHandler()
studentH := handler.NewStudentHandler()
attendH := handler.NewAttendanceHandler()
memH := handler.NewMemorizationHandler()
teacherH := handler.NewTeacherHandler()
schoolH := handler.NewSchoolHandler()
userH     := handler.NewUserHandler()
academicH := handler.NewAcademicHandler()
subjectH  := handler.NewSubjectHandler()

// -- Router ------------------------------------------------
if cfg.Env == "production" {
gin.SetMode(gin.ReleaseMode)
}

r := gin.New()
r.Use(gin.Logger())
r.Use(gin.Recovery())

// Simple CORS (keep for static HTML frontend)
r.Use(func(c *gin.Context) {
c.Header("Access-Control-Allow-Origin", "*")
c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
if c.Request.Method == http.MethodOptions {
c.AbortWithStatus(http.StatusNoContent)
return
}
c.Next()
})

// -- Health ------------------------------------------------
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

// -- Public auth -------------------------------------------
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

// -- Groups --------------------------------------------
grp := authed.Group("/groups", middleware.RequireNotAccountant())
{
grp.GET("", groupH.List)
grp.GET("/:id", groupH.Get)

grp.POST("/:id/attendance", attendH.BulkUpsert)
grp.GET("/:id/attendance", attendH.List)

adminWrite := grp.Group("", middleware.RequireWrite())
adminWrite.POST("", groupH.Create)
adminWrite.PUT("/:id", groupH.Update)
adminWrite.PATCH("/:id/archive", groupH.Archive)
}

// -- Students ------------------------------------------
stu := authed.Group("/students", middleware.RequireNotAccountant())
{
stu.GET("", studentH.List)
stu.GET("/:id", studentH.Get)
stu.GET("/:id/groups", studentH.ListGroups)
stu.GET("/:id/memorization", memH.List)

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
stu.POST("/:id/subjects", subjectH.AssignToStudent)
stu.GET("/:id/subjects",  subjectH.ListStudentSubjects)
}

// -- Student-group memberships -------------------------
sg := authed.Group("/student-groups",
middleware.RequireNotAccountant(),
middleware.RequireRole(model.RoleSuperAdmin, model.RoleSchoolAdmin),
)
{
sg.PATCH("/:id/close", studentH.CloseGroup)
}

// -- Attendance patch (standalone) ---------------------
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

// -- Memorization patch (standalone) -------------------
mem := authed.Group("/memorization",
middleware.RequireNotAccountant(),
middleware.RequireRole(
model.RoleSuperAdmin, model.RoleSchoolAdmin,
model.RoleSupervisor, model.RoleTeacher,
),
)
{
mem.PATCH("/:id", memH.Update)
// -- Teachers ----------------------------------------------
                tch := authed.Group("/teachers", middleware.RequireWrite())
                {
                        tch.GET("", teacherH.List)
                        tch.GET("/:id", teacherH.Get)
                        tch.POST("", teacherH.Create)
                        tch.PUT("/:id", teacherH.Update)
                        tch.PATCH("/:id/archive", teacherH.Archive)
                }
// -- Schools (super_admin only) ----------------------------
sch := authed.Group("/schools",
    middleware.RequireRole(model.RoleSuperAdmin),

)
{
    sch.GET("",      schoolH.List)
    sch.GET("/:id",  schoolH.Get)
    sch.POST("",     schoolH.Create)
    sch.PUT("/:id",  schoolH.Update)
}
usr := authed.Group("/users", middleware.RequireRole(model.RoleSuperAdmin, model.RoleSchoolAdmin))
{
    usr.GET("", userH.List)
    usr.GET("/:id", userH.Get)
    usr.POST("", userH.Create)
    usr.PUT("/:id", userH.Update)
    usr.PATCH("/:id/archive", userH.Archive)
}
}
}

// -- Settings & Staff ----------------------------------------
settingsH := handler.NewSettingsHandler()
stg := authed.Group("/settings", middleware.RequireWrite())
stg.GET("",          settingsH.GetSettings)
stg.PUT("",          settingsH.UpdateSettings)
stg.GET("/roles",    settingsH.ListRoles)
stg.PUT("/roles/:id/permissions", settingsH.UpdateRolePermissions)
stg.GET("/staff",    settingsH.ListStaff)
stg.POST("/staff",   settingsH.CreateStaff)
stg.PUT("/staff/:id", settingsH.UpdateStaff)

// -- Guardians -----------------------------------------------
guardianH := handler.NewGuardianHandler()
grd := authed.Group("/guardians", middleware.RequireWrite())
grd.GET("",      guardianH.List)
grd.POST("",     guardianH.Create)
grd.GET("/:id",  guardianH.Get)
grd.PUT("/:id",  guardianH.Update)
grd.GET("/:id/students", guardianH.GetStudents)

// -- Academic System ----------------------------------------
ay := authed.Group("/academic-years", middleware.RequireWrite())
{
ay.POST("",    academicH.CreateYear)
ay.GET("",     academicH.ListYears)
ay.GET("/:id", academicH.GetYear)
ay.PUT("/:id", academicH.UpdateYear)
ay.POST("/:id/enrollments", academicH.CreateEnrollment)
ay.GET("/:id/enrollments",  academicH.ListEnrollments)
ay.POST("/:id/exams", academicH.CreateExam)
ay.GET("/:id/exams",  academicH.ListExams)
}
exm := authed.Group("/exams", middleware.RequireWrite())
{
exm.GET("/:id",          academicH.GetExam)
exm.POST("/:id/results", academicH.BulkCreateResults)
exm.GET("/:id/results",  academicH.ListResults)
}
hol := authed.Group("/holidays", middleware.RequireWrite())
{
hol.POST("", academicH.CreateHoliday)
hol.GET("",  academicH.ListHolidays)
}

// -- Subjects -----------------------------------------------
sub := authed.Group("/subjects", middleware.RequireWrite())
{
sub.GET("",     subjectH.List)
sub.POST("",    subjectH.Create)
sub.GET("/:id", subjectH.Get)
sub.PUT("/:id", subjectH.Update)
sub.GET("/:id/levels",  subjectH.ListLevels)
sub.POST("/:id/levels", subjectH.CreateLevel)
}
lev := authed.Group("/levels", middleware.RequireWrite())
{
lev.PUT("/:id", subjectH.UpdateLevel)
}
ssg := authed.Group("/student-subjects", middleware.RequireWrite())
{
ssg.PUT("/:id", subjectH.UpdateStudentSubject)
}
sess := authed.Group("/subject-sessions", middleware.RequireRole(model.RoleSuperAdmin, model.RoleSchoolAdmin, model.RoleSupervisor, model.RoleTeacher))
{
sess.POST("", subjectH.CreateSession)
sess.GET("",  subjectH.ListSessions)
}

// -- Graceful shutdown -------------------------------------
srv := &http.Server{
Addr:         ":" + cfg.Port,
Handler:      r,
ReadTimeout:  10 * time.Second,
WriteTimeout: 30 * time.Second,
IdleTimeout:  60 * time.Second,
}

go func() {
log.Printf("?? listening on :%s (env=%s)", cfg.Port, cfg.Env)
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
log.Fatalf("server: %v", err)
}
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
log.Println("shutting down...")

shutCtx, shutCancel := context.WithTimeout(baseCtx, 10*time.Second)
defer shutCancel()
if err := srv.Shutdown(shutCtx); err != nil {
log.Printf("shutdown error: %v", err)
}
log.Println("bye")
}






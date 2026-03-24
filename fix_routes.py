content = open('cmd/api/main.go', encoding='utf-8').read()

old = '// -- Academic System ----------------------------------------'
new = '''// -- Guardians -----------------------------------------------
guardianH := handler.NewGuardianHandler()
grd := authed.Group("/guardians", middleware.RequireWrite())
grd.GET("",      guardianH.List)
grd.POST("",     guardianH.Create)
grd.GET("/:id",  guardianH.Get)
grd.PUT("/:id",  guardianH.Update)
grd.GET("/:id/students", guardianH.GetStudents)

// -- Academic System ----------------------------------------'''

if old in content:
    content = content.replace(old, new, 1)
    open('cmd/api/main.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')

content = open('cmd/api/main.go', encoding='utf-8').read()
old = '// -- Guardians -----------------------------------------------'
new = '''// -- Settings & Staff ----------------------------------------
settingsH := handler.NewSettingsHandler()
stg := authed.Group("/settings", middleware.RequireWrite())
stg.GET("",          settingsH.GetSettings)
stg.PUT("",          settingsH.UpdateSettings)
stg.GET("/roles",    settingsH.ListRoles)
stg.PUT("/roles/:id/permissions", settingsH.UpdateRolePermissions)
stg.GET("/staff",    settingsH.ListStaff)
stg.POST("/staff",   settingsH.CreateStaff)
stg.PUT("/staff/:id", settingsH.UpdateStaff)

// -- Guardians -----------------------------------------------'''
if old in content:
    content = content.replace(old, new, 1)
    open('cmd/api/main.go', 'w', encoding='utf-8').write(content)
    print('OK')
else:
    print('NOT FOUND')

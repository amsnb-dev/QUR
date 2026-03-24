with open('internal/handler/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

old = 'func (h *StudentHandler) List(c *gin.Context) {\n\ttx := middleware.TxFrom(c)\n\tf := model.ListStudentsFilter{\n\t\tIncludeArchived: c.Query("include_archived") == "1",\n\t\tStatus:          c.Query("status"),\n\t}'

new = 'func (h *StudentHandler) List(c *gin.Context) {\n\ttx := middleware.TxFrom(c)\n\tvar groupID *uuid.UUID\n\tif gid := c.Query("group_id"); gid != "" {\n\t\tif parsed, err := uuid.Parse(gid); err == nil {\n\t\t\tgroupID = &parsed\n\t\t}\n\t}\n\tf := model.ListStudentsFilter{\n\t\tIncludeArchived: c.Query("include_archived") == "1",\n\t\tStatus:          c.Query("status"),\n\t\tGroupID:         groupID,\n\t\tSearch:          c.Query("search"),\n\t}'

if old in content:
    content = content.replace(old, new)
    with open('internal/handler/student.go', 'w', encoding='utf-8') as f:
        f.write(content)
    print("SUCCESS")
else:
    print("NOT FOUND")

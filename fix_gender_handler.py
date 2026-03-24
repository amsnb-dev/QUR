with open('internal/handler/student.go', 'r', encoding='utf-8') as f:
    content = f.read()

old = 'f := model.ListStudentsFilter{\n\t\tIncludeArchived: c.Query("include_archived") == "1",\n\t\tStatus:          c.Query("status"),\n\t\tGroupID:         groupID,\n\t\tSearch:          c.Query("search"),\n\t}'
new = 'f := model.ListStudentsFilter{\n\t\tIncludeArchived: c.Query("include_archived") == "1",\n\t\tStatus:          c.Query("status"),\n\t\tGroupID:         groupID,\n\t\tSearch:          c.Query("search"),\n\t\tGender:          c.Query("gender"),\n\t}'

if old in content:
    content = content.replace(old, new)
    print("OK")
else:
    print("NOT FOUND")

with open('internal/handler/student.go', 'w', encoding='utf-8') as f:
    f.write(content)

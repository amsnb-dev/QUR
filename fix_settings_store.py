content = open("internal/store/settings.go", encoding="utf-8-sig").read()
# Find where the real settings code starts
idx = content.find("// -- Roles")
if idx > 0:
    content = "package store\n\nimport (\n\t\"context\"\n\t\"encoding/json\"\n\t\"fmt\"\n\n\t\"github.com/google/uuid\"\n\t\"github.com/jackc/pgx/v5\"\n\t\"golang.org/x/crypto/bcrypt\"\n\t\"github.com/quran-school/api/internal/model\"\n)\n\n" + content[idx:]
    open("internal/store/settings.go", "w", encoding="utf-8").write(content)
    print("OK - trimmed guardian code")
else:
    print("NOT FOUND - content starts with:", repr(content[:100]))

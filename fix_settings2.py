content = open("internal/store/settings.go", "r", encoding="utf-8-sig").read()
idx = content.find("// -- Roles")
if idx < 0:
    idx = content.find("func ListRoles")
if idx > 0:
    header = "package store\n\nimport (\n\t\"context\"\n\t\"encoding/json\"\n\t\"fmt\"\n\n\t\"github.com/google/uuid\"\n\t\"github.com/jackc/pgx/v5\"\n\t\"golang.org/x/crypto/bcrypt\"\n\t\"github.com/quran-school/api/internal/model\"\n)\n\n"
    open("internal/store/settings.go", "w", encoding="utf-8").write(header + content[idx:])
    print("OK - fixed imports")
else:
    print("Guardian code detected, need full rewrite")
    print(repr(content[:300]))

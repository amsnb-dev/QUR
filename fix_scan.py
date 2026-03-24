content = open('internal/store/guardian.go', encoding='utf-8').read()
old = '&g.Username, &g.PasswordPlain, &g.Relation, &g.IsActive'
new = '&g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive'
c = content.count(old)
content = content.replace(old, new)
print(f"Fixed {c}x")
open('internal/store/guardian.go', 'w', encoding='utf-8').write(content)

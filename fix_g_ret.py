content = open('internal/store/guardian.go', encoding='utf-8').read()
# Fix INSERT RETURNING
old = 'RETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, is_active, created_at, updated_at'
new = 'RETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at'
c = content.count(old)
content = content.replace(old, new)
print(f"RETURNING fixes: {c}")
open('internal/store/guardian.go', 'w', encoding='utf-8').write(content)

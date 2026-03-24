content = open('internal/store/guardian.go', encoding='utf-8').read()

# Fix INSERT
old1 = 'INSERT INTO guardians (school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)\n\t\tRETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, is_active, created_at, updated_at\n\t`, schoolID, r.FullName, r.Phone, r.Phone2, r.Email, r.Address, r.NationalID, r.Username, r.PasswordPlain, r.Notes,'
new1 = 'INSERT INTO guardians (school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)\n\t\tRETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at\n\t`, schoolID, r.FullName, r.Phone, r.Phone2, r.Email, r.Address, r.NationalID, r.Username, r.PasswordPlain, r.Notes, r.Relation,'

# Fix SELECT in Get and List
old2 = 'SELECT id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, is_active, created_at, updated_at'
new2 = 'SELECT id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at'

# Fix all Scan
old3 = '&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.IsActive, &g.CreatedAt, &g.UpdatedAt'
new3 = '&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive, &g.CreatedAt, &g.UpdatedAt'

# Fix UPDATE
old4 = 'UPDATE guardians SET full_name=$2, phone=$3, phone2=$4, email=$5, address=$6, national_id=$7, username=$8, password_plain=$9, notes=$10, is_active=$11\n\t\tWHERE id=$1\n\t\tRETURNING id, school_id, full_name, phone, phone2, email, address, national_id, u

$script = @'
content = open('internal/store/guardian.go', encoding='utf-8').read()

fixes = [
    ('INSERT INTO guardians (school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)\n\t\tRETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, is_active, created_at, updated_at\n\t`, schoolID, r.FullName, r.Phone, r.Phone2, r.Email, r.Address, r.NationalID, r.Username, r.PasswordPlain, r.Notes,',
     'INSERT INTO guardians (school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation)\n\t\tVALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)\n\t\tRETURNING id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at\n\t`, schoolID, r.FullName, r.Phone, r.Phone2, r.Email, r.Address, r.NationalID, r.Username, r.PasswordPlain, r.Notes, r.Relation,'),
    ('SELECT id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, is_active, created_at, updated_at',
     'SELECT id, school_id, full_name, phone, phone2, email, address, national_id, username, password_plain, notes, relation, is_active, created_at, updated_at'),
    ('&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.IsActive, &g.CreatedAt, &g.UpdatedAt',
     '&g.ID, &g.SchoolID, &g.FullName, &g.Phone, &g.Phone2, &g.Email, &g.Address, &g.NationalID, &g.Username, &g.PasswordPlain, &g.Notes, &g.Relation, &g.IsActive, &g.CreatedAt, &g.UpdatedAt'),
]

for old, new in fixes:
    c = content.count(old)
    content = content.replace(old, new)
    print(f"{'OK' if c else 'NOT FOUND'} ({c}x)")

open('internal/store/guardian.go', 'w', encoding='utf-8').write(content)

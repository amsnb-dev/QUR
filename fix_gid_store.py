content = open('internal/store/student.go', encoding='utf-8').read()

# Add guardian_id to INSERT
old1 = 'monthly_fee, fee_exemption, notes, gender, username, password_plain)'
new1 = 'monthly_fee, fee_exemption, notes, gender, username, password_plain, guardian_id)'

old2 = 'VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)'
new2 = 'VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)'

old3 = 'monthly_fee, fee_exemption, status, notes, gender, username, password_plain,'
new3 = 'monthly_fee, fee_exemption, status, notes, gender, username, password_plain, guardian_id,'

old4 = 'r.MonthlyFee, exemption, r.Notes, r.Gender, r.Username, r.PasswordPlain,'
new4 = 'r.MonthlyFee, exemption, r.Notes, r.Gender, r.Username, r.PasswordPlain, r.GuardianID,'

for i,(old,new) in enumerate([(old1,new1),(old2,new2),(old3,new3),(old4,new4)]):
    if old in content:
        content = content.replace(old, new, 1)
        print(f"OK: {i+1}")
    else:
        print(f"NOT FOUND: {i+1}")

open('internal/store/student.go', 'w', encoding='utf-8').write(content)

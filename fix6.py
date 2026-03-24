content = open('internal/store/student.go', encoding='utf-8').read()
idx = content.find('gender          = $11')
old = content[idx:idx+280]
new = old.replace(
    'monthly_fee, fee_exemption, status, notes, gender,\n\t\t          is_archived',
    'monthly_fee, fee_exemption, status, notes, gender, username,\n\t\t          is_archived'
)
content = content.replace(old, new, 1)
open('internal/store/student.go', 'w', encoding='utf-8').write(content)
print('OK')

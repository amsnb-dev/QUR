with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()
idx = content.find('MonthlyFee, &s.FeeExemption')
print(repr(content[idx-5:idx+150]))

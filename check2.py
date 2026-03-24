with open('internal/store/student.go', 'r', encoding='utf-8') as f:
    content = f.read()
# find all scan occurrences
import re
for m in re.finditer(r'MonthlyFee.*?UpdatedAt', content, re.DOTALL):
    print(repr(m.group()[:150]))
    print("---")

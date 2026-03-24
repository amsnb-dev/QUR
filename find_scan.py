content = open('internal/store/student.go', encoding='utf-8').read()
import re
for m in re.finditer(r"MonthlyFee.*?IsArchived", content, re.DOTALL):
    g = m.group()[:120]
    if "Username" in g and "PasswordPlain" not in g:
        print(repr(g))
        print("---")

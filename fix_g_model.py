content = open('internal/model/guardian.go', encoding='utf-8').read()
old = 'IsActive      bool      `json:"is_active"`'
new = 'Relation      *string   `json:"relation,omitempty"`\n\tIsActive      bool      `json:"is_active"`'
old2 = 'Notes         *string `json:"notes"`\n}'
new2 = 'Notes         *string `json:"notes"`\n\tRelation      *string `json:"relation"`\n}'
old3 = 'Notes         *string `json:"notes"`\n\tIsActive      *bool   `json:"is_active"`'
new3 = 'Notes         *string `json:"notes"`\n\tRelation      *string `json:"relation"`\n\tIsActive      *bool   `json:"is_active"`'
for old, new in [(old, new), (old2, new2), (old3, new3)]:
    if old in content:
        content = content.replace(old, new, 1)
        print("OK")
    else:
        print("NOT FOUND")
open('internal/model/guardian.go', 'w', encoding='utf-8').write(content)

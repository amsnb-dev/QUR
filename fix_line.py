with open('quran_school_fixed.html', 'r', encoding='utf-8') as f:
    lines = f.readlines()
lines[3163] = "      return '<div class=\"group-dropdown-item\" onmousedown=\"selectGuardian(\\''+g.id+'\\',\\''+nm+'\\',\\'\"+(g.phone||\"\")+\"\\')\">' + g.full_name + rel + ' \u2014 ' + (g.phone||'') + '</div>';\n"
with open('quran_school_fixed.html', 'w', encoding='utf-8') as f:
    f.writelines(lines)
print('OK')

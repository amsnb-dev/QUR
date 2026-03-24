# Quran School API — Backend MVP

**Stack:** Go 1.22 · Gin · pgx/v5 · PostgreSQL 16 · JWT HS256 · bcrypt cost=12

> كل شيء يعمل عبر Docker — لا يلزم تثبيت Go محلياً.

---

## التشغيل السريع

### 1. إعداد المتغيرات
```bash
cp .env.example .env
# افتح .env وعدّل:
#   JWT_SECRET=<openssl rand -hex 32>
#   JWT_REFRESH_SECRET=<openssl rand -hex 32>
```

### 2. تشغيل الـ Migrations (مرة واحدة فقط)
```bash
make migrate
```
يشغّل postgres، يطبّق كل الـ SQL files بالترتيب، ثم يوقف الـ api.

### 3. تشغيل الـ Stack العادي
```bash
make up
# أو مباشرة:
docker compose up --build
```
API متاح على `http://localhost:8080`

---

## Migrations عند إضافة ملفات جديدة

```bash
# أوقف الـ stack إذا كان شغّالاً
make down

# شغّل الـ migrations فقط
make migrate

# ثم أعد تشغيل الـ stack
make up
```

أو بخطوة واحدة (migrations + تشغيل مستمر):
```bash
make up-migrate
# مكافئ لـ: RUN_MIGRATIONS=true docker compose up --build
```

---

## Endpoints الكاملة

### Public (لا يحتاج Auth)
| Method | Path | الوصف |
|--------|------|-------|
| GET | /health | فحص الخادم + DB ping |
| POST | /auth/login | → `{access_token, refresh_token, expires_in}` |
| POST | /auth/refresh | تجديد access token |

### Auth (يحتاج Bearer token)
| Method | Path | الوصف |
|--------|------|-------|
| POST | /auth/logout | إبطال كل refresh tokens |
| GET | /me | بيانات المستخدم الحالي |

### Groups `/groups`
| Method | Path | الأدوار |
|--------|------|---------|
| GET | /groups?limit=&offset=&include_archived=1 | all except accountant |
| GET | /groups/:id | all except accountant |
| POST | /groups | school_admin · super_admin |
| PUT | /groups/:id | school_admin · super_admin |
| PATCH | /groups/:id/archive | school_admin · super_admin |

### Students `/students`
| Method | Path | الأدوار |
|--------|------|---------|
| GET | /students?limit=&offset=&status=&include_archived=1 | all except accountant |
| GET | /students/:id | all except accountant |
| POST | /students | school_admin · super_admin |
| PUT | /students/:id | school_admin · super_admin |
| PATCH | /students/:id/archive | school_admin · super_admin |

### Student Memberships
| Method | Path | الوصف |
|--------|------|-------|
| POST | /students/:id/groups | إلحاق بحلقة (auto-close old primary) |
| GET | /students/:id/groups?current=1 | كل العضويات / الحالية فقط |
| PATCH | /student-groups/:id/close | إغلاق عضوية |

### Attendance
| Method | Path | الأدوار |
|--------|------|---------|
| POST | /groups/:id/attendance | supervisor · school_admin · super_admin · teacher* |
| GET | /groups/:id/attendance?date=&limit=&offset= | all except accountant |
| PATCH | /attendance/:id | supervisor · school_admin · super_admin · teacher* |

*teacher: يكتب فقط لمجموعاته (groups.teacher_id = user.teacher_id)

### Memorization
| Method | Path | الأدوار |
|--------|------|---------|
| POST | /students/:id/memorization | supervisor · school_admin · super_admin · teacher* |
| GET | /students/:id/memorization?limit=&offset= | all except accountant |
| PATCH | /memorization/:id | supervisor · school_admin · super_admin · teacher* |

*teacher: يحدد group_id إلزامي — يجب أن يكون معلّم ذلك الحلقة

---

## Pagination

جميع قوائم الـ GET ترجع:

```json
{
  "data": [...],
  "meta": { "limit": 50, "offset": 0, "total": 243 }
}
```

| param | default | max |
|-------|---------|-----|
| limit | 50 | 200 |
| offset | 0 | — |

---

## أمثلة cURL كاملة

```bash
BASE=http://localhost:8080

# ── Auth ──────────────────────────────────────────────────────
TOKEN=$(curl -s -X POST $BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@quran.dev","password":"Admin1234!"}' \
  | jq -r .access_token)

AUTH="Authorization: Bearer $TOKEN"

# ── Health ────────────────────────────────────────────────────
curl -s $BASE/health | jq

# ── Groups ─────────────────────────────────────────────────────
# إنشاء
curl -s -X POST $BASE/groups -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"teacher_id":"<UUID>","name":"حلقة الفجر","level":"beginner","capacity":20,"days":[0,1,3,4],"start_time":"06:00","end_time":"07:30"}' | jq

# قائمة (مع pagination)
curl -s "$BASE/groups?limit=20&offset=0" -H "$AUTH" | jq
curl -s "$BASE/groups?include_archived=1" -H "$AUTH" | jq

# أرشفة
curl -s -X PATCH "$BASE/groups/<UUID>/archive" -H "$AUTH" | jq

# ── Students ───────────────────────────────────────────────────
# إنشاء
curl -s -X POST $BASE/students -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"full_name":"محمد أحمد","date_of_birth":"2012-05-10T00:00:00Z","guardian_name":"أحمد","memorized_parts":5.5,"monthly_fee":150}' | jq

# قائمة (مع فلتر status)
curl -s "$BASE/students?status=active&limit=50" -H "$AUTH" | jq

# إلحاق بحلقة
curl -s -X POST "$BASE/students/<UUID>/groups" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"group_id":"<GROUP_UUID>","is_primary":true,"start_date":"2025-09-01T00:00:00Z"}' | jq

# إغلاق عضوية
curl -s -X PATCH "$BASE/student-groups/<MEMBERSHIP_UUID>/close" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"end_date":"2025-12-31T00:00:00Z"}' | jq

# ── Attendance ─────────────────────────────────────────────────
# تسجيل حضور جماعي (upsert)
curl -s -X POST "$BASE/groups/<GROUP_UUID>/attendance" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "date": "2025-10-01",
    "items": [
      {"student_id":"<UUID1>","status":"present"},
      {"student_id":"<UUID2>","status":"absent","note":"مرض"},
      {"student_id":"<UUID3>","status":"late"}
    ]
  }' | jq

# قراءة حضور يوم معين
curl -s "$BASE/groups/<GROUP_UUID>/attendance?date=2025-10-01" -H "$AUTH" | jq

# قراءة بـ pagination
curl -s "$BASE/groups/<GROUP_UUID>/attendance?limit=30&offset=0" -H "$AUTH" | jq

# تحديث سجل واحد
curl -s -X PATCH "$BASE/attendance/<ATT_UUID>" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"status":"excused","note":"عذر رسمي"}' | jq

# ── Memorization ───────────────────────────────────────────────
# تسجيل حفظ جديد
curl -s -X POST "$BASE/students/<STUDENT_UUID>/memorization" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{
    "date": "2025-10-01T08:00:00Z",
    "surah_number": 2,
    "from_verse": 1,
    "to_verse": 20,
    "entry_type": "new",
    "grade": 4,
    "notes": "جيد مع بعض الأخطاء البسيطة",
    "group_id": "<GROUP_UUID>"
  }' | jq

# قراءة سجلات الحفظ (مع pagination)
curl -s "$BASE/students/<UUID>/memorization?limit=20&offset=0" -H "$AUTH" | jq

# تحديث سجل
curl -s -X PATCH "$BASE/memorization/<MEM_UUID>" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"grade":5,"notes":"ممتاز"}' | jq
```

---

## RBAC

| الدور | GET | POST/PUT | PATCH write | archive |
|-------|-----|----------|-------------|---------|
| super_admin | ✅ | ✅ | ✅ | ✅ |
| school_admin | ✅ | ✅ | ✅ | ✅ |
| supervisor | ✅ | attendance/mem فقط | ✅ | ❌ |
| teacher | ✅ | مجموعاته فقط | مجموعاته فقط | ❌ |
| accountant | ❌ 403 | ❌ 403 | ❌ 403 | ❌ 403 |

---

## Migrations

```bash
# تشغيل كل الـ migrations (مرة واحدة أو عند إضافة ملفات جديدة)
make migrate

# تشغيل مع إبقاء الـ api شغّالاً
make up-migrate

# الوضع العادي (بدون migrations)
make up
```

> `RUN_MIGRATIONS=false` هو الافتراضي — الـ api يبدأ بدون تشغيل أي migrations.

ترتيب التطبيق:
```
00_init.sql               → pgcrypto
01_schema.sql             → schools · roles · users · refresh_tokens + RLS
02_seed.sql               → super@quran.dev + admin@quran.dev
03_core_mvp.sql           → teachers · groups · students · attendance · memorization + RLS
04_attendance_memorization.sql → updated_at trigger + surahs seed (114)
```

---

## هيكل المشروع

```
quran-api/
├── cmd/api/main.go
├── internal/
│   ├── auth/           ← bcrypt · JWT · refresh tokens
│   ├── config/         ← .env + RunMigrations flag
│   ├── db/             ← pgxpool + TxWithTenant()
│   ├── migrate/        ← migration runner
│   ├── middleware/
│   │   ├── auth.go     ← RequireAuth · TxFrom · UserFrom · RequireRole
│   │   └── rbac.go     ← RequireWrite · RequireNotAccountant
│   ├── store/
│   │   ├── errors.go   ← ErrNotFound · ErrConflict · ErrForbidden
│   │   ├── pagination.go
│   │   ├── audit.go    ← InsertAudit (INSERT ONLY)
│   │   ├── group.go
│   │   ├── student.go
│   │   ├── attendance.go
│   │   └── memorization.go
│   ├── handler/
│   │   ├── auth.go
│   │   ├── group.go
│   │   ├── student.go
│   │   ├── attendance.go
│   │   └── memorization.go
│   └── model/
│       ├── model.go
│       ├── group_student.go
│       └── attendance_memorization.go
└── migrations/
    ├── 00_init.sql
    ├── 01_schema.sql
    ├── 02_seed.sql
    ├── 03_core_mvp.sql
    └── 04_attendance_memorization.sql
```

---

## ضمانات التصميم

| القاعدة | التطبيق |
|---------|---------|
| Tx واحدة لكل request | `TxFrom(c)` في جميع handlers |
| school_id من التوكن فقط | `resolveSchoolID(c)` تقرأ JWT claims |
| لا DELETE | DB triggers تمنعه · store يستخدم archive |
| Multi-tenant RLS | `SET LOCAL app.school_id` في كل tx |
| Audit trail | `InsertAudit` على كل write + archive |
| Pagination موحّدة | `ParsePage` + `PagedResult[T]` + `Meta` |

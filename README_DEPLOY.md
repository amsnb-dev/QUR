# دليل النشر — Quran School API (Beta)

دليل نشر كامل خطوة بخطوة على VPS Ubuntu 22.04. لا يلزم تثبيت Go محلياً — كل شيء يعمل عبر Docker.

---

## المتطلبات

| المتطلب | التفصيل |
|---------|---------|
| VPS | Ubuntu 22.04 LTS (512MB RAM كحد أدنى، يُوصى بـ 2GB) |
| CPU | vCPU واحد على الأقل |
| Storage | 20GB على الأقل |
| Domain | دومين يشير إلى IP الـ VPS (A record) |
| Port | 22 (SSH) · 80 (HTTP) · 443 (HTTPS) |

---

## الخطوة 1 — إنشاء مستخدم غير root

**على جهازك المحلي:** ادخل على الـ VPS بـ root أولاً.

```bash
ssh root@YOUR_VPS_IP
```

**على الـ VPS:**

```bash
# إنشاء مستخدم جديد
adduser quran
# أضفه لمجموعة sudo
usermod -aG sudo quran
# أضفه لمجموعة docker (بعد تثبيته)
usermod -aG docker quran
```

**نقل SSH key من root إلى المستخدم الجديد:**

```bash
mkdir -p /home/quran/.ssh
cp /root/.ssh/authorized_keys /home/quran/.ssh/
chown -R quran:quran /home/quran/.ssh
chmod 700 /home/quran/.ssh
chmod 600 /home/quran/.ssh/authorized_keys
```

**اختبر الدخول بالمستخدم الجديد (في terminal جديد):**

```bash
ssh quran@YOUR_VPS_IP
```

**تعطيل دخول root عبر SSH** (بعد التأكد من عمل الدخول بالمستخدم الجديد):

```bash
sudo sed -i 's/^PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo systemctl restart sshd
```

---

## الخطوة 2 — تثبيت Docker + Compose Plugin

```bash
# تحديث النظام
sudo apt update && sudo apt upgrade -y

# تثبيت المتطلبات
sudo apt install -y ca-certificates curl gnupg lsb-release

# إضافة Docker GPG key
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# إضافة Docker repo
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
    | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# تثبيت Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io \
                    docker-buildx-plugin docker-compose-plugin

# تشغيل Docker تلقائياً عند الـ boot
sudo systemctl enable --now docker

# إضافة المستخدم لمجموعة docker
sudo usermod -aG docker $USER
newgrp docker

# اختبار
docker --version
docker compose version
```

**إعداد Docker log rotation** (منع تضخم ملفات الـ logs):

```bash
sudo tee /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "5"
  }
}
EOF
sudo systemctl restart docker
```

---

## الخطوة 3 — إعداد DNS

في لوحة إدارة الدومين الخاصة بك، أضف **A record**:

```
Type:  A
Name:  api          (أو @ للدومين الرئيسي)
Value: YOUR_VPS_IP
TTL:   300
```

تحقق من الانتشار (قد يستغرق 5-60 دقيقة):

```bash
dig api.yourdomain.com +short
# يجب أن يظهر IP الـ VPS
```

---

## الخطوة 4 — رفع المشروع ونشره

**على الـ VPS:**

```bash
# إنشاء مجلد المشروع
mkdir -p /home/quran/quran-api-prod
cd /home/quran/quran-api-prod

# مجلد الـ backups
mkdir -p /opt/quran/backups
```

**من جهازك المحلي — رفع الملفات:**

```bash
scp quran-api-prod.zip quran@YOUR_VPS_IP:/home/quran/
ssh quran@YOUR_VPS_IP "cd /home/quran && unzip quran-api-prod.zip"
```

أو عبر git:

```bash
# على الـ VPS
git clone https://github.com/YOUR_ORG/quran-api-prod.git /home/quran/quran-api-prod
```

**إعداد ملف .env.prod:**

```bash
cd /home/quran/quran-api-prod/deploy
cp .env.prod.example .env.prod
chmod 600 .env.prod    # قراءة للمالك فقط
```

عدّل القيم الإلزامية:

```bash
nano .env.prod
```

```ini
POSTGRES_PASSWORD=<كلمة مرور قوية: openssl rand -base64 32 | tr -d /=+>
JWT_SECRET=<openssl rand -hex 32>
JWT_REFRESH_SECRET=<openssl rand -hex 32>
DOMAIN=api.yourdomain.com
BACKUP_DIR=/opt/quran/backups
```

---

## الخطوة 5 — تفعيل UFW (Firewall)

```bash
sudo apt install -y ufw

# السياسة الافتراضية: منع الدخول، السماح بالخروج
sudo ufw default deny incoming
sudo ufw default allow outgoing

# السماح بالبورتات المطلوبة فقط
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS

# تفعيل UFW
sudo ufw enable
sudo ufw status verbose
```

---

## الخطوة 6 — تفعيل fail2ban (حماية SSH)

```bash
sudo apt install -y fail2ban

sudo tee /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime  = 1h
findtime = 10m
maxretry = 5

[sshd]
enabled  = true
port     = ssh
logpath  = %(sshd_log)s
backend  = %(syslog_backend)s
EOF

sudo systemctl enable --now fail2ban
sudo fail2ban-client status sshd
```

---

## الخطوة 7 — بناء الـ Docker Image

```bash
cd /home/quran/quran-api-prod

# بناء الـ image (يشغّل gofmt + go test + go build داخل Docker)
docker build -t quran-api:latest .
```

> الـ Dockerfile يفشل تلقائياً إذا كان الكود غير مُنسَّق أو الاختبارات تفشل.

---

## الخطوة 8 — تشغيل Migrations (مرة واحدة)

```bash
cd /home/quran/quran-api-prod/deploy

# تشغيل postgres أولاً
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d postgres

# انتظر حتى يكون postgres healthy
until docker compose -f docker-compose.prod.yml --env-file .env.prod \
    exec postgres pg_isready -U quran -d quran_school; do
    sleep 2
done

# تشغيل المigrations
RUN_MIGRATIONS=true \
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d api

# تابع الـ logs حتى ترى رسالة النجاح
docker logs -f quran_api_prod
# ابحث عن: ✅ migrations applied

# أوقف الـ api مؤقتاً
docker compose -f docker-compose.prod.yml --env-file .env.prod stop api
```

> بعد نجاح الـ migrations، تأكد أن `RUN_MIGRATIONS=false` في `.env.prod`.

---

## الخطوة 9 — تفعيل HTTPS عبر Certbot

**تثبيت certbot:**

```bash
sudo apt install -y certbot
```

**الحصول على شهادة TLS:**

> **لماذا webroot وليس standalone؟**
> وضع `--standalone` يشغّل خادم HTTP مؤقت على بورت 80، لكن nginx يكون شغّالاً بالفعل على نفس البورت — مما يسبب تعارضاً ويمنع التجديد التلقائي لاحقاً. وضع `--webroot` يكتب ملف التحقق داخل مجلد يخدمه nginx مباشرة دون إيقافه.

```bash
# إنشاء مجلد webroot (يخدمه nginx عبر /.well-known/acme-challenge/)
sudo mkdir -p /var/www/certbot

# توليد الشهادة — nginx يجب أن يكون شغّالاً على 80
sudo certbot certonly --webroot \
    -w /var/www/certbot \
    -d api.yourdomain.com \
    --agree-tos \
    --non-interactive \
    --email admin@yourdomain.com
```

**تحديث site.conf بالدومين:**

```bash
cd /home/quran/quran-api-prod/deploy/nginx

# استبدل YOUR_DOMAIN بالدومين الفعلي
sed -i 's/YOUR_DOMAIN/api.yourdomain.com/g' site.conf
```

---

## الخطوة 10 — تشغيل الـ Stack الكامل

```bash
cd /home/quran/quran-api-prod/deploy

docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

# تحقق من الحالة
docker compose -f docker-compose.prod.yml --env-file .env.prod ps
```

يجب أن يظهر ثلاثة containers بحالة **healthy**:

```
NAME               STATUS          PORTS
quran_db_prod      healthy         (داخلي فقط)
quran_api_prod     healthy         (داخلي فقط)
quran_nginx_prod   healthy         0.0.0.0:80->80, 0.0.0.0:443->443
```

**اختبار:**

```bash
curl -s https://api.yourdomain.com/health | jq
# {"status":"ok"}

curl -s -X POST https://api.yourdomain.com/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@quran.dev","password":"Admin1234!"}' | jq
```

---

## الخطوة 11 — إعداد Backups التلقائية

```bash
# التأكد من صلاحيات السكريبت
chmod +x /home/quran/quran-api-prod/deploy/scripts/backup_postgres.sh
chmod +x /home/quran/quran-api-prod/deploy/scripts/restore_postgres.sh

# اختبار يدوي
/home/quran/quran-api-prod/deploy/scripts/backup_postgres.sh

# التحقق من وجود الملف
ls -lh /opt/quran/backups/

# إضافة cron job
crontab -e
```

أضف هذا السطر:

```cron
0 3 * * * /home/quran/quran-api-prod/deploy/scripts/backup_postgres.sh >> /var/log/quran_backup.log 2>&1
```

**اختبار Restore** (على بيئة تجريبية — لا تفعل على production إلا عند الحاجة):

```bash
# أوقف الـ api أثناء الاسترجاع
docker stop quran_api_prod

# استرجع من آخر backup
LATEST=$(ls -t /opt/quran/backups/*.sql.gz | head -1)
/home/quran/quran-api-prod/deploy/scripts/restore_postgres.sh "$LATEST"

# أعد تشغيل الـ api
docker start quran_api_prod
```

---

## الخطوة 12 — تجديد شهادة TLS التلقائي

Certbot يجدد تلقائياً. أضف هذا السطر لـ cron لضمان تحديث nginx:

```bash
crontab -e
```

```cron
0 12 * * * certbot renew --quiet && docker exec quran_nginx_prod nginx -s reload
```

---

## التحديثات (Zero Downtime تقريباً)

### تحديث عادي (بدون migrations جديدة)

```bash
cd /home/quran/quran-api-prod

# سحب التغييرات الجديدة
git pull origin main

# إعادة بناء الـ image
docker build -t quran-api:latest .

# إعادة تشغيل الـ api فقط (postgres و nginx يبقيان شغّالين)
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env.prod \
    up -d --no-deps --build api

# تحقق من الـ health
sleep 5
curl -s https://api.yourdomain.com/health | jq
```

### تحديث مع migrations جديدة

```bash
cd /home/quran/quran-api-prod
git pull origin main

# بناء الـ image الجديد
docker build -t quran-api:latest .

# تشغيل الـ migrations (api سيُعاد تشغيله بـ RUN_MIGRATIONS=true)
RUN_MIGRATIONS=true \
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env.prod \
    up -d --no-deps api

# تابع الـ logs
docker logs -f quran_api_prod
# ابحث عن: ✅ migrations applied

# بعد نجاح الـ migrations، أعد تشغيله بالوضع العادي
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env.prod \
    stop api
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env.prod \
    up -d api
```

---

## المراقبة والـ Logs

```bash
# logs الـ api (آخر 100 سطر + متابعة)
docker logs -f --tail=100 quran_api_prod

# logs nginx
docker logs -f quran_nginx_prod

# logs postgres
docker logs -f quran_db_prod

# حالة جميع الـ containers
docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env.prod ps

# استخدام الموارد
docker stats --no-stream
```

---

## استكشاف المشاكل الشائعة

| المشكلة | الحل |
|---------|------|
| `api: service_healthy` لا يتحقق | تحقق `docker logs quran_api_prod` — ربما env var ناقص |
| `502 Bad Gateway` | الـ api لم يبدأ — `docker ps` + `docker logs quran_api_prod` |
| `SSL_ERROR_RX_RECORD_TOO_LONG` | الشهادة غير موجودة — تحقق `/etc/letsencrypt/live/` |
| Postgres لا يبدأ | تحقق من صلاحيات `/var/lib/docker/volumes/quran_pgdata_prod` |
| Backup يفشل | تأكد أن `POSTGRES_USER` و `CONTAINER_NAME` صحيحان في `.env.prod` |
| `429 Too Many Requests` | rate limit للـ /auth/login (5 طلبات/دقيقة من نفس IP) |

---

## Secrets المطلوبة للتوليد

```bash
# Database password
openssl rand -base64 32 | tr -d /=+ | cut -c1-32

# JWT secrets (× 2)
openssl rand -hex 32
openssl rand -hex 32
```

---

## بنية ملفات Deploy

```
deploy/
├── docker-compose.prod.yml   ← stack الإنتاج (api + postgres + nginx)
├── .env.prod.example         ← قالب المتغيرات (انسخه إلى .env.prod)
├── nginx/
│   ├── nginx.conf            ← إعدادات nginx الرئيسية (gzip · rate limiting · websocket)
│   ├── site.conf             ← Virtual host (HTTP redirect + HTTPS reverse proxy)
│   └── proxy_params.conf     ← Headers مشتركة لكل proxy locations
└── scripts/
    ├── backup_postgres.sh    ← pg_dump + gzip + retention 14 يوم
    ├── restore_postgres.sh   ← استرجاع من backup
    ├── rotate_logs.sh        ← تنظيف logs
    └── make_backup_cron.txt  ← أسطر crontab جاهزة
```

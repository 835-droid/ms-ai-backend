# إصلاح مشكلة جداول Favorite Lists

## المشكلة
خطأ `pq: relation "favorite_lists" does not exist` يظهر عند محاولة الوصول لصفحة المفضلة أو القوائم.

## السبب
جداول `favorite_lists` و `favorite_list_items` غير موجودة in قاعدة بيانات PostgreSQL.

## الحل

### الطريقة 1: تشغيل ملف SQL المباشر (الأسرع)
```bash
# قم بتعديل المتغيرات التالية حسب بيئة عملك
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=ms_ai_db

# تشغيل ملف الإصلاح
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f MS-AI/scripts/fix_favorite_lists.sql
```

أو ببساطة:
```bash
cd MS-AI/scripts
psql -U postgres -d ms_ai_db -f fix_favorite_lists.sql
```

### الطريقة 2: إعادة تشغيل ملف migration الكامل
إذا كنت تريد التأكد من أن جميع الجداول موجودة:
```bash
psql -U postgres -d ms_ai_db -f MS-AI/scripts/migrate_postgres.sql
```

### الطريقة 3: استخدام Docker (إذا كنت تستخدم docker-compose)
```bash
# الدخول إلى container قاعدة البيانات
docker exec -it postgres-container psql -U postgres -d ms_ai_db

# ثم داخل psql
\i /docker-entrypoint-initdb.d/fix_favorite_lists.sql
```

## التحقق من نجاح الإصلاح

بعد تشغيل الـ migration، يمكنك التحقق من وجود الجداول:

```sql
-- قائمة جميع الجداول
\dt

-- التحقق من جدول favorite_lists
SELECT table_name FROM information_schema.tables 
WHERE table_name IN ('favorite_lists', 'favorite_list_items');

-- عرض بنية الجدول
\d favorite_lists
\d favorite_list_items
```

## ملاحظات مهمة

1. **لا تحذف البيانات الموجودة**: استخدام `CREATE TABLE IF NOT EXISTS` يضمن عدم حذف أي بيانات موجودة.

2. **النسخ الاحتياطي**: يفضل أخذ نسخة احتياطية من قاعدة البيانات قبل تشغيل أي migration:
   ```bash
   pg_dump -U postgres -d ms_ai_db > backup_$(date +%Y%m%d).sql
   ```

3. **إذا كان التطبيق يعمل**: بعد تشغيل الـ migration، قم بإعادة تشغيل الخادم (Go server) لضمان تحميل التغييرات.

## استعلامات الاختبار

بعد الإصلاح، جرب هذه الاستعلامات للتأكد:

```sql
-- إدراج قائمة تجريبية
INSERT INTO favorite_lists (id, user_id, name, description, is_public) 
VALUES ('test-list-1', 'user-123', 'قائمتي التجريبية', 'قائمة للمانجا المفضلة', false);

-- إدراج عنصر تجريبي
INSERT INTO favorite_list_items (list_id, manga_id, notes) 
VALUES ('test-list-1', 'manga-123', 'مانجا رائعة!');

-- عرض القوائم
SELECT * FROM favorite_lists WHERE user_id = 'user-123';

-- عرض عناصر قائمة معينة
SELECT fl.name, fli.manga_id, fli.notes 
FROM favorite_lists fl
JOIN favorite_list_items fli ON fl.id = fli.list_id
WHERE fl.id = 'test-list-1';
```

## المشاكل الشائعة

### خطأ "relation already exists"
إذا ظهر هذا الخطأ، فهذا يعني أن الجداول موجودة بالفعل. تحقق من اسم قاعدة البيانات أو اسم المستخدم.

### خطأ "permission denied"
تأكد من أن مستخدم قاعدة البيانات لديه صلاحيات CREATE TABLE.

### خطأ "database does not exist"
تأكد من أن قاعدة البيانات `ms_ai_db` (أو الاسم الذي تستخدمه) موجودة.

## الدعم

إذا استمرت المشكلة، تحقق من:
1. ملف `.env` للتأكد من إعدادات قاعدة البيانات
2. سجلات PostgreSQL (`/var/log/postgresql/` على Linux)
3. تأكد من أن خادم PostgreSQL يعمل
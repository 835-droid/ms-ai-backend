# إصلاح مشكلة "User Not Found" وإنشاء الروايات

## المشكلة

كانت هناك مشكلتان رئيسيتان تمنعان إنشاء روايات جديدة وتسببان خطأ "user not found":

### المشكلة 1: أسماء مفاتيح السياق الخاطئة (تم الإصلاح)

في `novel_handler.go`، كانت الدالة `getCallerInfo` تقرأ من مفاتيح خاطئة في سياق Gin:
- كانت تقرأ من `"userID"` و `"roles"`
- لكن الـ auth middleware كان يعيّن `"user_id"` و `"user_roles"`

هذا التناقض menyebabkan المستخدم لا يُوجد أبداً.

### المشكلة 2: Novel Repository يعمل فقط مع MongoDB (تم الإصلاح)

النظام كان مصمماً للعمل فقط مع MongoDB:
- `NovelRepository` كان يُنشأ فقط عندما يكون MongoDB متاحاً
- في وضع PostgreSQL أو Hybrid، كان NovelRepository يكون nil
- هذا يسبب أخطاء عند محاولة إنشاء روايات

## الحل المطبق

### 1. إصلاح أسماء المفاتيح في `novel_handler.go`

```go
// قبل الإصلاح
userID, exists := c.Get("userID")  // خطأ
roles, _ := c.Get("roles")          // خطأ

// بعد الإصلاح
uid, exists := c.Get("user_id")     // صحيح
r, _ := c.Get("user_roles")         // صحيح
```

### 2. إنشاء دالتين منفصلتين

- `getCallerInfo(c)`: ترجع user ID كـ string (متوافق مع كلا النظامين)
- `getCallerObjectID(c)`: ترجع user ID كـ MongoDB ObjectID (يحول من string)

### 3. إنشاء PostgreSQL Novel Repository

تم إنشاء ملف جديد `internal/data/content/novel/postgres_novel_repository.go` الذي:
- يطبق واجهة `NovelRepository` كاملة (24 دالة)
- يستخدم PostgreSQL بدلاً من MongoDB
- يدعم جميع عمليات CRUD والتفاعلات

### 4. تحديث repo_initializers.go

تم تعديل منطق إنشاء NovelRepository ليدعم:
- وضع PostgreSQL: يستخدم `PostgresNovelRepository`
- وضع Hybrid: يفضل PostgreSQL إذا كان متاحاً، وإلا يستخدم MongoDB
- وضع MongoDB: يستخدم `MongoNovelRepository`

## الملفات المعدلة

1. **`MS-AI/internal/api/handler/content/novel/novel_handler.go`**
   - إصلاح `getCallerInfo` لقراءة المفاتيح الصحيحة
   - إضافة `getCallerObjectID` لتحويل string إلى ObjectID
   - تحديث جميع الدوال لاستخدام الدالة المناسبة

2. **`MS-AI/internal/api/handler/content/novel/novel_interaction_handler.go`**
   - تحديث جميع الدوال لاستخدام `getCallerObjectID`

3. **`MS-AI/internal/data/content/novel/postgres_novel_repository.go`** (جديد)
   - إنشاء PostgreSQL Novel Repository كامل

4. **`MS-AI/internal/container/repo_initializers.go`**
   - تحديث منطق إنشاء NovelRepository ليدعم PostgreSQL

## ملاحظات مهمة

1. **قواعد البيانات**: نظام الروايات الآن يعمل مع كل من MongoDB و PostgreSQL.

2. **الوضع الهجين (Hybrid)**: حسب ملف `.env`، النظام يعمل في وضع `hybrid` مما يعني أن كلا من MongoDB و PostgreSQL مطلوبان.

3. **معرفات المستخدمين**: يجب أن تكون معرفات المستخدمين بصيغة MongoDB ObjectID صحيحة (24 حرف hexadecimal) أو UUID.

4. **الجداول المطلوبة**: تأكد من تشغيل `scripts/novel_tables.sql` لإنشاء جداول PostgreSQL المطلوبة.

## الاختبار

بعد تطبيق الإصلاحات:
1. تأكد من أن قواعد البيانات متصلة وتعمل
2. قم بتشغيل migration للجداول: `psql -f scripts/novel_tables.sql`
3. سجل دخول كمستخدم صالح
4. حاول إنشاء رواية جديدة عبر API

```bash
curl -X POST http://localhost:8080/novels \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Novel",
    "description": "A test novel",
    "tags": ["test", "fiction"]
  }'
```

إذا نجح الطلب، فستحصل على استجابة 201 مع تفاصيل الرواية الجديدة.

## الخلاصة

تم إصلاح المشكلتين:
1. ✅ أسماء مفاتيح السياق الآن صحيحة
2. ✅ Novel Repository يعمل مع PostgreSQL و MongoDB

الآن يمكن إنشاء روايات جديدة بنجاح في جميع أوضاع قواعد البيانات.
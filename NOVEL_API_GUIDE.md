# دليل API الروايات

## نظرة عامة
تم إضافة نظام الروايات (Novels) بشكل كامل إلى الباك إند. النظام يعتمد على MongoDB لتخزين البيانات.

## نقاط النهاية (Endpoints)

### الروايات

| الطريقة | المسار | الوصف |
|---------|--------|-------|
| GET | `/api/novels` | قائمة جميع الروايات |
| GET | `/api/novels/:novelID` | الحصول على رواية محددة |
| POST | `/api/novels` | إضافة رواية جديدة (يتطلب صلاحية admin) |
| PUT | `/api/novels/:novelID` | تحديث رواية (يتطلب صلاحية admin) |
| DELETE | `/api/novels/:novelID` | حذف رواية (يتطلب صلاحية admin) |

### الإحصائيات

| الطريقة | المسار | الوصف |
|---------|--------|-------|
| GET | `/api/novels/most-viewed` | الروايات الأكثر مشاهدة |
| GET | `/api/novels/recently-updated` | الروايات المحدثة حديثاً |
| GET | `/api/novels/most-followed` | الروايات الأكثر متابعة |
| GET | `/api/novels/top-rated` | الروايات الأعلى تقييماً |

### التفاعلات (تتطلب تسجيل دخول)

| الطريقة | المسار | الوصف |
|---------|--------|-------|
| POST | `/api/novels/:novelID/view` | تسجيل مشاهدة |
| POST | `/api/novels/:novelID/favorite` | إضافة للمفضلة |
| DELETE | `/api/novels/:novelID/favorite` | إزالة من المفضلة |
| GET | `/api/novels/:novelID/favorite` | التحقق من المفضلة |
| POST | `/api/novels/:novelID/react` | إضافة رد فعل |
| GET | `/api/novels/:novelID/my-reaction` | الحصول على رد الفعل |
| POST | `/api/novels/:novelID/rate` | تقييم الرواية |
| POST | `/api/novels/:novelID/comments` | إضافة تعليق |
| GET | `/api/novels/:novelID/comments` | الحصول على التعليقات |
| DELETE | `/api/novels/:novelID/comments/:comment_id` | حذف تعليق |

## إضافة رواية جديدة

### الطلب (Request)
```json
POST /api/novels
{
  "title": "اسم الرواية",
  "description": "وصف الرواية",
  "author_id": "معرف المؤلف",
  "cover_image": "https://example.com/cover.jpg",
  "status": "ongoing",
  "tags": ["خيال", "مغامرات"],
  "categories": ["fantasy", "adventure"]
}
```

### الحقول:
- `title` (مطلوب): عنوان الرواية
- `description` (اختياري): وصف الرواية
- `author_id` (اختياري): معرف المؤلف
- `cover_image` (اختياري): رابط صورة الغلاف
- `status` (اختياري): حالة الرواية (ongoing, completed, hiatus)
- `tags` (اختياري): قائمة العلامات
- `categories` (اختياري): قائمة التصنيفات

### الاستجابة (Response)
```json
{
  "success": true,
  "data": {
    "id": "...",
    "title": "اسم الرواية",
    "slug": "اسم-الرواية",
    ...
  }
}
```

## إضافة فصل للرواية

### الطلب (Request)
```json
POST /api/novels/:novelID/chapters
{
  "number": 1,
  "title": "عنوان الفصل",
  "content": "محتوى الفصل...",
  "word_count": 5000
}
```

## ملاحظات هامة

1. **المصادقة**: جميع عمليات التعديل (POST, PUT, DELETE) تتطلب توكن صلاحية admin
2. **MongoDB**: نظام الروايات يعمل فقط مع MongoDB
3. **الصور**: يتم تخزين روابط الصور فقط، يجب رفع الصور بشكل منفصل

## اختبار النظام

### 1. تشغيل قاعدة البيانات
```bash
docker-compose up -d
```

### 2. ترحيل الجداول
```bash
psql -U postgres -d msai_db -f scripts/novel_tables.sql
```

### 3. تشغيل التطبيق
```bash
cd MS-AI
go run cmd/web/main.go
```

### 4. تسجيل دخول كـ admin
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

### 5. إضافة رواية
```bash
curl -X POST http://localhost:8080/api/novels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "رواية تجريبية",
    "description": "هذه رواية للتجربة",
    "status": "ongoing"
  }'
```

## الفرونت إند

### صفحات الروايات المتاحة:
- `/novels.html` - قائمة الروايات
- `/novel-library.html` - مكتبة الروايات
- `/novel-details.html` - تفاصيل الرواية
- `/novel-reader.html` - قارئ الرواية
- `/admin.html` - لوحة الإدارة (تدعم إضافة الروايات)

### استخدام API في الفرونت إند:
```javascript
// إضافة رواية
await apiFetch('/novels', {
    method: 'POST',
    body: JSON.stringify(novelData)
});

// الحصول على قائمة الروايات
const novels = await apiFetch('/novels');

// الحصول على تفاصيل رواية
const novel = await apiFetch('/novels/:id');
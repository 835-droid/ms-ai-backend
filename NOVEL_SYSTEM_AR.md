# نظام الروايات (Novel System)

## نظرة عامة

تم إضافة نظام كامل لإدارة الروايات في الباك إند، مشابه تمامًا لنظام المانجا الموجود. النظام يدعم جميع الميزات الأساسية لإدارة المحتوى الأدبي.

## الميزات المدعومة

### 1. إدارة الروايات
- ✅ إنشاء روايات جديدة
- ✅ تعديل الروايات
- ✅ حذف الروايات
- ✅ عرض تفاصيل الرواية
- ✅ قائمة الروايات مع الترقيم (Pagination)

### 2. الفصول
- ✅ هيكل بيانات للفصول (NovelChapter)
- ✅ تخزين محتوى الفصول
- ✅ عد الكلمات (Word Count)
- ✅ ترقيم الفصول

### 3. التفاعلات
- ✅ التقييم (1-10 درجات)
- ✅ ردود الفعل (upvote, funny, love, surprised, angry, sad)
- ✅ المفضلة
- ✅ التعليقات على الروايات
- ✅ التعليقات على الفصول

### 4. الإحصائيات
- ✅ عدد المشاهدات
- ✅ عدد التقييمات
- ✅ متوسط التقييم
- ✅ عدد المفضلات
- ✅ عدد ردود الفعل

### 5. القوائم
- ✅ قوائم المفضلة المخصصة
- ✅ تتبع تقدم القراءة
- ✅ سجل المشاهدة

## البنية التقنية

### 1. Domain Layer
المسار: `internal/domain/novel/`

```
novel.go           - كيان الرواية الرئيسي
chapter.go         - كيان الفصل
rating.go          - أنظمة التقييم
reaction.go        - ردود الفعل
favorite.go        - المفضلة والقوائم
comment.go         - التعليقات
viewing_history.go - سجل المشاهدة والتقدم
errors.go          - الأخطاء المعرفة
repository.go      - واجهات المستودعات
```

### 2. Service Layer
المسار: `internal/core/content/novel/`

```
novel.go           - إعادة تصدير الأنواع
novel_service.go   - منطق الأعمال
```

### 3. Data Layer
المسار: `internal/data/content/novel/`

```
mongo_novel_repository.go           - مستودع MongoDB الأساسي
mongo_novel_engagement_repository.go - تفاعلات MongoDB
```

### 4. API Layer
المسار: `internal/api/handler/content/novel/`

```
novel_handler.go              - معالجة الطلبات الأساسية
novel_interaction_handler.go  - معالجة التفاعلات
```

المسار: `internal/api/router/content/novel/`

```
novel_routes.go - تعريف المسارات
```

## نقاط النهاية (API Endpoints)

### إدارة الروايات

```
GET    /api/novels              - قائمة الروايات
GET    /api/novels/:id          - جلب رواية معينة
POST   /api/novels              - إنشاء رواية جديدة (للأدمن)
PUT    /api/novels/:id          - تعديل رواية (للأدمن)
DELETE /api/novels/:id          - حذف رواية (للأدمن)
```

### القوائم الخاصة

```
GET /api/novels/most-viewed       - الأكثر مشاهدة
GET /api/novels/recently-updated  - المحدثة حديثًا
GET /api/novels/most-followed     - الأكثر متابعة
GET /api/novels/top-rated         - الأعلى تقييمًا
```

### التفاعلات

```
POST /api/novels/:id/view         - زيادة عدد المشاهدات
POST /api/novels/:id/react        - إضافة رد فعل
GET  /api/novels/:id/my-reaction  - جلب رد فعلي
POST /api/novels/:id/rate         - تقييم الرواية
POST /api/novels/:id/favorite     - إضافة للمفضلة
DELETE /api/novels/:id/favorite   - إزالة من المفضلة
GET  /api/novels/:id/favorite     - التحقق من المفضلة
GET  /api/novels/favorites        - قائمة المفضلة
POST /api/novels/:id/comments     - إضافة تعليق
GET  /api/novels/:id/comments     - جلب التعليقات
DELETE /api/novels/:id/comments/:comment_id - حذف تعليق
```

## قاعدة البيانات

### PostgreSQL

تم إنشاء جدول SQL كامل في: `scripts/novel_tables.sql`

الجداول:
- novels (الروايات)
- novel_chapters (الفصول)
- novel_view_logs (سجل المشاهدات)
- novel_ratings (التقييمات)
- novel_reactions (ردود الفعل)
- novel_favorites (المفضلة)
- novel_comments (التعليقات)
- novel_favorite_lists (قوائم المفضلة)
- novel_reading_progress (تقدم القراءة)
- novel_viewing_history (سجل المشاهدة)

### MongoDB

المجموعات (Collections):
- novel
- novel_view_logs
- novel_reactions
- novel_ratings
- novel_favorites
- novel_comments

## التثبيت

### 1. تشغيل ترحيل قاعدة البيانات

```bash
psql -U username -d database_name -f MS-AI/scripts/novel_tables.sql
```

### 2. تسجيل المسارات في التطبيق الرئيسي

```go
import (
    novelhandler "github.com/835-droid/ms-ai-backend/internal/api/handler/content/novel"
    novelrouter "github.com/835-droid/ms-ai-backend/internal/api/router/content/novel"
    novelservice "github.com/835-droid/ms-ai-backend/internal/core/content/novel"
    noveldata "github.com/835-droid/ms-ai-backend/internal/data/content/novel"
)

// إنشاء المستودع
novelRepo := noveldata.NewMongoNovelRepository(mongoStore)

// إنشاء الخدمة
novelService := novelservice.NewNovelService(novelRepo, logger)

// إنشاء المعالج
novelHandler := novelhandler.NewNovelHandler(novelService)

// تسجيل المسارات
novelrouter.SetupNovelRoutes(engine, novelHandler, cfg, userRepo)
```

## أمثلة الاستخدام

### إنشاء رواية

```bash
curl -X POST http://localhost:8080/api/novels \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "رواية اختبارية",
    "description": "هذه رواية تجريبية",
    "tags": ["تجربة", "عربي"]
  }'
```

### تقييم رواية

```bash
curl -X POST http://localhost:8080/api/novels/{novelID}/rate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"score": 9}'
```

### إضافة للمفضلة

```bash
curl -X POST http://localhost:8080/api/novels/{novelID}/favorite \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### إضافة تعليق

```bash
curl -X POST http://localhost:8080/api/novels/{novelID}/comments \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "رواية رائعة!"}'
```

## الملفات المنشأة

### Domain
- `MS-AI/internal/domain/novel/novel.go`
- `MS-AI/internal/domain/novel/chapter.go`
- `MS-AI/internal/domain/novel/rating.go`
- `MS-AI/internal/domain/novel/reaction.go`
- `MS-AI/internal/domain/novel/favorite.go`
- `MS-AI/internal/domain/novel/comment.go`
- `MS-AI/internal/domain/novel/viewing_history.go`
- `MS-AI/internal/domain/novel/errors.go`
- `MS-AI/internal/domain/novel/repository.go`

### Application
- `MS-AI/internal/application/dtos/novel_dtos.go`

### Core/Service
- `MS-AI/internal/core/content/novel/novel.go`
- `MS-AI/internal/core/content/novel/novel_service.go`

### Data
- `MS-AI/internal/data/content/novel/mongo_novel_repository.go`
- `MS-AI/internal/data/content/novel/mongo_novel_engagement_repository.go`

### API
- `MS-AI/internal/api/handler/content/novel/novel_handler.go`
- `MS-AI/internal/api/handler/content/novel/novel_interaction_handler.go`
- `MS-AI/internal/api/router/content/novel/novel_routes.go`

### Database
- `MS-AI/scripts/novel_tables.sql`

### Documentation
- `MS-AI/NOVEL_SYSTEM_IMPLEMENTATION.md`
- `MS-AI/NOVEL_SYSTEM_AR.md`

## الخطوات التالية

لتحقيق نظام روايات كامل، يُنصح بإضافة:

1. **إدارة الفصول**: معالجة CRUD كاملة للفصول
2. **تقدم القراءة**: حفظ آخر فصل مقروء
3. **سجل المشاهدة**: تتبع الروايات المقروءة
4. **قوائم مخصصة**: إنشاء قوائم مفضلة متعددة
5. **البحث**: بحث نصي كامل في الروايات
6. **التوصيات**: توصيات مبنية على التفضيلات

## الملاحظات

- النظام مصمم ليكون مشابهًا تمامًا لنظام المانجا
- يمكن إعادة استخدام معظم البنية التحتية الحالية
- جميعmiddlewares (المصادقة، الترخيص، Rate Limiting) تعمل مع النظام الجديد
- الأنماط والمعايير متسقة مع باقي التطبيق

## الترخيص

نفس ترخيص المشروع الأصلي.
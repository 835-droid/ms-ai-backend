# دليل دمج واجهة الويب مع نظام الروايات

## نظرة عامة

تم إضافة نظام الروايات بشكل كامل إلى الباك اند والفرونت اند. يوضح هذا الدليل كيفية تكامل صفحات الويب مع واجهات برمجة التطبيقات (APIs) الخاصة بالروايات.

## الصفحات المضافة

### 1. صفحة قائمة الروايات (`novels.html`)
- **المسار**: `MS-AI/cmd/web/novels.html`
- **الوصف**: تعرض قائمة الروايات المتاحة مع إمكانية البحث والفلترة
- **المميزات**:
  - عرض الروايات بنظام البطاقات
  - فلترة حسب التصنيفات (رومانسية، خيال، إثارة، إلخ)
  - ترتيب حسب (الأحدث، الأكثر شعبية، الأعلى تقييمًا)
  - ترقيم الصفحات

### 2. صفحة تفاصيل الرواية (`novel-details.html`)
- **المسار**: `MS-AI/cmd/web/novel-details.html`
- **الوصف**: تعرض تفاصيل الرواية مع الفصول والتعليقات
- **المميزات**:
  - عرض معلومات الرواية (العنوان، الوصف، التقييم، إلخ)
  - عرض الفصول المتاحة
  - إضافة تعليقات وتقييمات
  - إضافة الرواية للمفضلة
  - التفاعل مع الرواية (إعجاب، إلخ)
  - اقتراح روايات مشابهة

### 3. صفحة قراءة الرواية (`novel-reader.html`)
- **المسار**: `MS-AI/cmd/web/novel-reader.html`
- **الوصف**: واجهة قراءة الفصول
- **المميزات**:
  - عرض محتوى الفصل
  - التحكم في حجم الخط
  - التنقل بين الفصول
  - حفظ تقدم القراءة

### 4. لوحة تحكم الأدمن
- **المسار**: `MS-AI/cmd/web/admin.html`
- **التحديثات**:
  - قسم "إدارة الروايات" لعرض وتعديل الروايات
  - قسم "إضافة رواية" لإضافة روايات جديدة
  - إحصائيات الروايات في لوحة التحكم الرئيسية

## واجهات برمجة التطبيقات (APIs)

### الروايات الأساسية

```javascript
// جلب قائمة الروايات
API.getNovels(page, limit)

// جلب تفاصيل رواية
API.getNovel(id)

// جلب الروايات الأكثر مشاهدة
API.getMostViewedNovels(period, limit)

// جلب الروايات المحدثة حديثاً
API.getRecentlyUpdatedNovels(limit)

// جلب الروايات الأعلى تقييماً
API.getTopRatedNovels(limit)
```

### التفاعل مع الروايات

```javascript
// زيادة عدد المشاهدات
API.incrementNovelViews(id)

// إضافة تفاعل (إعجاب، إلخ)
API.setNovelReaction(id, type)

// الحصول على تفاعل المستخدم
API.getNovelUserReaction(id)

// تقييم الرواية
API.rateNovel(id, score)

// إضافة للمفضلة
API.addNovelFavorite(id)

// إزالة من المفضلة
API.removeNovelFavorite(id)

// التحقق من حالة المفضلة
API.checkNovelFavorite(id)

// جلب قائمة المفضلة
API.listNovelFavorites(page, limit)
```

### التعليقات

```javascript
// إضافة تعليق
API.addNovelComment(id, content)

// جلب التعليقات
API.getNovelComments(id, page, limit, sort)

// حذف تعليق
API.deleteNovelComment(id, commentId)
```

### تقدم القراءة

```javascript
// جلب تقدم القراءة
API.getNovelReadingProgress(id)
```

## قاعدة البيانات

### جداول الروايات (PostgreSQL)

تم إنشاء الجداول التالية في `scripts/novel_tables.sql`:

- `novels` - جدول الروايات الرئيسي
- `novel_chapters` - جدول فصول الروايات
- `novel_ratings` - جدول تقييمات الروايات
- `novel_favorites` - جدول المفضلة
- `novel_comments` - جدول التعليقات
- `novel_reactions` - جدول التفاعلات
- `novel_views` - جدول المشاهدات
- `novel_reading_progress` - جدول تقدم القراءة

### مستودعات البيانات (MongoDB)

تم إنشاء المستودعات التالية في `internal/data/content/novel/`:

- `mongo_novel_repository.go` - مستودع الروايات
- `mongo_novel_engagement_repository.go` - مستودع التفاعلات

## تشغيل التطبيق

### 1. تشغيل قاعدة البيانات

```bash
# تشغيل PostgreSQL و MongoDB
docker-compose up -d
```

### 2. تشغيل الترحيل

```bash
# ترحيل جداول الروايات
psql -U postgres -d msai_db -f scripts/novel_tables.sql
```

### 3. تشغيل التطبيق

```bash
cd MS-AI
go run cmd/web/main.go
```

### 4. الوصول للواجهة

- الرئيسية: `http://localhost:8080`
- الروايات: `http://localhost:8080/novels.html`
- لوحة الأدمن: `http://localhost:8080/admin.html`

## التحديثات المستقبلية

### 1. إضافة فصول الروايات
- تطوير واجهة لإضافة فصول الروايات
- دعم تنسيقات النصوص المختلفة (TXT, EPUB, PDF)

### 2. تحسين البحث
- إضافة بحث متقدم في الروايات
- فلترة حسب المؤلف والتصنيفات

### 3. الإشعارات
- إشعار عند إضافة فصل جديد
- إشعار عند وجود تعليقات جديدة

### 4. الإحصائيات
- لوحة إحصائيات متقدمة للروايات
- رسوم بيانية لتطور القراءة

## الدعم الفني

للدعم والاستفسارات:
- GitHub Issues: [MS-AI Issues](https://github.com/your-repo/ms-ai/issues)
- Discord: [MS-AI Discord](https://discord.gg/6tTdAh4JVd)

## الترخيص

هذا المشروع مرخص بموجب ترخيص MIT.
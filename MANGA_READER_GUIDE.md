# 🎯 قارئ المانجا المتطور - Manga Reader

## نظرة عامة

قارئ مانجا احترافي بسيط وفعال مصمم خصيصاً لصفحات القراءة. يوفر تجربة قراءة سلسة مع ميزات متقدمة دون تعقيد غير ضروري.

## ✨ الميزات الأساسية

### 📖 تجربة القراءة
- **تنقل سلس**: أزرار، لوحة المفاتيح، واللمس
- **حفظ التقدم**: استئناف القراءة من حيث توقفت
- **شاشة كاملة**: وضع القراءة المركز
- **خلفيات متعددة**: أسود، أبيض، رمادي

### 🎨 خيارات العرض
- **ملء العرض**: `fit-width`
- **ملء الارتفاع**: `fit-height`
- **تلقائي**: `fit-auto`
- **تخصيص الخلفية**: للراحة البصرية

### ⌨️ التحكم المتقدم
```
التنقل:
• ← → : الصفحات السابقة/التالية
• Space : الصفحة التالية
• F : شاشة كاملة
• S : إعدادات
• C : قائمة الفصول
• Esc : خروج من الشاشة الكاملة

اللمس:
• النقر على اليسار: الصفحة السابقة
• النقر على اليمين: الصفحة التالية
• السحب لليسار: الصفحة التالية
• السحب لليمين: الصفحة السابقة
```

## 🏗️ البنية التقنية

### الملفات الأساسية:
```
cmd/web/
├── manga-reader.html    # الصفحة الرئيسية
├── manga-reader.js      # منطق القارئ
├── reader.css          # التصميم المخصص
├── common.js           # الوظائف المشتركة
└── style.css           # التصميم العام
```

### الهيكل التقني:
```javascript
class MangaReader {
  constructor() {
    this.currentPage = 1;
    this.totalPages = 0;
    this.fitMode = 'width';
    this.backgroundMode = 'black';
  }

  // الوظائف الأساسية
  loadMangaData()     // تحميل بيانات المانجا
  loadPage()         // تحميل صفحة معينة
  nextPage()         // الانتقال للصفحة التالية
  prevPage()         // الانتقال للصفحة السابقة
  saveProgress()     // حفظ التقدم
  loadProgress()     // تحميل التقدم المحفوظ
}
```

## 🎯 كيفية الاستخدام

### 1. الوصول للقارئ:
```html
<!-- من صفحة تفاصيل المانجا -->
<a href="manga-reader.html?manga=MANGA_ID&chapter=CHAPTER_ID">
    قراءة الفصل
</a>
```

### 2. المعاملات المطلوبة:
- `manga`: معرف المانجا
- `chapter`: معرف الفصل

### 3. مثال على الاستخدام:
```javascript
// في JavaScript
function openChapter(chapterId, chapterTitle, chapterNumber) {
    const mangaId = getMangaIdFromUrl();
    window.location.href = `manga-reader.html?manga=${mangaId}&chapter=${chapterId}`;
}
```

## 💾 التخزين المحلي

### حفظ الإعدادات:
```javascript
// حفظ وضع العرض
localStorage.setItem('manga-reader-settings', JSON.stringify({
  fitMode: 'width',
  backgroundMode: 'black'
}));

// حفظ التقدم
localStorage.setItem(`manga-progress-${mangaId}`, JSON.stringify({
  chapterId: chapterId,
  page: currentPage,
  timestamp: Date.now()
}));
```

### استرجاع البيانات:
```javascript
const settings = JSON.parse(localStorage.getItem('manga-reader-settings') || '{}');
const progress = JSON.parse(localStorage.getItem(`manga-progress-${mangaId}`) || 'null');
```

## 🎨 التخصيص

### إضافة خلفية جديدة:
```css
.reader-container.bg-custom {
  background: linear-gradient(45deg, #ff6b6b, #4ecdc4);
  color: white;
}
```

### تخصيص أزرار التنقل:
```css
.btn-custom {
  background: linear-gradient(45deg, #667eea, #764ba2);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}
```

## 📱 الاستجابة للأجهزة

### الهواتف المحمولة:
- إخفاء نص الأزرار لتوفير المساحة
- تحسين منطقة اللمس
- تبديل تلقائي للوضع العمودي

### الأجهزة اللوحية:
- أزرار أكبر لسهولة اللمس
- تحسين سرعة الاستجابة
- دعم الشاشة الكاملة

## 🔧 التكامل مع API

### نقاط النهاية المطلوبة:
```javascript
// تحميل بيانات الفصل
GET /api/manga-chapters/{chapterId}

// تحميل قائمة الفصول
GET /api/mangas/{mangaId}/chapters
```

### تنسيق البيانات المتوقع:
```json
{
  "id": "chapter-123",
  "mangaTitle": "عنوان المانجا",
  "number": 1,
  "title": "عنوان الفصل",
  "pages": [
    "https://example.com/page1.jpg",
    "https://example.com/page2.jpg"
  ],
  "createdAt": "2024-01-01T00:00:00Z"
}
```

## 🚀 الميزات المستقبلية

### قصير المدى:
- [ ] دعم PWA للقراءة offline
- [ ] تحسين تحميل الصور (lazy loading)
- [ ] إضافة مؤشرات التحميل
- [ ] تحسين الأداء على الأجهزة القديمة

### متوسط المدى:
- [ ] دعم تنزيل الفصول
- [ ] إضافة علامات مرجعية
- [ ] تتبع القراءة الإحصائي
- [ ] تخصيص سرعة الانتقال

### طويل المدى:
- [ ] دعم قراءة مزدوجة الصفحات
- [ ] تكامل مع خدمات خارجية
- [ ] دعم الترجمة
- [ ] نظام تقييم وتعليقات

## 🐛 استكشاف الأخطاء

### مشاكل شائعة:

#### 1. عدم تحميل الصور:
```javascript
// تحقق من صحة URLs
console.log('Page URLs:', chapterData.pages);

// تحقق من CORS
// تأكد من أن الخادم يدعم CORS
```

#### 2. عدم حفظ التقدم:
```javascript
// تحقق من localStorage
console.log('Progress saved:', localStorage.getItem(`manga-progress-${mangaId}`));

// تحقق من وجود mangaId
console.log('Manga ID:', this.mangaId);
```

#### 3. مشاكل اللمس:
```javascript
// اختبار أحداث اللمس
document.addEventListener('touchstart', (e) => {
  console.log('Touch started:', e.touches[0]);
});
```

## 📊 الأداء

### المقاييس المستهدفة:
- **وقت التحميل الأولي**: < 2 ثانية
- **انتقال الصفحات**: < 300ms
- **حجم الحزمة**: < 50KB (gzip)
- **استخدام الذاكرة**: < 10MB

### التحسينات المطبقة:
- تحميل الصور عند الحاجة
- تخزين مؤقت للإعدادات
- تحسين أحداث اللمس
- تقليل إعادة الرسم

## 🎉 الخلاصة

قارئ مانجا احترافي يركز على تجربة القراءة المثالية مع الحفاظ على البساطة والأداء العالي. مصمم خصيصاً لصفحات القراءة دون تعقيدات غير ضرورية.

---

**🚀 جاهز للاستخدام الآن!**
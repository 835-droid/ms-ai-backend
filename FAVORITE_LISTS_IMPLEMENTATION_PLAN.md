# خطة تطوير نظام القوائم المخصصة للمفضلة

## 📋 نظرة عامة

نظام القوائم المخصصة يسمح للمستخدمين بإنشاء قوائم مفضلة متعددة، وإضافة المانجا إليها، مع إمكانية إضافة ملاحظات وترتيب المانجا داخل كل قائمة.

## 🏗️ البنية المعمارية

### 1. قاعدة البيانات
- `favorite_lists` - جداول القوائم
- `favorite_list_items` - عناصر القوائم (المانجا في كل قائمة)

### 2. Backend (Go)
- **Models**: `FavoriteList`, `FavoriteListItem`
- **Repository Interface**: `FavoriteListRepository`
- **Repository Implementation**: `postgres_favorite_list_repository.go`
- **Service Interface**: `FavoriteListService`
- **Service Implementation**: `favorite_list_service.go`
- **Handlers**: إضافات إلى `manga_interaction_handler.go`
- **Routes**: إضافات إلى `manga_routes.go`

### 3. Frontend (JavaScript)
- **API Functions**: إضافات إلى `api.js`
- **UI Components**: إضافات إلى `favorites.html` و `manga-details.html`
- **Logic**: إضافات إلى `favorites.js` و `manga-details-enhanced.js`

## 🚀 خطوات التنفيذ

### المرحلة 1: Repository Layer
1. إنشاء ملف `internal/data/content/manga/postgres_favorite_list_repository.go`
2. تنفيذ جميع دوال `FavoriteListRepository`

### المرحلة 2: Service Layer
1. إنشاء ملف `internal/core/content/manga/favorite_list_service.go`
2. تعريف واجهة `FavoriteListService`
3. تنفيذ الدوال

### المرحلة 3: API Handlers
1. إضافة دوال جديدة إلى `manga_interaction_handler.go`:
   - `CreateFavoriteList`
   - `GetFavoriteList`
   - `ListMyFavoriteLists`
   - `UpdateFavoriteList`
   - `DeleteFavoriteList`
   - `AddMangaToList`
   - `RemoveMangaFromList`
   - `MoveMangaToList`

### المرحلة 4: Routes
1. تحديث `manga_routes.go` لإضافة الراوتات الجديدة

### المرحلة 5: Container Setup
1. تحديث `internal/container/repo_initializers.go`
2. تحديث `internal/container/service_initializers.go`
3. تحديث `internal/container/handler_initializers.go`

### المرحلة 6: Frontend
1. تحديث `favorites.html` لعرض القوائم
2. تحديث `favorites.js` لإضافة منطق إدارة القوائم
3. تحديث `manga-details-enhanced.js` لإضافة modal اختيار القوائم

## 📝 API Endpoints

### إدارة القوائم
- `POST /api/mangas/lists` - إنشاء قائمة جديدة
- `GET /api/mangas/lists` - الحصول على قوائم المستخدم
- `GET /api/mangas/lists/:listID` - الحصول على تفاصيل قائمة
- `PUT /api/mangas/lists/:listID` - تحديث قائمة
- `DELETE /api/mangas/lists/:listID` - حذف قائمة

### إدارة عناصر القوائم
- `POST /api/mangas/lists/:listID/items` - إضافة مانجا للقائمة
- `DELETE /api/mangas/lists/:listID/items/:mangaID` - إزالة مانجا من القائمة
- `PUT /api/mangas/lists/:listID/items/:mangaID` - تحديث ملاحظات المانجا
- `POST /api/mangas/lists/:listID/items/:mangaID/move` - نقل مانجا لقائمة أخرى
- `GET /api/mangas/:mangaID/lists` - الحصول على قوائم المستخدم التي تحتوي المانجا

## 🗄️ هيكل البيانات

### FavoriteList
```json
{
  "id": "list-uuid",
  "user_id": "user-uuid",
  "name": "قراءة لاحقاً",
  "description": "مانجا أخطط لقراءتها",
  "is_public": false,
  "sort_order": 0,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "manga_count": 5
}
```

### FavoriteListItem
```json
{
  "list_id": "list-uuid",
  "manga_id": "manga-uuid",
  "notes": "مانجا رائعة أنصح بها",
  "added_at": "2024-01-01T00:00:00Z",
  "sort_order": 0,
  "manga": { /* Manga object */ }
}
```

## ✅ معايير القبول

1. [ ] يمكن للمستخدم إنشاء قوائم متعددة
2. [ ] يمكن إضافة مانجا لقوائم متعددة
3. [ ] يمكن إضافة ملاحظات لكل مانجا في القائمة
4. [ ] يمكن إعادة ترتيب المانجا داخل القائمة
5. [ ] يمكن نقل المانجا بين القوائم
6. [ ] يمكن جعل القوائم عامة أو خاصة
7. [ ] الواجهة تعرض القوائم بشكل منظم
8. [ ] يمكن حذف القوائم والعناصر

## 🔧 ملاحظات التنفيذ

- استخدام UUIDs للقوائم والعناصر
- التحقق من ملكية القائمة قبل التعديل
- استخدام معاملات لمنع race conditions
- إضافة indexes لقاعدة البيانات للأداء
- تطبيق pagination للقوائم الطويلة
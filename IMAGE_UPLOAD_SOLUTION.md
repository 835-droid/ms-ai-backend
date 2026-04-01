# حل مشكلة رفع الصور في MS-AI

## المشكلة الحالية
المشكلة الرئيسية في رفع الصور هي أن الخادم لا يحتوي على endpoint مخصص لرفع الملفات. الصور يتم تحويلها إلى base64 في المتصفح وتُحفظ كـ string في قاعدة البيانات.

## الحل المقترح

### 1. إضافة endpoint رفع الصور في الخادم

أضف endpoint جديد في `internal/api/router/router.go`:

```go
// في دالة Setup
r.POST("/api/upload/image", middleware.AuthMiddleware(cfg), handler.UploadImage)
```

### 2. إنشاء handler لرفع الصور

أنشئ `internal/api/handler/upload_handler.go`:

```go
package handler

import (
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
    return &UploadHandler{}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(400, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()

    // التحقق من نوع الملف
    if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
        c.JSON(400, gin.H{"error": "File must be an image"})
        return
    }

    // إنشاء اسم ملف فريد
    ext := filepath.Ext(header.Filename)
    filename := uuid.New().String() + ext

    // إنشاء المجلد إذا لم يكن موجوداً
    uploadDir := "./uploads/images"
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        c.JSON(500, gin.H{"error": "Failed to create upload directory"})
        return
    }

    // حفظ الملف
    dst := filepath.Join(uploadDir, filename)
    out, err := os.Create(dst)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to save file"})
        return
    }
    defer out.Close()

    if _, err := io.Copy(out, file); err != nil {
        c.JSON(500, gin.H{"error": "Failed to copy file"})
        return
    }

    // إرجاع رابط الصورة
    imageUrl := "/uploads/images/" + filename
    c.JSON(200, gin.H{
        "success": true,
        "data": gin.H{
            "image_url": imageUrl,
            "filename": filename
        }
    })
}
```

### 3. تحديث router.go

```go
// أضف الـ import
uploadHandler "github.com/835-droid/ms-ai-backend/internal/api/handler"

// في دالة Setup
uploadHandler := handler.NewUploadHandler()

// أضف المسار
r.POST("/api/upload/image", middleware.AuthMiddleware(cfg), uploadHandler.UploadImage)

// خدمة الملفات المرفوعة
r.Static("/uploads", "./uploads")
```

### 4. تحديث JavaScript

في `app.js`، حدث دالة `uploadImage`:

```javascript
// رفع الصورة للخادم
async function uploadImage(file) {
    const formData = new FormData();
    formData.append('image', file);

    const response = await fetch('/api/upload/image', {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${getAccessToken()}`
        },
        body: formData
    });

    const data = await response.json();

    if (!response.ok) {
        throw new Error(data.error || 'فشل في رفع الصورة');
    }

    return data.data.image_url;
}
```

## فوائد هذا الحل

1. **أداء أفضل**: الصور تُحفظ كملفات منفصلة بدلاً من base64
2. **توفير مساحة**: base64 يزيد حجم البيانات بنسبة 33%
3. **تحميل أسرع**: المتصفح يمكنه تخزين الصور مؤقتاً
4. **إدارة أسهل**: يمكن حذف وتعديل الصور بسهولة

## خطوات التنفيذ

1. أنشئ المجلد `uploads/images` في جذر المشروع
2. أضف الكود المذكور أعلاه
3. اختبر رفع الصور
4. أضف ضغط الصور إذا لزم الأمر
5. أضف حذف الصور غير المستخدمة

## ملاحظات أمنية

- تحقق من نوع الملف قبل الرفع
- حدد حجم الملف الأقصى
- استخدم UUID لتجنب تضارب الأسماء
- أضف مصادقة للمستخدمين المصرح لهم فقط
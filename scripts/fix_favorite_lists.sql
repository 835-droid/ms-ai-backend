-- ==========================================
-- إصلاح جدول favorite_lists المفقود
-- قم بتشغيل هذا الملف على قاعدة البيانات لإضافة الجداول الناقصة
-- ==========================================

-- جدول قوائم المفضلة المخصصة
CREATE TABLE IF NOT EXISTS favorite_lists (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- جدول عناصر قوائم المفضلة (ربط بين القوائم والمانجا)
CREATE TABLE IF NOT EXISTS favorite_list_items (
    list_id TEXT NOT NULL REFERENCES favorite_lists(id) ON DELETE CASCADE,
    manga_id TEXT NOT NULL REFERENCES mangas(id) ON DELETE CASCADE,
    notes TEXT,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sort_order INTEGER DEFAULT 0,
    PRIMARY KEY (list_id, manga_id)
);

-- Indexes for favorite_lists
CREATE INDEX IF NOT EXISTS idx_favorite_lists_user_id ON favorite_lists(user_id);
CREATE INDEX IF NOT EXISTS idx_favorite_lists_is_public ON favorite_lists(is_public);

-- Indexes for favorite_list_items
CREATE INDEX IF NOT EXISTS idx_favorite_list_items_list_id ON favorite_list_items(list_id);
CREATE INDEX IF NOT EXISTS idx_favorite_list_items_manga_id ON favorite_list_items(manga_id);
CREATE INDEX IF NOT EXISTS idx_favorite_list_items_sort_order ON favorite_list_items(sort_order);

-- رسالة تأكيد
DO $$
BEGIN
    RAISE NOTICE 'تم إنشاء جداول favorite_lists و favorite_list_items بنجاح!';
END $$;
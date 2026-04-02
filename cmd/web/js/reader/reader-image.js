// reader-image.js - معالجة الصور والتحميل المسبق

// تطبيع روابط الصور
function normalizeChapterPages(chapter) {
    if (!chapter) return chapter;
    const pages = Array.isArray(chapter.pages) ? chapter.pages : [];
    chapter.pages = pages.map(normalizeImageUrl).filter(Boolean);
    return chapter;
}

// التحقق من نوع الرابط
function isAbsoluteHttpUrl(url) {
    return /^https?:\/\//i.test(String(url || ''));
}

function isDataUrl(url) {
    return String(url || '').startsWith('data:');
}

// حل رابط الصورة (مع Proxy)
function resolveReaderImageUrl(url) {
    const normalized = normalizeImageUrl(url);
    if (!normalized) return '';

    if (normalized.startsWith('/uploads/') || normalized.startsWith('/static/')) {
        if (normalized.startsWith('/')) {
            return `${CONFIG.BACKEND_URL}${normalized}`;
        }
        return normalized;
    }

    if (isAbsoluteHttpUrl(normalized)) {
        return `${CONFIG.API_BASE}/assets/image-proxy?url=${encodeURIComponent(normalized)}`;
    }

    if (isDataUrl(normalized)) {
        return normalized;
    }

    return normalized;
}

// إعادة محاولة تحميل الصورة عند الفشل
function retryImageLoad(imgElement, originalUrl, onComplete, maxRetries = 2, delay = 1000) {
    let retries = 0;
    
    const retry = () => {
        if (retries >= maxRetries) {
            imgElement.alt = `❌ فشل تحميل الصورة: ${originalUrl}`;
            const errorDiv = document.getElementById('reader-error');
            if (errorDiv) {
                errorDiv.textContent = `⚠️ فشل تحميل الصورة بعد عدة محاولات: ${originalUrl.substring(0, 50)}...`;
            }
            if (onComplete) onComplete();
            return;
        }
        retries++;
        setTimeout(() => {
            imgElement.src = resolveReaderImageUrl(originalUrl);
        }, delay);
    };
    
    imgElement.onerror = retry;
    imgElement.onload = () => {
        imgElement.onerror = null;
        const errorDiv = document.getElementById('reader-error');
        if (errorDiv && errorDiv.textContent.includes(originalUrl)) {
            errorDiv.textContent = '';
        }
        if (onComplete) onComplete();
    };
}

// تحميل مسبق للصورة التالية (لتحسين الأداء)
function prefetchImage(url) {
    if (!readerState.prefetchEnabled) return;
    const img = new Image();
    img.src = resolveReaderImageUrl(normalizeImageUrl(url));
}

// تحميل مسبق للصفحات القادمة
function prefetchNextPages(currentIndex, pages, count = 2) {
    if (!readerState.prefetchEnabled) return;
    
    for (let i = 1; i <= count; i++) {
        const nextIndex = currentIndex + i;
        if (nextIndex < pages.length) {
            prefetchImage(pages[nextIndex]);
        }
    }
}

// تحميل جميع صور الفصل (للويب توون)
function prefetchAllChapterImages(chapter) {
    if (!readerState.prefetchEnabled || !chapter?.pages) return;
    
    const pages = chapter.pages;
    pages.forEach(page => {
        prefetchImage(page);
    });
}
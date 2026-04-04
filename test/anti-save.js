// حفظ هذا الملف في public/js/anti-save.js
(function() {
  // تخطي الحماية في وضع التطوير المحلي (localhost)
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    console.log('وضع التطوير المحلي - تم تخطي حماية المحتوى');
    return;
  }

  // منع حفظ الصفحة
  document.addEventListener('keydown', function(e) {
    // Ctrl+S / Command+S
    if (e.key === 's' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      alert('حفظ الصفحة غير مسموح به');
      return false;
    }
    
    // Ctrl+P / Command+P (طباعة/حفظ كملف PDF)
    if (e.key === 'p' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      alert('طباعة الصفحة غير مسموح بها');
      return false;
    }
    
    // Ctrl+Shift+I / Command+Option+I (أدوات المطور)
    if ((e.key === 'i' || e.key === 'I') && (e.ctrlKey || e.metaKey) && e.shiftKey) {
      e.preventDefault();
      return false;
    }
  });
  
  // منع النقر بالزر الأيمن
  document.addEventListener('contextmenu', function(e) {
    e.preventDefault();
    return false;
  });
  
  // منع نسخ النص
  document.addEventListener('copy', function(e) {
    e.preventDefault();
    return false;
  });
  
  // منع الوصول من بروتوكول file:// (الصفحات المحفوظة)
  if (window.location.protocol === 'file:') {
    document.body.innerHTML = '<div style="text-align: center; padding: 50px; font-family: Arial;">' +
      '<h1>غير مسموح الوصول</h1>' +
      '<p>لا يمكن الوصول إلى هذا المحتوى من صفحة محفوظة.</p>' +
      '<p>يرجى زيارة <a href="https://realmnovel.com">realmnovel.com</a> لعرض المحتوى.</p>' +
      '</div>';
  }
  
  // الكشف عن تحميل الصفحة من ذاكرة التخزين المؤقت
  window.addEventListener('load', function() {
    // تخزين رمز مميز في sessionStorage
    if (!sessionStorage.getItem('pageToken')) {
      sessionStorage.setItem('pageToken', Date.now().toString());
    }
    
    // التحقق مما إذا كنا في وضع عدم الاتصال ولكن ليس في وضع عدم الاتصال الرسمي للموقع
    if (!navigator.onLine && !window.location.href.includes('offline=true')) {
      // قد تكون هذه صفحة محفوظة يتم عرضها دون اتصال
      document.body.innerHTML = '<div style="text-align: center; padding: 50px; font-family: Arial;">' +
        '<h1>غير مسموح الوصول في وضع عدم الاتصال</h1>' +
        '<p>لا يمكن عرض هذا المحتوى بدون اتصال بالإنترنت إلا من خلال التطبيق الرسمي.</p>' +
        '<p>يرجى زيارة <a href="https://realmnovel.com">realmnovel.com</a> عند توفر الاتصال.</p>' +
        '</div>';
    }
  });
  
  // تعطيل وظيفة الطباعة
  window.addEventListener('beforeprint', function(e) {
    e.preventDefault();
    alert('طباعة الصفحة غير مسموح بها');
    return false;
  });
})();
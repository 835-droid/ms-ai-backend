(() => {
    let INIT_CALLED = false;
    
    // دالة منع تحديد النص ونسخه مع السماح بذلك في حقول الإدخال
    function preventTextSelection() {
        // إضافة CSS لمنع تحديد النص مع استثناء حقول الإدخال
        const styleElement = document.createElement('style');
        styleElement.textContent = `
            *:not(input):not(textarea):not([contenteditable="true"]) {
                -webkit-user-select: none !important;
                -moz-user-select: none !important;
                -ms-user-select: none !important;
                user-select: none !important;
                -webkit-touch-callout: none !important;
            }

            /* منع إظهار تحديد النص باستثناء حقول الإدخال */
            :not(input):not(textarea):not([contenteditable="true"])::selection {
                background: transparent !important;
                color: inherit !important;
            }

            :not(input):not(textarea):not([contenteditable="true"])::-moz-selection {
                background: transparent !important;
                color: inherit !important;
            }
            
            /* تعطيل سحب الصور */
            img {
                pointer-events: none !important;
                -webkit-user-drag: none !important;
                -khtml-user-drag: none !important;
                -moz-user-drag: none !important;
                -o-user-drag: none !important;
            }
        `;
        document.head.appendChild(styleElement);

        // منع النسخ والقص واللصق عبر الضغط على مفاتيح مع ctrl باستثناء حقول الإدخال
        document.addEventListener('copy', function(e) {
            if (!isInputField(e.target)) {
                e.preventDefault();
                return false;
            }
        }, true);

        document.addEventListener('cut', function(e) {
            if (!isInputField(e.target)) {
                e.preventDefault();
                return false;
            }
        }, true);

        document.addEventListener('paste', function(e) {
            if (!isInputField(e.target)) {
                e.preventDefault();
                return false;
            }
        }, true);

        // منع السحب والإفلات باستثناء حقول الإدخال
        document.addEventListener('dragstart', function(e) {
            if (!isInputField(e.target)) {
                e.preventDefault();
                return false;
            }
        }, true);
    }

    // دالة للتحقق مما إذا كان العنصر حقل إدخال
    function isInputField(element) {
        const tagName = element.tagName ? element.tagName.toLowerCase() : '';
        const isContentEditable = element.isContentEditable || element.getAttribute('contenteditable') === 'true';
        return tagName === 'input' || tagName === 'textarea' || isContentEditable;
    }

    // معالجة مفاتيح التحكم (Ctrl و F12) مع السماح باستخدام Ctrl في حقول الإدخال
    function handleKeyControls() {
        document.addEventListener('keydown', function(event) {
            // السماح باستخدام Ctrl في حقول الإدخال
            if (event.ctrlKey && !isInputField(event.target)) {
                event.preventDefault();
                event.stopPropagation();
                return false;
            }
            
            // تعطيل F12
            if (event.key === 'F12' || event.keyCode === 123) {
                event.preventDefault();
                event.stopPropagation();
                return false;
            }
        }, true);
    }

    // كشف أدوات المطور - بدون إعادة تحميل الصفحة
    function detectDevTools() {
        const widthThreshold = window.outerWidth - window.innerWidth > 160;
        const heightThreshold = window.outerHeight - window.innerHeight > 160;
        
        // طرق إضافية للكشف - بدون إعادة تحميل الصفحة
        const devToolsOpen = widthThreshold || heightThreshold;
        
        // تم إزالة أي كود قد يؤدي إلى إعادة تحميل الصفحة
    }
    
    // تفعيل الحماية
    window.addEventListener('load', () => {
        // تفعيل الحماية ضد النسخ وتحديد النص
        preventTextSelection();
        // تفعيل الحماية ضد مفتاح Ctrl و F12
        handleKeyControls();
        // بدء فحص دوري لأدوات المطور
        setInterval(detectDevTools, 1000);
    });

    // تهيئة النظام
    window.initProtectionSystem = function(apiKey) {
        if (INIT_CALLED) return;
        INIT_CALLED = true;

        if (!apiKey) {
            console.error('API key is required');
            return;
        }

        preventTextSelection();
        handleKeyControls();
        setInterval(detectDevTools, 1000);
    };

    Object.defineProperty(window, 'initProtectionSystem', {
        writable: false,
        configurable: false
    });

    // تفعيل الحماية فوراً حتى قبل استدعاء init
    preventTextSelection();
    handleKeyControls();
    setInterval(detectDevTools, 1000);
})();
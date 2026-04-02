// toast.js - نظام إشعارات غير مزعج
let toastContainer = null;

function initToast() {
    if (toastContainer) return;
    toastContainer = document.createElement('div');
    toastContainer.id = 'toast-container';
    toastContainer.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        z-index: 9999;
        display: flex;
        flex-direction: column;
        gap: 10px;
        pointer-events: none;
        direction: rtl;
    `;
    document.body.appendChild(toastContainer);
}

function showToast(message, type = 'info', duration = 3000) {
    if (!toastContainer) initToast();
    
    const toast = document.createElement('div');
    const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️',
        info: 'ℹ️'
    };
    const icon = icons[type] || icons.info;
    
    toast.style.cssText = `
        background: ${type === 'error' ? 'rgba(239,68,68,0.95)' : type === 'success' ? 'rgba(16,185,129,0.95)' : type === 'warning' ? 'rgba(245,158,11,0.95)' : 'rgba(15,23,42,0.95)'};
        color: white;
        padding: 12px 20px;
        border-radius: 12px;
        font-size: 0.875rem;
        font-weight: 500;
        backdrop-filter: blur(8px);
        box-shadow: 0 10px 25px -5px rgba(0,0,0,0.2);
        transform: translateX(0);
        transition: transform 0.3s ease, opacity 0.3s ease;
        display: flex;
        align-items: center;
        gap: 10px;
        pointer-events: auto;
        direction: rtl;
        font-family: inherit;
    `;
    toast.innerHTML = `<span style="font-size:1.2rem">${icon}</span> <span>${message}</span>`;
    
    toastContainer.appendChild(toast);
    
    // تأثير الدخول
    setTimeout(() => {
        toast.style.transform = 'translateX(0)';
    }, 10);
    
    // الإزالة التلقائية
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(20px)';
        setTimeout(() => toast.remove(), 300);
    }, duration);
}

// استبدال showMessage القديمة تدريجياً
function showMessageWithToast(elementId, message, type) {
    // للتوافق مع الكود القديم
    const oldBox = document.getElementById(elementId);
    if (oldBox) {
        oldBox.textContent = message;
        oldBox.className = `message ${type}`;
        oldBox.style.display = 'block';
        setTimeout(() => {
            oldBox.style.display = 'none';
        }, 4000);
    }
    showToast(message, type);
}
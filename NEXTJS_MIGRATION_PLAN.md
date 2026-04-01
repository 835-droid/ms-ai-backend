# 🚀 خطة تحويل الواجهة الأمامية إلى Next.js

## الهدف: تحويل المشروع إلى مستوى الشركات الكبرى

## 📋 المراحل التفصيلية

### المرحلة 1: الإعداد والتخطيط (يوم 1-2)

#### 1.1 إنشاء مشروع Next.js
```bash
# إنشاء مجلد frontend بجانب cmd/web
cd /media/msai/5c595286-cb2f-4ad7-a578-36b2a8395839/go/MS-AI/backend/MS-AI

# إنشاء مشروع Next.js
npx create-next-app@latest frontend --typescript --tailwind --app --src-dir --import-alias "@/*"

cd frontend
```

#### 1.2 تثبيت الحزم المطلوبة
```bash
npm install zustand @tanstack/react-query lucide-react @headlessui/react
npm install -D @testing-library/react @testing-library/jest-dom @testing-library/user-event
npm install -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin prettier
```

#### 1.3 إعداد الملفات الأساسية
```typescript
// lib/store/auth.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: string
  username: string
  role: 'user' | 'admin'
}

interface AuthState {
  user: User | null
  token: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      login: async (username: string, password: string) => {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password })
        })
        const data = await response.json()
        set({ user: data.user, token: data.token })
      },
      logout: () => set({ user: null, token: null })
    }),
    { name: 'auth-storage' }
  )
)
```

### المرحلة 2: تحويل التصميم (يوم 3-5)

#### 2.1 نسخ CSS إلى Next.js
```typescript
// styles/globals.css
import './variables.css'
import './components.css'
import './utilities.css'

// styles/variables.css (نسخ من style.css الحالي)
:root {
  --primary-color: #6366f1;
  --primary-dark: #4f46e5;
  // ... باقي المتغيرات
}
```

#### 2.2 إنشاء Layout أساسي
```typescript
// app/layout.tsx
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'MS-AI - نظام إدارة المانجا',
  description: 'نظام إدارة المانجا المتطور',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ar" dir="rtl">
      <body className={inter.className}>
        {children}
      </body>
    </html>
  )
}
```

#### 2.3 تحويل صفحة تسجيل الدخول
```typescript
// app/login/page.tsx
'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/store/auth'

export default function LoginPage() {
  const [isLogin, setIsLogin] = useState(true)
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    inviteCode: ''
  })
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)

  const { login } = useAuthStore()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setMessage('')

    try {
      if (isLogin) {
        await login(formData.username, formData.password)
        router.push('/dashboard')
      } else {
        // Signup logic
        const response = await fetch('/api/auth/signup', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData)
        })
        if (response.ok) {
          setMessage('تم التسجيل بنجاح! يمكنك الآن تسجيل الدخول')
          setIsLogin(true)
        }
      }
    } catch (error) {
      setMessage('حدث خطأ في العملية')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container">
      <div className="auth-card">
        <h2>MS-AI</h2>
        <p style={{ color: 'var(--text-secondary)', textAlign: 'center', marginBottom: '2rem' }}>
          نظام إدارة المانجا المتطور
        </p>

        {message && (
          <div className={`message ${message.includes('خطأ') ? 'error' : 'success'}`}>
            {message}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <h3>{isLogin ? 'تسجيل الدخول' : 'تسجيل حساب جديد'}</h3>

          <div className="form-group">
            <label htmlFor="username">اسم المستخدم</label>
            <input
              type="text"
              id="username"
              value={formData.username}
              onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
              placeholder="أدخل اسم المستخدم"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">كلمة المرور</label>
            <input
              type="password"
              id="password"
              value={formData.password}
              onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
              placeholder={isLogin ? "أدخل كلمة المرور" : "اختر كلمة مرور قوية"}
              required
            />
          </div>

          {!isLogin && (
            <div className="form-group">
              <label htmlFor="inviteCode">رمز الدعوة</label>
              <input
                type="text"
                id="inviteCode"
                value={formData.inviteCode}
                onChange={(e) => setFormData(prev => ({ ...prev, inviteCode: e.target.value }))}
                placeholder="أدخل رمز الدعوة"
                required
              />
            </div>
          )}

          <button type="submit" className="btn btn-primary" disabled={loading}>
            <span>{isLogin ? '🔐' : '✨'}</span>
            {loading ? 'جاري التحميل...' : (isLogin ? 'دخول' : 'تسجيل')}
          </button>

          <p style={{ textAlign: 'center', marginTop: '1.5rem' }}>
            <button
              type="button"
              onClick={() => setIsLogin(!isLogin)}
              style={{ color: 'var(--primary-color)', background: 'none', border: 'none', cursor: 'pointer' }}
            >
              {isLogin ? 'ليس لديك حساب؟ سجل الآن' : 'لديك حساب؟ سجل الدخول'}
            </button>
          </p>
        </form>
      </div>
    </div>
  )
}
```

### المرحلة 3: إنشاء المكونات (يوم 6-8)

#### 3.1 مكونات UI أساسية
```typescript
// components/ui/Button.tsx
import { ButtonHTMLAttributes, ReactNode } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'btn',
  {
    variants: {
      variant: {
        primary: 'btn-primary',
        secondary: 'btn-secondary',
        danger: 'btn-danger',
        success: 'btn-success',
      },
      size: {
        sm: 'text-sm px-3 py-1.5',
        md: 'text-base px-4 py-2',
        lg: 'text-lg px-6 py-3',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  }
)

interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  children: ReactNode
  loading?: boolean
}

export function Button({
  className,
  variant,
  size,
  children,
  loading,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(buttonVariants({ variant, size }), className)}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <span>⏳</span>}
      {children}
    </button>
  )
}
```

#### 3.2 مكون البطاقة
```typescript
// components/ui/Card.tsx
import { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
  hover?: boolean
}

export function Card({ className, children, hover, ...props }: CardProps) {
  return (
    <div
      className={cn('card', hover && 'hover-lift', className)}
      {...props}
    >
      {children}
    </div>
  )
}

export function CardHeader({ className, children, ...props }: { className?: string, children: ReactNode }) {
  return (
    <div className={cn('card-header', className)} {...props}>
      {children}
    </div>
  )
}

export function CardBody({ className, children, ...props }: { className?: string, children: ReactNode }) {
  return (
    <div className={cn('card-body', className)} {...props}>
      {children}
    </div>
  )
}
```

### المرحلة 4: إعداد API والتوجيه (يوم 9-10)

#### 4.1 إعداد API Routes
```typescript
// app/api/auth/login/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function POST(request: NextRequest) {
  try {
    const { username, password } = await request.json()

    // استدعاء API الخلفي الحالي
    const backendResponse = await fetch(`${process.env.BACKEND_URL}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    })

    const data = await backendResponse.json()

    if (!backendResponse.ok) {
      return NextResponse.json(
        { error: data.error || 'فشل في تسجيل الدخول' },
        { status: backendResponse.status }
      )
    }

    return NextResponse.json(data)
  } catch (error) {
    return NextResponse.json(
      { error: 'حدث خطأ في الخادم' },
      { status: 500 }
    )
  }
}
```

#### 4.2 إعداد Middleware للمصادقة
```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('auth-token')?.value

  // الصفحات التي تحتاج مصادقة
  const protectedPaths = ['/dashboard', '/admin', '/manga']
  const isProtectedPath = protectedPaths.some(path =>
    request.nextUrl.pathname.startsWith(path)
  )

  if (isProtectedPath && !token) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  // إعادة توجيه المستخدمين المصادقين من صفحة تسجيل الدخول
  if (request.nextUrl.pathname === '/login' && token) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/admin/:path*',
    '/manga/:path*',
    '/login'
  ]
}
```

### المرحلة 5: تحويل باقي الصفحات (يوم 11-14)

#### 5.1 صفحة Dashboard
```typescript
// app/dashboard/page.tsx
'use client'

import { useEffect, useState } from 'react'
import { useAuthStore } from '@/lib/store/auth'
import { Card } from '@/components/ui/Card'
import Link from 'next/link'

interface Manga {
  id: string
  title: string
  description: string
  coverUrl?: string
}

export default function DashboardPage() {
  const [mangas, setMangas] = useState<Manga[]>([])
  const [loading, setLoading] = useState(true)
  const { user, logout } = useAuthStore()

  useEffect(() => {
    loadMangas()
  }, [])

  const loadMangas = async () => {
    try {
      const response = await fetch('/api/mangas')
      const data = await response.json()
      setMangas(data.mangas || [])
    } catch (error) {
      console.error('فشل في تحميل المانجا:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="container">
        <div className="loading">
          <div className="loading-spinner"></div>
          <p>جاري تحميل المانجا...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="container">
      <nav className="nav-bar">
        <h1>مرحباً، {user?.username}!</h1>
        <div className="nav-links">
          <Link href="/dashboard" className="btn btn-secondary">الرئيسية</Link>
          {user?.role === 'admin' && (
            <Link href="/admin" className="btn btn-primary">لوحة الإدارة</Link>
          )}
          <button onClick={logout} className="btn btn-danger">تسجيل الخروج</button>
        </div>
      </nav>

      <div className="manga-grid">
        {mangas.map((manga) => (
          <Link key={manga.id} href={`/manga/${manga.id}`}>
            <Card hover>
              <div className="manga-cover">
                {manga.coverUrl ? (
                  <img src={manga.coverUrl} alt={manga.title} />
                ) : (
                  <div className="placeholder-cover">
                    <span>📚</span>
                  </div>
                )}
              </div>
              <div className="manga-info">
                <h3 className="manga-title">{manga.title}</h3>
                <p className="manga-description">{manga.description}</p>
              </div>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}
```

### المرحلة 6: الاختبار والتحسين (يوم 15-16)

#### 6.1 كتابة الاختبارات
```typescript
// __tests__/components/Button.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from '@/components/ui/Button'

describe('Button', () => {
  it('renders children correctly', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByText('Click me')).toBeInTheDocument()
  })

  it('handles click events', () => {
    const handleClick = jest.fn()
    render(<Button onClick={handleClick}>Click me</Button>)

    fireEvent.click(screen.getByText('Click me'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('shows loading state', () => {
    render(<Button loading>Click me</Button>)
    expect(screen.getByText('⏳')).toBeInTheDocument()
  })
})
```

#### 6.2 إعداد CI/CD
```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run linting
      run: npm run lint

    - name: Run tests
      run: npm run test

    - name: Build
      run: npm run build
```

### المرحلة 7: النشر (يوم 17-18)

#### 7.1 إعداد Vercel للنشر
```json
// vercel.json
{
  "buildCommand": "npm run build",
  "outputDirectory": ".next",
  "framework": "nextjs",
  "rewrites": [
    {
      "source": "/api/(.*)",
      "destination": "http://your-backend-url/api/$1"
    }
  ]
}
```

#### 7.2 إعداد متغيرات البيئة
```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_APP_URL=http://localhost:3000

# .env.production
NEXT_PUBLIC_API_URL=https://your-backend-domain.com/api
NEXT_PUBLIC_APP_URL=https://your-frontend-domain.com
```

## 🎯 النتيجة النهائية

بعد إكمال هذه الخطة، ستحصل على:

- ✅ **تطبيق React حديث** مع TypeScript
- ✅ **أداء محسن** مع Code Splitting
- ✅ **تجربة مطور ممتازة** مع Hot Reload
- ✅ **اختبارات آلية** للجودة
- ✅ **نشر تلقائي** مع CI/CD
- ✅ **مكونات قابلة لإعادة الاستخدام**
- ✅ **إدارة حالة محسنة**
- ✅ **تصميم متجاوب** محفوظ

## 💡 نصائح للتنفيذ

1. **ابدأ صغيراً**: ركز على صفحة واحدة أولاً
2. **احتفظ بالتصميم**: انقل CSS الحالي كما هو
3. **اختبر تدريجياً**: اختبر كل مكون قبل الانتقال للآخر
4. **استخدم Git**: احفظ التغييرات في branches منفصلة
5. **اطلب المساعدة**: إذا واجهت صعوبة في أي خطوة

---

**هل تريد أن نبدأ بتنفيذ هذه الخطة؟ 🚀**
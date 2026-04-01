'use client'

import { FormEvent, useState } from 'react'
import { useRouter } from 'next/navigation'
import { HiOutlineLockClosed } from 'react-icons/hi2'
import { IoEyeOffOutline, IoEyeOutline, IoEnterOutline } from 'react-icons/io5'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { api } from '@/lib/api'
import { setToken } from '@/lib/auth'
import { LogoMark } from '@/components/logo-mark'

export default function LoginPage() {
  const router = useRouter()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showPassword, setShowPassword] = useState(false)

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      const data = await api.login({ username, password })
      setToken(data.token)
      router.push('/dashboard')
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="min-h-screen flex items-center justify-center p-4"
      style={{
        background: 'var(--background)',
      }}
    >
      <div className="w-full max-w-[440px]">
        <div className="mb-8 flex flex-col items-start gap-4">
          <div className="flex items-center gap-3">
            <LogoMark className="h-12 w-12" />
            <div>
              <h1 className="text-xl font-medium" style={{ color: 'var(--foreground)' }}>LocalAIHub</h1>
            </div>
          </div>
        </div>

        <Card className="backdrop-blur">
          <CardHeader className="p-6 pb-0">
            <CardTitle className="text-sm font-medium">管理员登录</CardTitle>
          </CardHeader>

          <CardContent className="p-6">
          <form className="space-y-4" onSubmit={onSubmit}>
            <div>
              <label className="mb-1.5 block text-xs font-medium" style={{ color: 'var(--foreground)' }}>账号</label>
              <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="账号" />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium" style={{ color: 'var(--foreground)' }}>密码</label>
              <div className="relative">
                <Input className="pr-12" type={showPassword ? 'text' : 'password'} value={password} onChange={(e) => setPassword(e.target.value)} placeholder="密码" />
                <button type="button" className="absolute right-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--muted-foreground)' }} onClick={() => setShowPassword((value) => !value)}>
                  {showPassword ? <IoEyeOffOutline className="h-4 w-4" /> : <IoEyeOutline className="h-4 w-4" />}
                </button>
              </div>
            </div>
            {error ? <div className="rounded-xl border px-4 py-3 text-sm" style={{ borderColor: 'var(--destructive)', background: 'rgba(239,68,68,0.1)', color: 'var(--destructive)' }}>{error}</div> : null}
            <Button className="w-full" disabled={loading} type="submit">
              <IoEnterOutline className="mr-2 h-4 w-4" />
              {loading ? '登录中...' : '登录'}
            </Button>
          </form>
          <div className="mt-6 flex items-center gap-2 border-t pt-4 text-xs" style={{ borderColor: 'var(--border)', color: 'var(--muted-foreground)' }}>
            <HiOutlineLockClosed className="h-4 w-4" />
            管理员访问
          </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

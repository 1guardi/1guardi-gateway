import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { useLogin } from '../api/auth'

interface LoginProps {
  onLogin: () => void
}

export default function Login({ onLogin }: LoginProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const login = useLogin(onLogin)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    login.mutate({ email, password })
  }

  return (
    <div className="flex-1 min-h-screen flex items-center justify-center bg-background">
      <div
        className="fixed inset-0 pointer-events-none"
        style={{
          backgroundImage:
            'linear-gradient(var(--grid-color) 1px, transparent 1px), linear-gradient(90deg, var(--grid-color) 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />
      <div className="relative w-full max-w-md px-4">
        <div className="rounded-xl border bg-card p-8 shadow-lg space-y-6">
          {/* Logo */}
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-primary/8 border border-primary/20">
              <svg viewBox="0 0 32 32" fill="none" className="w-5 h-5">
                <circle cx="16" cy="16" r="14" stroke="currentColor" strokeWidth="1.5" strokeOpacity="0.3" className="text-primary" />
                <circle cx="16" cy="16" r="8" stroke="currentColor" strokeWidth="1.5" strokeOpacity="0.5" className="text-primary" />
                <circle cx="16" cy="16" r="3" fill="currentColor" className="text-primary" />
                <path d="M16 4V16L24 24" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-primary" />
              </svg>
            </div>
            <div>
              <p className="font-mono font-black text-sm tracking-widest text-foreground">
                AI <span className="text-primary">GATEWAY</span>
              </p>
              <p className="font-mono text-[9px] tracking-widest text-muted-foreground">ADMIN CONSOLE</p>
            </div>
          </div>

          <div>
            <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-1">ACCESS CONTROL</p>
            <h1 className="text-lg font-semibold text-foreground">Sign in</h1>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <label className="font-mono text-[9px] tracking-widest text-muted-foreground">EMAIL</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full h-9 rounded-md border border-input bg-background px-3 font-mono text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                placeholder="admin@example.com"
                autoComplete="email"
                required
              />
            </div>

            <div className="space-y-1.5">
              <label className="font-mono text-[9px] tracking-widest text-muted-foreground">PASSWORD</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full h-9 rounded-md border border-input bg-background px-3 font-mono text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                placeholder="••••••••"
                autoComplete="current-password"
                required
              />
            </div>

            {login.isError && (
              <div className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2">
                <p className="font-mono text-[10px] text-destructive">Invalid credentials</p>
              </div>
            )}

            <Button
              type="submit"
              className="w-full font-mono text-xs tracking-widest"
              disabled={login.isPending}
            >
              {login.isPending ? 'AUTHENTICATING...' : 'SIGN IN'}
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}

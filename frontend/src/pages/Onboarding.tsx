import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { useCreateOrganization } from '../api/tenants'

interface OnboardingProps {
  /** Called once the organization has been created. */
  onComplete?: () => void
}

/**
 * Onboarding collects organization details from a user who belongs to no
 * tenant yet and creates their first organization (they become its admin).
 */
export default function Onboarding({ onComplete }: OnboardingProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  const createOrg = useCreateOrganization()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      return
    }
    createOrg.mutate(
      { name: name.trim(), description: description.trim() || undefined },
      { onSuccess: () => onComplete?.() },
    )
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
          <div>
            <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-1">GET STARTED</p>
            <h1 className="text-lg font-semibold text-foreground">Create your organization</h1>
            <p className="text-xs text-muted-foreground mt-1">
              You're not part of an organization yet. Set one up to continue — you'll be its admin.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <label className="font-mono text-[9px] tracking-widest text-muted-foreground">ORGANIZATION NAME</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full h-9 rounded-md border border-input bg-background px-3 font-mono text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                placeholder="Acme Inc."
                autoFocus
                required
              />
            </div>

            <div className="space-y-1.5">
              <label className="font-mono text-[9px] tracking-widest text-muted-foreground">DESCRIPTION (OPTIONAL)</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                className="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary resize-none"
                placeholder="What this organization is for"
              />
            </div>

            {createOrg.isError && (
              <div className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2">
                <p className="font-mono text-[10px] text-destructive">Failed to create organization. Try again.</p>
              </div>
            )}

            <Button
              type="submit"
              className="w-full font-mono text-xs tracking-widest"
              disabled={createOrg.isPending || !name.trim()}
            >
              {createOrg.isPending ? 'CREATING...' : 'CREATE ORGANIZATION'}
            </Button>
          </form>
        </div>
      </div>
    </div>
  )
}

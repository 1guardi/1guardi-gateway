import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Plus, Trash2, Building2 } from 'lucide-react'
import { useTenants, useCreateTenant, useDeleteTenant } from '../api/tenants'
import { jwtDecode } from 'jwt-decode'

interface JWTPayload {
  is_super_admin: boolean
  [key: string]: any
}

interface TenantsProps {
  activeTenantId: string | null
  onTenantSelect: (id: string) => void
}

export default function Tenants({ activeTenantId, onTenantSelect }: TenantsProps) {
  const { data: tenants = [], isLoading } = useTenants()
  const createTenant = useCreateTenant()
  const deleteTenant = useDeleteTenant()

  const token = localStorage.getItem('admin_token')
  const isSuperAdmin = token ? jwtDecode<JWTPayload>(token).is_super_admin : false

  const [showCreate, setShowCreate] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<number | null>(null)

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    await createTenant.mutateAsync({ name, description })
    setName('')
    setDescription('')
    setShowCreate(false)
  }

  const handleDelete = async (id: number) => {
    await deleteTenant.mutateAsync(id)
    setDeleteConfirm(null)
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-1">MULTI-TENANT</p>
          <h1 className="text-xl font-bold text-foreground">Tenants</h1>
        </div>
        {isSuperAdmin && (
          <Button
            onClick={() => setShowCreate(true)}
            className="font-mono text-xs tracking-widest gap-2"
          >
            <Plus className="w-3.5 h-3.5" />
            NEW TENANT
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 max-w-sm">
        <div className="rounded-lg border bg-card p-4">
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-1">TOTAL TENANTS</p>
          <p className="font-mono text-2xl font-bold text-foreground">{tenants.length}</p>
        </div>
        <div className="rounded-lg border bg-card p-4">
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-1">ACTIVE SCOPE</p>
          <p className="font-mono text-2xl font-bold text-primary">
            {activeTenantId ? tenants.find(t => String(t.ID) === activeTenantId)?.Name?.slice(0, 6) || '—' : '—'}
          </p>
        </div>
      </div>

      {/* Tenant list */}
      <div className="rounded-xl border bg-card overflow-hidden">
        <div className="px-4 py-3 border-b bg-muted/30">
          <div className="grid grid-cols-[1fr_2fr_auto_auto] gap-4">
            <span className="font-mono text-[9px] tracking-widest text-muted-foreground">NAME</span>
            <span className="font-mono text-[9px] tracking-widest text-muted-foreground">DESCRIPTION</span>
            <span className="font-mono text-[9px] tracking-widest text-muted-foreground">CREATED</span>
            <span className="font-mono text-[9px] tracking-widest text-muted-foreground">ACTIONS</span>
          </div>
        </div>

        {isLoading ? (
          <div className="px-4 py-8 text-center">
            <p className="font-mono text-xs text-muted-foreground">LOADING...</p>
          </div>
        ) : tenants.length === 0 ? (
          <div className="px-4 py-8 text-center">
            <Building2 className="w-8 h-8 text-muted-foreground/30 mx-auto mb-2" />
            <p className="font-mono text-xs text-muted-foreground">NO TENANTS YET</p>
          </div>
        ) : (
          tenants.map((tenant) => {
            const isActive = String(tenant.ID) === activeTenantId
            return (
              <div
                key={tenant.ID}
                className={`px-4 py-3.5 border-b last:border-b-0 grid grid-cols-[1fr_2fr_auto_auto] gap-4 items-center cursor-pointer hover:bg-muted/20 transition-colors ${isActive ? 'bg-primary/4' : ''}`}
                onClick={() => onTenantSelect(String(tenant.ID))}
              >
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs font-semibold text-foreground">{tenant.Name}</span>
                  {isActive && (
                    <Badge variant="outline" className="font-mono text-[8px] tracking-widest text-primary border-primary/30 py-0 px-1.5">
                      ACTIVE
                    </Badge>
                  )}
                </div>
                <span className="font-mono text-xs text-muted-foreground truncate">
                  {tenant.Description || '—'}
                </span>
                <span className="font-mono text-[10px] text-muted-foreground whitespace-nowrap">
                  {tenant.CreatedAt ? new Date(tenant.CreatedAt).toLocaleDateString() : '—'}
                </span>
                <div className="flex justify-end" onClick={(e) => e.stopPropagation()}>
                  {isSuperAdmin && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-muted-foreground hover:text-destructive"
                      onClick={() => setDeleteConfirm(tenant.ID)}
                      disabled={deleteTenant.isPending}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  )}
                </div>
              </div>
            )
          })
        )}
      </div>

      {/* Create dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="font-mono text-sm tracking-widest">NEW TENANT</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-1.5">
              <Label className="font-mono text-[9px] tracking-widest text-muted-foreground">NAME</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="my-project"
                className="font-mono text-xs"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label className="font-mono text-[9px] tracking-widest text-muted-foreground">DESCRIPTION</Label>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description"
                className="font-mono text-xs resize-none"
                rows={2}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="ghost" onClick={() => setShowCreate(false)} className="font-mono text-xs">
                CANCEL
              </Button>
              <Button type="submit" disabled={createTenant.isPending} className="font-mono text-xs tracking-widest">
                {createTenant.isPending ? 'CREATING...' : 'CREATE'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete confirm dialog */}
      <Dialog open={deleteConfirm !== null} onOpenChange={() => setDeleteConfirm(null)}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="font-mono text-sm tracking-widest text-destructive">DELETE TENANT</DialogTitle>
          </DialogHeader>
          <p className="font-mono text-xs text-muted-foreground">
            This will soft-delete the tenant and all associated agents, keys, and upstreams. Cannot be undone.
          </p>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDeleteConfirm(null)} className="font-mono text-xs">
              CANCEL
            </Button>
            <Button
              variant="destructive"
              onClick={() => deleteConfirm !== null && handleDelete(deleteConfirm)}
              disabled={deleteTenant.isPending}
              className="font-mono text-xs tracking-widest"
            >
              {deleteTenant.isPending ? 'DELETING...' : 'DELETE'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

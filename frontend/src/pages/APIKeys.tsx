import { useState } from 'react'
import { Plus, Key, Copy, Shield, ShieldAlert, Trash2, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { AgentSummary } from '../api/agents.ts'
import { useAPIKeys, useCreateAPIKey, useDeleteAPIKey } from '../api/keys.ts'
import type { APIKeyResponse } from '../api/keys.ts'

interface APIKeyVM {
  id: string
  name: string
  prefix: string
  tenantId: string
  agentId?: string
  lastUsed: string
  isActive: boolean
  createdAt: string
}

function toVM(k: APIKeyResponse): APIKeyVM {
  return {
    id: String(k.ID),
    name: k.Name,
    prefix: k.Prefix,
    tenantId: String(k.TenantID),
    agentId: k.AgentID != null ? String(k.AgentID) : undefined,
    lastUsed: k.LastUsedAt ? new Date(k.LastUsedAt).toLocaleDateString() : 'Never',
    isActive: k.IsActive,
    createdAt: new Date(k.CreatedAt).toLocaleDateString(),
  }
}

interface APIKeysProps {
  selectedAgent: string
  tenantId: string | null
  agents: AgentSummary[]
}

export default function APIKeys({ selectedAgent, tenantId, agents }: APIKeysProps) {
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyScope, setNewKeyScope] = useState<'project' | 'agent'>('project')
  const [newKeyAgentId, setNewKeyAgentId] = useState<string>('')
  const [createdKey, setCreatedKey] = useState<string | null>(null)
  const [isCreateOpen, setIsCreateOpen] = useState(false)

  const { data: keysData = [] } = useAPIKeys(tenantId)
  const keys = keysData.map(toVM)

  const { mutate: createAPIKey } = useCreateAPIKey(tenantId)
  const { mutate: deleteAPIKey } = useDeleteAPIKey(tenantId)

  const agentMap: Record<string, string> = Object.fromEntries(agents.map((a) => [String(a.ID), a.Name]))

  const filteredKeys = keys.filter(k =>
    selectedAgent === 'all' || !k.agentId || k.agentId === selectedAgent
  )

  const handleCreateKey = () => {
    if (!tenantId) return
    const body: { name: string; agent_id?: number } = { name: newKeyName || 'Untitled Key' }
    if (newKeyScope === 'agent' && newKeyAgentId) body.agent_id = Number(newKeyAgentId)
    
    createAPIKey(body, {
      onSuccess: (data) => {
        setCreatedKey(data.key)
        setNewKeyName('')
        setNewKeyScope('project')
        setNewKeyAgentId('')
      }
    })
  }

  const handleRevoke = (id: string) => {
    if (!tenantId) return
    deleteAPIKey(id)
  }

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">API Keys</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Access credentials · {keys.length} entries</p>
        </div>

        <Dialog open={isCreateOpen} onOpenChange={(open) => {
          setIsCreateOpen(open)
          if (!open) setCreatedKey(null)
        }}>
          <DialogTrigger asChild>
            <Button size="sm" className="font-mono text-[10px] tracking-widest uppercase gap-2 h-9">
              <Plus className="w-3.5 h-3.5" /> Create New Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            {!createdKey ? (
              <>
                <DialogHeader>
                  <DialogTitle className="font-black tracking-tight">Create API Key</DialogTitle>
                  <DialogDescription className="font-mono text-xs">
                    Generate a new key to authenticate requests to the gateway.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Key Name</Label>
                    <Input
                      id="name"
                      placeholder="e.g. Production Frontend"
                      value={newKeyName}
                      onChange={(e) => setNewKeyName(e.target.value)}
                      className="font-mono text-xs"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Scope</Label>
                    <Select value={newKeyScope} onValueChange={(v) => { setNewKeyScope(v as 'project' | 'agent'); setNewKeyAgentId('') }}>
                      <SelectTrigger className="font-mono text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="project" className="font-mono text-xs">Project Level (All Agents)</SelectItem>
                        <SelectItem value="agent" className="font-mono text-xs" disabled={agents.length === 0}>
                          Agent Specific{agents.length === 0 ? ' (no agents)' : ''}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  {newKeyScope === 'agent' && agents.length > 0 && (
                    <div className="grid gap-2">
                      <Label className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Agent</Label>
                      <Select value={newKeyAgentId} onValueChange={setNewKeyAgentId}>
                        <SelectTrigger className="font-mono text-xs">
                          <SelectValue placeholder="Select agent" />
                        </SelectTrigger>
                        <SelectContent>
                          {agents.map((a) => (
                            <SelectItem key={a.ID} value={String(a.ID)} className="font-mono text-xs">
                              {a.Name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)} className="font-mono text-xs">Cancel</Button>
                  <Button
                    onClick={handleCreateKey}
                    className="font-mono text-xs"
                    disabled={!tenantId || (newKeyScope === 'agent' && !newKeyAgentId)}
                  >
                    Generate Key
                  </Button>
                </DialogFooter>
              </>
            ) : (
              <>
                <DialogHeader>
                  <DialogTitle className="flex items-center gap-2 font-black tracking-tight">
                    <Shield className="w-5 h-5 text-primary" /> Key Generated
                  </DialogTitle>
                  <DialogDescription className="font-mono text-xs">
                    Copy this key now. For security, it won't be shown again.
                  </DialogDescription>
                </DialogHeader>
                <div className="py-6">
                  <div className="flex items-center gap-2 p-3 bg-muted rounded-lg border font-mono text-sm break-all">
                    <span className="flex-1">{createdKey}</span>
                    <Button size="icon" variant="ghost" onClick={() => navigator.clipboard.writeText(createdKey)}>
                      <Copy className="w-4 h-4" />
                    </Button>
                  </div>
                  <div className="mt-4 flex items-start gap-2 p-3 bg-primary/5 rounded-md border border-primary/20">
                    <ShieldAlert className="w-4 h-4 text-primary mt-0.5" />
                    <p className="text-[11px] text-primary/80 leading-relaxed font-mono uppercase">
                      ANYONE WITH THIS KEY CAN MAKE REQUESTS ON BEHALF OF YOUR TENANT.
                      TREAT IT AS A SENSITIVE CREDENTIAL.
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <Button className="w-full font-mono text-xs" onClick={() => setIsCreateOpen(false)}>I have saved the key</Button>
                </DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
      </div>

      {/* Status strip */}
      <div className="grid grid-cols-3 gap-2">
        <Card className="bg-primary/5 border-primary/10">
          <CardContent className="p-3">
            <p className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase mb-2">Active Keys</p>
            <p className="font-mono text-2xl font-black text-foreground">{keys.filter(k => k.isActive).length}</p>
            <div className="flex items-center gap-1.5 mt-1">
              <span className="text-[10px] font-mono text-primary uppercase">Provisioned</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3">
            <p className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase mb-2">Total Requests (24h)</p>
            <p className="font-mono text-2xl font-black text-foreground">—</p>
            <div className="flex items-center gap-1.5 mt-1">
              <span className="text-[10px] font-mono text-muted-foreground uppercase">Not tracked</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3">
            <p className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase mb-2">Key Usage</p>
            <p className="font-mono text-2xl font-black text-foreground">—</p>
            <div className="flex items-center gap-1.5 mt-1">
              <span className="text-[10px] font-mono text-muted-foreground uppercase">Not tracked</span>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">ALL KEYS</CardTitle>
          <span className="font-mono text-xs text-muted-foreground/40">{filteredKeys.length} results</span>
        </CardHeader>
        <CardContent className="p-0">
          <ScrollArea>
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  {['NAME', 'SCOPE', 'LAST USED', 'CREATED', 'STATUS', 'ACTIONS'].map((h) => (
                    <TableHead key={h} className={`font-mono text-[10px] tracking-widest text-muted-foreground/50 ${h === 'ACTIONS' ? 'text-right' : ''}`}>{h}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredKeys.map((key) => (
                  <TableRow key={key.id} className={`border-border transition-colors hover:bg-muted/30 ${!key.isActive ? 'opacity-50' : ''}`}>
                    <TableCell className="font-mono text-xs">
                      <div className="flex items-center gap-2">
                        <div className={`p-1 rounded ${key.isActive ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}>
                          <Key className="w-3 h-3" />
                        </div>
                        <span className="font-bold text-foreground">{key.name}</span>
                      </div>
                    </TableCell>
                    <TableCell>
                      {key.agentId ? (
                        <Badge variant="outline" className="font-mono text-[9px] h-3.5 px-1 text-primary border-primary/20 bg-primary/5 uppercase">
                          {agentMap[key.agentId] ?? `AGENT ${key.agentId}`}
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="font-mono text-[9px] h-3.5 px-1 text-muted-foreground border-muted-foreground/20 uppercase">
                          Global
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{key.lastUsed}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{key.createdAt}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className={`font-mono text-[9px] h-3.5 px-1 uppercase ${key.isActive ? 'text-primary border-primary/30 bg-primary/8' : 'text-muted-foreground border-border bg-muted/40'}`}>
                        {key.isActive ? 'Active' : 'Revoked'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-foreground">
                          <ExternalLink className="w-3 h-3" />
                        </Button>
                        {key.isActive && (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 text-muted-foreground hover:text-error hover:bg-error/10"
                            onClick={() => handleRevoke(key.id)}
                          >
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>

      <div className="p-4 rounded-lg border border-border border-dashed flex items-center justify-between bg-muted/10">
        <div className="flex items-center gap-3">
          <Shield className="w-4 h-4 text-muted-foreground" />
          <div>
            <p className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">Security Protocol</p>
            <p className="font-mono text-xs text-muted-foreground/60">Emergency key rotation available in tenant configuration.</p>
          </div>
        </div>
        <Button variant="outline" size="sm" className="font-mono text-[10px] tracking-widest uppercase h-8 px-3">
          Settings
        </Button>
      </div>
    </div>
  )
}

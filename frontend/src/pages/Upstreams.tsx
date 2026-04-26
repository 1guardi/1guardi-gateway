import { useState } from 'react'
import { Plus, Key, Trash2, Globe, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useUpstreams, useCreateUpstream, useDeleteUpstream } from '../api/upstreams.ts'
import type { UpstreamResponse } from '../api/upstreams.ts'

interface UpstreamVM {
  id: string
  keyId: string
  model: string
  baseUrl: string
  createdAt: string
}

function toVM(u: UpstreamResponse): UpstreamVM {
  return {
    id: String(u.ID),
    keyId: u.key_id,
    model: u.model,
    baseUrl: u.base_url,
    createdAt: new Date(u.CreatedAt).toLocaleDateString(),
  }
}

interface UpstreamsProps {
  tenantId: string | null
}

export default function Upstreams({ tenantId }: UpstreamsProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [newKeyID, setNewKeyID] = useState('')
  const [newModel, setNewModel] = useState('')
  const [newBaseURL, setNewBaseURL] = useState('https://api.openai.com')
  const [newAPIKey, setNewAPIKey] = useState('')

  const { data = [] } = useUpstreams(tenantId)
  const upstreams = data.map(toVM)

  const { mutate: createUpstream } = useCreateUpstream(tenantId)
  const { mutate: deleteUpstream } = useDeleteUpstream(tenantId)

  const handleCreate = () => {
    if (!tenantId) return
    createUpstream(
      { key_id: newKeyID, model: newModel, base_url: newBaseURL, api_key: newAPIKey },
      {
        onSuccess: () => {
          setIsCreateOpen(false)
          setNewKeyID('')
          setNewModel('')
          setNewBaseURL('https://api.openai.com')
          setNewAPIKey('')
        },
      }
    )
  }

  const handleDelete = (keyId: string) => {
    if (!tenantId) return
    deleteUpstream(keyId)
  }

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Provider Keys</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Upstream LLM credentials · {upstreams.length} configured</p>
        </div>

        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button size="sm" className="font-mono text-[10px] tracking-widest uppercase gap-2 h-9">
              <Plus className="w-3.5 h-3.5" /> Add Provider Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle className="font-black tracking-tight">Add Provider Key</DialogTitle>
              <DialogDescription className="font-mono text-xs">
                Configure a new upstream LLM endpoint. This key will be used by the router.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="keyId" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Key Identifier</Label>
                <Input id="keyId" placeholder="e.g. openai-primary" value={newKeyID} onChange={(e) => setNewKeyID(e.target.value)} className="font-mono text-xs" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="model" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Model Name</Label>
                <Input id="model" placeholder="e.g. gpt-4o" value={newModel} onChange={(e) => setNewModel(e.target.value)} className="font-mono text-xs" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="baseUrl" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Base URL</Label>
                <Input id="baseUrl" value={newBaseURL} onChange={(e) => setNewBaseURL(e.target.value)} className="font-mono text-xs" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="apiKey" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">API Key</Label>
                <Input id="apiKey" type="password" value={newAPIKey} onChange={(e) => setNewAPIKey(e.target.value)} className="font-mono text-xs" />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateOpen(false)} className="font-mono text-xs">Cancel</Button>
              <Button onClick={handleCreate} className="font-mono text-xs" disabled={!newKeyID || !newModel || !newAPIKey}>Add Key</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent border-b border-border/50">
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10">Identifier</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10">Model</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10">Base URL</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {upstreams.map((u) => (
                <TableRow key={u.id} className="group border-b border-border/40">
                  <TableCell className="py-4">
                    <div className="flex items-center gap-2">
                      <Key className="w-3.5 h-3.5 text-muted-foreground" />
                      <span className="font-bold text-sm text-foreground">{u.keyId}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="font-mono text-[10px] bg-muted/50">{u.model}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{u.baseUrl}</TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity" onClick={() => handleDelete(u.keyId)}>
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
              {upstreams.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-32 text-center">
                    <div className="flex flex-col items-center justify-center text-muted-foreground">
                      <Globe className="w-8 h-8 mb-2 opacity-20" />
                      <p className="font-mono text-xs">No provider keys configured</p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="bg-primary/5 border-primary/10">
          <CardHeader className="pb-2">
            <CardTitle className="font-mono text-[10px] tracking-widest text-primary uppercase flex items-center gap-2">
              <Shield className="w-3 h-3" /> Security Note
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Provider keys are stored encrypted at rest. They are only used by the gateway to authenticate requests to upstream LLM providers on your behalf.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

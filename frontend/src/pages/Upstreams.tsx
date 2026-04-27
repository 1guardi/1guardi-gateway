import { useState, useEffect } from 'react'
import { Plus, Key, Trash2, Globe, Shield, Check, ChevronsUpDown, Search, Edit2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useUpstreams, useCreateUpstream, useDeleteUpstream, useProviderModels, useUpdateUpstream } from '../api/upstreams.ts'
import type { UpstreamResponse } from '../api/upstreams.ts'
import { cn } from '@/lib/utils'

interface UpstreamVM {
  id: string
  keyId: string
  provider: string
  models: string[]
  baseUrl: string
  createdAt: string
}

function toVM(u: UpstreamResponse): UpstreamVM {
  return {
    id: String(u.ID),
    keyId: u.key_id,
    provider: u.provider,
    models: u.models ? u.models.split(',') : [],
    baseUrl: u.base_url,
    createdAt: new Date(u.CreatedAt).toLocaleDateString(),
  }
}

interface UpstreamsProps {
  tenantId: string | null
}

const PROVIDERS = [
  { id: 'openai', name: 'OpenAI', defaultUrl: 'https://api.openai.com' },
  { id: 'anthropic', name: 'Anthropic', defaultUrl: 'https://api.anthropic.com' },
  { id: 'gemini', name: 'Gemini', defaultUrl: 'https://generativelanguage.googleapis.com' },
  { id: 'openai-compatible', name: 'Custom (OpenAI Compatible)', defaultUrl: '' },
]

export default function Upstreams({ tenantId }: UpstreamsProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [editingKeyId, setEditingKeyId] = useState<string | null>(null)
  const [newKeyID, setNewKeyID] = useState('')
  const [newProvider, setNewProvider] = useState('openai')
  const [selectedModels, setSelectedModels] = useState<string[]>([])
  const [newBaseURL, setNewBaseURL] = useState('https://api.openai.com')
  const [newAPIKey, setNewAPIKey] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [modelSearch, setModelSearch] = useState('')

  const { data = [] } = useUpstreams(tenantId)
  const upstreams = data.map(toVM)

  const filteredUpstreams = upstreams.filter(u => 
    u.keyId.toLowerCase().includes(searchTerm.toLowerCase()) || 
    u.models.some(m => m.toLowerCase().includes(searchTerm.toLowerCase()))
  )

  const { data: availableModels = [], isLoading: isLoadingModels } = useProviderModels(
    newProvider, 
    newAPIKey,
    tenantId,
    isEditing ? editingKeyId : null
  )
  const filteredAvailableModels = availableModels.filter(m => 
    m.toLowerCase().includes(modelSearch.toLowerCase())
  )

  const { mutate: createUpstream } = useCreateUpstream(tenantId)
  const { mutate: deleteUpstream } = useDeleteUpstream(tenantId)
  const { mutate: updateUpstream } = useUpdateUpstream(tenantId)

  const handleProviderChange = (val: string) => {
    setNewProvider(val)
    setSelectedModels([])
    setModelSearch('')
    const p = PROVIDERS.find((p) => p.id === val)
    if (p && p.defaultUrl) {
      setNewBaseURL(p.defaultUrl)
    } else {
      setNewBaseURL('')
    }
  }

  const resetForm = () => {
    setIsCreateOpen(false)
    setIsEditing(false)
    setEditingKeyId(null)
    setNewKeyID('')
    setNewProvider('openai')
    setSelectedModels([])
    setNewBaseURL('https://api.openai.com')
    setNewAPIKey('')
    setModelSearch('')
  }

  const handleSubmit = () => {
    if (!tenantId) return

    const body = {
      key_id: newKeyID,
      provider: newProvider,
      models: selectedModels,
      base_url: newBaseURL,
      api_key: newAPIKey
    }

    if (isEditing && editingKeyId) {
      updateUpstream(
        { keyId: editingKeyId, body },
        { onSuccess: resetForm }
      )
    } else {
      createUpstream(body, { onSuccess: resetForm })
    }
  }

  const handleEdit = (u: UpstreamVM) => {
    setIsEditing(true)
    setEditingKeyId(u.keyId)
    setNewKeyID(u.keyId)
    setNewProvider(u.provider)
    setSelectedModels(u.models)
    setNewBaseURL(u.baseUrl)
    setNewAPIKey('') // Don't show old key
    setIsCreateOpen(true)
  }

  const toggleModel = (model: string) => {
    setSelectedModels(prev => 
      prev.includes(model) ? prev.filter(m => m !== model) : [...prev, model]
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

        <div className="flex items-center gap-3">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
            <Input 
              placeholder="Search keys or models..." 
              value={searchTerm} 
              onChange={(e) => setSearchTerm(e.target.value)} 
              className="pl-8 h-9 w-64 font-mono text-[10px]"
            />
          </div>
        </div>

          <Dialog open={isCreateOpen} onOpenChange={(open) => {
            setIsCreateOpen(open)
            if (!open) resetForm()
          }}>
            <DialogTrigger asChild>
              <Button size="sm" className="font-mono text-[10px] tracking-widest uppercase gap-2 h-9" onClick={() => {
                setIsEditing(false)
                setIsCreateOpen(true)
              }}>
                <Plus className="w-3.5 h-3.5" /> Add Provider Key
              </Button>
            </DialogTrigger>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle className="font-black tracking-tight">{isEditing ? 'Edit' : 'Add'} Provider Key</DialogTitle>
              <DialogDescription className="font-mono text-xs">
                {isEditing ? 'Update existing' : 'Configure a new'} upstream LLM endpoint. This key will be used by the router.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="keyId" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Key Identifier</Label>
                <Input 
                  id="keyId" 
                  placeholder="e.g. openai-primary" 
                  value={newKeyID} 
                  onChange={(e) => setNewKeyID(e.target.value)} 
                  className="font-mono text-xs" 
                  disabled={isEditing}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="provider" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Provider</Label>
                <Select value={newProvider} onValueChange={handleProviderChange}>
                  <SelectTrigger className="font-mono text-xs">
                    <SelectValue placeholder="Select provider" />
                  </SelectTrigger>
                  <SelectContent>
                    {PROVIDERS.map((p) => (
                      <SelectItem key={p.id} value={p.id} className="font-mono text-xs">{p.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="apiKey" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">API Key</Label>
                <Input 
                  id="apiKey" 
                  type="password" 
                  placeholder={isEditing ? "••••••••••••••••" : "sk-..."} 
                  value={newAPIKey} 
                  onChange={(e) => setNewAPIKey(e.target.value)} 
                  className="font-mono text-xs" 
                />
                {isEditing && !newAPIKey && (
                  <p className="text-[9px] font-mono text-muted-foreground italic">Leave blank to keep existing key</p>
                )}
              </div>
              
              <div className="grid gap-2">
                <Label className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Models</Label>
                <div className="relative">
                  <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
                  <Input 
                    placeholder="Search models..." 
                    value={modelSearch} 
                    onChange={(e) => setModelSearch(e.target.value)} 
                    className="pl-7 h-7 font-mono text-[10px] mb-2"
                  />
                </div>
                <div className="border rounded-md p-2 max-h-40 overflow-y-auto space-y-1">
                  {isLoadingModels ? (
                    <p className="text-[10px] font-mono text-muted-foreground animate-pulse p-2">Fetching models...</p>
                  ) : filteredAvailableModels.length > 0 ? (
                    filteredAvailableModels.map(model => (
                      <div 
                        key={model} 
                        className={cn(
                          "flex items-center justify-between px-2 py-1.5 rounded-sm cursor-pointer hover:bg-accent group",
                          selectedModels.includes(model) && "bg-accent"
                        )}
                        onClick={() => toggleModel(model)}
                      >
                        <span className="font-mono text-[10px] truncate">{model}</span>
                        {selectedModels.includes(model) && <Check className="w-3 h-3 text-primary" />}
                      </div>
                    ))
                  ) : (
                    <p className="text-[10px] font-mono text-muted-foreground p-2 text-center">Enter API key to load models</p>
                  )}
                </div>
                {selectedModels.length > 0 && (
                  <p className="text-[10px] font-mono text-muted-foreground italic">
                    {selectedModels.length} models selected
                  </p>
                )}
              </div>

              {newProvider === 'openai-compatible' && (
                <div className="grid gap-2">
                  <Label htmlFor="baseUrl" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Base URL</Label>
                  <Input id="baseUrl" value={newBaseURL} onChange={(e) => setNewBaseURL(e.target.value)} className="font-mono text-xs" />
                </div>
              )}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={resetForm} className="font-mono text-xs">Cancel</Button>
              <Button onClick={handleSubmit} className="font-mono text-xs" disabled={!newKeyID || selectedModels.length === 0 || (!isEditing && !newAPIKey)}>
                {isEditing ? 'Save Changes' : 'Add Key'}
              </Button>
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
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10">Provider</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10">Models</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10">Base URL</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest uppercase h-10 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredUpstreams.map((u) => (
                <TableRow key={u.id} className="group border-b border-border/40">
                  <TableCell className="py-4">
                    <div className="flex items-center gap-2">
                      <Key className="w-3.5 h-3.5 text-muted-foreground" />
                      <span className="font-bold text-sm text-foreground">{u.keyId}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="font-mono text-[10px] uppercase tracking-wider">{u.provider}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {u.models.map(model => (
                        <Badge key={model} variant="outline" className="font-mono text-[9px] bg-muted/30 py-0 h-4">{model}</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{u.baseUrl}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button 
                        variant="ghost" 
                        size="icon" 
                        className="h-8 w-8 text-muted-foreground hover:text-primary opacity-0 group-hover:opacity-100 transition-opacity" 
                        onClick={() => handleEdit(u)}
                      >
                        <Edit2 className="w-3.5 h-3.5" />
                      </Button>
                      <Button 
                        variant="ghost" 
                        size="icon" 
                        className="h-8 w-8 text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity" 
                        onClick={() => handleDelete(u.keyId)}
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {upstreams.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="h-32 text-center">
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

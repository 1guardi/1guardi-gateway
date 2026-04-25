import { useState } from 'react'
import { Plus, Key, Copy, Shield, ShieldAlert, Trash2, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { apiKeys as mockKeys, type APIKey } from '../data/mock'

export default function APIKeys() {
  const [keys, setKeys] = useState<APIKey[]>(mockKeys)
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyScope, setNewKeyScope] = useState('project')
  const [createdKey, setCreatedKey] = useState<string | null>(null)
  const [isCreateOpen, setIsCreateOpen] = useState(false)

  const handleCreateKey = () => {
    // Simulate API call
    const newKey: APIKey = {
      id: `key-${Math.floor(Math.random() * 1000)}`,
      name: newKeyName || 'Untitled Key',
      prefix: 'sk',
      tenantId: 'acme-corp',
      agentId: newKeyScope === 'project' ? undefined : 'AGT-001',
      lastUsed: 'Never',
      isActive: true,
      createdAt: new Date().toISOString().split('T')[0],
    }
    
    setKeys([newKey, ...keys])
    setCreatedKey(`sk_${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`)
    setNewKeyName('')
  }

  const handleRevoke = (id: string) => {
    setKeys(keys.map(k => k.id === id ? { ...k, isActive: false } : k))
  }

  return (
    <div className="p-8 max-w-6xl mx-auto space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">API Keys</h1>
          <p className="text-muted-foreground mt-1">Manage access credentials for your agents and projects.</p>
        </div>
        
        <Dialog open={isCreateOpen} onOpenChange={(open) => {
          setIsCreateOpen(open)
          if (!open) setCreatedKey(null)
        }}>
          <DialogTrigger asChild>
            <Button className="gap-2">
              <Plus className="w-4 h-4" /> Create New Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            {!createdKey ? (
              <>
                <DialogHeader>
                  <DialogTitle>Create API Key</DialogTitle>
                  <DialogDescription>
                    Generate a new key to authenticate requests to the gateway.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Key Name</Label>
                    <Input 
                      id="name" 
                      placeholder="e.g. Production Frontend" 
                      value={newKeyName}
                      onChange={(e) => setNewKeyName(e.target.value)}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="scope">Scope</Label>
                    <Select value={newKeyScope} onValueChange={setNewKeyScope}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select scope" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="project">Project Level (All Agents)</SelectItem>
                        <SelectItem value="agent">Agent Specific (Support Agent)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                  <Button onClick={handleCreateKey}>Generate Key</Button>
                </DialogFooter>
              </>
            ) : (
              <>
                <DialogHeader>
                  <DialogTitle className="flex items-center gap-2">
                    <Shield className="w-5 h-5 text-primary" /> Key Generated
                  </DialogTitle>
                  <DialogDescription>
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
                    <p className="text-[11px] text-primary/80 leading-relaxed font-mono">
                      ANYONE WITH THIS KEY CAN MAKE REQUESTS ON BEHALF OF YOUR TENANT. 
                      TREAT IT AS A SENSITIVE CREDENTIAL.
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <Button className="w-full" onClick={() => setIsCreateOpen(false)}>I have saved the key</Button>
                </DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="bg-primary/5 border-primary/10">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-mono tracking-widest text-muted-foreground uppercase">Active Keys</CardTitle>
            <CardDescription className="text-2xl font-bold text-foreground">{keys.filter(k => k.isActive).length}</CardDescription>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-mono tracking-widest text-muted-foreground uppercase">Total Requests (24h)</CardTitle>
            <CardDescription className="text-2xl font-bold text-foreground">14.2k</CardDescription>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-xs font-mono tracking-widest text-muted-foreground uppercase">Key Usage</CardTitle>
            <CardDescription className="text-2xl font-bold text-foreground">89%</CardDescription>
          </CardHeader>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Keys</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Scope</TableHead>
                <TableHead>Last Used</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((key) => (
                <TableRow key={key.id} className={!key.isActive ? 'opacity-50' : ''}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <div className={`p-1.5 rounded-md ${key.isActive ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}>
                        <Key className="w-3.5 h-3.5" />
                      </div>
                      <span className="font-medium">{key.name}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    {key.agentId ? (
                      <Badge variant="outline" className="font-mono text-[10px] gap-1">
                        AGENT: {key.agentId}
                      </Badge>
                    ) : (
                      <Badge variant="secondary" className="font-mono text-[10px] gap-1">
                        PROJECT LEVEL
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{key.lastUsed}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{key.createdAt}</TableCell>
                  <TableCell>
                    <Badge variant={key.isActive ? 'default' : 'secondary'} className={key.isActive ? 'bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/10 border-emerald-500/20' : ''}>
                      {key.isActive ? 'Active' : 'Revoked'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <ExternalLink className="w-3.5 h-3.5" />
                      </Button>
                      {key.isActive && (
                        <Button 
                          variant="ghost" 
                          size="icon" 
                          className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                          onClick={() => handleRevoke(key.id)}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="p-4 rounded-lg border border-dashed flex items-center justify-between bg-muted/30">
        <div className="flex items-center gap-3">
          <Shield className="w-5 h-5 text-muted-foreground" />
          <div>
            <p className="text-sm font-medium">Need rotate all keys?</p>
            <p className="text-xs text-muted-foreground">Emergency rotation is available in tenant settings.</p>
          </div>
        </div>
        <Button variant="outline" size="sm">Go to Settings</Button>
      </div>
    </div>
  )
}

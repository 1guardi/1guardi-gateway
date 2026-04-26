import { useState } from 'react'
import { Plus, Bot } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { AgentSummary } from '../App.tsx'

interface AgentsProps {
  tenantId: string | null
  agents: AgentSummary[]
  onAgentCreated: () => void
}

export default function Agents({ tenantId, agents, onAgentCreated }: AgentsProps) {
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [isOpen, setIsOpen] = useState(false)

  const handleCreate = () => {
    if (!tenantId || !newName.trim()) return
    fetch(`/api/v1/tenants/${tenantId}/agents`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newName.trim(), description: newDesc.trim() }),
    })
      .then((r) => { if (r.ok) return r.json() })
      .then(() => {
        setNewName('')
        setNewDesc('')
        setIsOpen(false)
        onAgentCreated()
      })
      .catch(() => {})
  }

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Agents</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Fleet registry · {agents.length} registered</p>
        </div>

        <Dialog open={isOpen} onOpenChange={setIsOpen}>
          <DialogTrigger asChild>
            <Button size="sm" className="font-mono text-[10px] tracking-widest uppercase gap-2 h-9">
              <Plus className="w-3.5 h-3.5" /> Register Agent
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle className="font-black tracking-tight">Register Agent</DialogTitle>
              <DialogDescription className="font-mono text-xs">
                Add a new agent to this tenant's fleet. Scope API keys to this agent after creation.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="agent-name" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Name *</Label>
                <Input
                  id="agent-name"
                  placeholder="e.g. customer-support"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="agent-desc" className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground">Description</Label>
                <Textarea
                  id="agent-desc"
                  placeholder="Optional description"
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  className="font-mono text-xs resize-none"
                  rows={3}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsOpen(false)} className="font-mono text-xs">Cancel</Button>
              <Button
                onClick={handleCreate}
                className="font-mono text-xs"
                disabled={!tenantId || !newName.trim()}
              >
                Register
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid grid-cols-3 gap-2">
        <Card className="bg-primary/5 border-primary/10">
          <CardContent className="p-3">
            <p className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase mb-2">Registered</p>
            <p className="font-mono text-2xl font-black text-foreground">{agents.length}</p>
            <div className="flex items-center gap-1.5 mt-1">
              <span className="text-[10px] font-mono text-primary uppercase">Agents</span>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">FLEET REGISTRY</CardTitle>
          <span className="font-mono text-xs text-muted-foreground/40">{agents.length} agents</span>
        </CardHeader>
        <CardContent className="p-0">
          <ScrollArea>
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  {['NAME', 'DESCRIPTION', 'ID', 'REGISTERED'].map((h) => (
                    <TableHead key={h} className="font-mono text-[10px] tracking-widest text-muted-foreground/50">{h}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {agents.map((agent) => (
                  <TableRow key={agent.ID} className="border-border transition-colors hover:bg-muted/30">
                    <TableCell className="font-mono text-xs">
                      <div className="flex items-center gap-2">
                        <div className="p-1 rounded bg-primary/10 text-primary">
                          <Bot className="w-3 h-3" />
                        </div>
                        <span className="font-bold text-foreground">{agent.Name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground max-w-xs truncate">
                      {agent.Description || <span className="text-muted-foreground/30">—</span>}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="font-mono text-[9px] h-3.5 px-1 text-muted-foreground border-muted-foreground/20">
                        {agent.ID}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {new Date(agent.CreatedAt).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  )
}

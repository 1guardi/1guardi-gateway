import { useState } from 'react'
import { Plus, Trash2, User, RefreshCw, Eye, EyeOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useMembers, useAddMember, useRemoveMember, useRoles, useUsers, type RoleResponse, type TenantMemberResponse } from '../api/members.ts'

interface MembersProps {
  tenantId: string | null
}

export default function Members({ tenantId }: MembersProps) {
  const [activeTab, setActiveTab] = useState<'existing' | 'new'>('existing')
  const [newMemberUserId, setNewMemberUserId] = useState<string>('')
  const [newMemberEmail, setNewMemberEmail] = useState<string>('')
  const [newMemberName, setNewMemberName] = useState<string>('')
  const [newMemberPassword, setNewMemberPassword] = useState<string>('')
  const [showPassword, setShowPassword] = useState(false)
  const [newMemberRole, setNewMemberRole] = useState<string>('')
  const [isAddOpen, setIsAddOpen] = useState(false)

  const { data: members = [] } = useMembers(tenantId)
  const { data: roles = [] } = useRoles()
  const { data: users = [] } = useUsers()

  const { mutate: addMember } = useAddMember(tenantId)
  const { mutate: removeMember } = useRemoveMember(tenantId)

  const generatePassword = () => {
    const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*'
    let pass = ''
    for (let i = 0; i < 12; i++) {
      pass += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    setNewMemberPassword(pass)
    setShowPassword(true)
  }

  const handleAddMember = () => {
    if (!tenantId || !newMemberRole) return
    const payload: any = { role_id: Number(newMemberRole) }
    if (activeTab === 'existing' && newMemberUserId) {
      payload.user_id = Number(newMemberUserId)
    } else if (activeTab === 'new' && newMemberEmail && newMemberName) {
      payload.email = newMemberEmail
      payload.name = newMemberName
      payload.password = newMemberPassword
    } else {
      return
    }

    addMember(payload, {
      onSuccess: () => {
        setNewMemberUserId('')
        setNewMemberEmail('')
        setNewMemberName('')
        setNewMemberPassword('')
        setNewMemberRole('')
        setIsAddOpen(false)
      }
    })
  }

  const handleRemove = (userId: string) => {
    if (!tenantId) return
    removeMember(userId)
  }

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Members</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Tenant users and roles · {members.length} entries</p>
        </div>

        <Dialog open={isAddOpen} onOpenChange={setIsAddOpen}>
          <DialogTrigger asChild>
            <Button size="sm" className="font-mono text-[10px] tracking-widest uppercase gap-2 h-9">
              <Plus className="w-3.5 h-3.5" />
              Add Member
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle className="font-mono text-sm tracking-widest uppercase">Add Member</DialogTitle>
              <DialogDescription className="font-mono text-xs">
                Select an existing user or invite a new one.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)} className="w-full">
                <TabsList className="grid w-full grid-cols-2 mb-4">
                  <TabsTrigger value="existing" className="font-mono text-[10px] tracking-widest uppercase">Existing User</TabsTrigger>
                  <TabsTrigger value="new" className="font-mono text-[10px] tracking-widest uppercase">New User</TabsTrigger>
                </TabsList>
                <TabsContent value="existing" className="space-y-4">
                  <div className="grid gap-2">
                    <Label htmlFor="user" className="font-mono text-[10px] tracking-widest text-muted-foreground">USER</Label>
                    <Select value={newMemberUserId} onValueChange={setNewMemberUserId}>
                      <SelectTrigger className="font-mono text-xs">
                        <SelectValue placeholder="Select a user" />
                      </SelectTrigger>
                      <SelectContent>
                        {users.map((user) => (
                          <SelectItem key={user.ID} value={String(user.ID)} className="font-mono text-xs">
                            {user.Name} ({user.Email})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </TabsContent>
                <TabsContent value="new" className="space-y-4">
                  <div className="grid gap-2">
                    <Label htmlFor="new-name" className="font-mono text-[10px] tracking-widest text-muted-foreground">NAME</Label>
                    <Input
                      id="new-name"
                      value={newMemberName}
                      onChange={(e) => setNewMemberName(e.target.value)}
                      className="font-mono text-xs"
                      placeholder="e.g. Alice Smith"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="new-email" className="font-mono text-[10px] tracking-widest text-muted-foreground">EMAIL</Label>
                    <Input
                      id="new-email"
                      type="email"
                      value={newMemberEmail}
                      onChange={(e) => setNewMemberEmail(e.target.value)}
                      className="font-mono text-xs"
                      placeholder="alice@example.com"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="new-password" content-between className="font-mono text-[10px] tracking-widest text-muted-foreground flex items-center justify-between">
                      PASSWORD
                      <Button variant="link" onClick={generatePassword} className="h-auto p-0 font-mono text-[9px] text-primary">
                        <RefreshCw className="w-3 h-3 mr-1" /> AUTO-GENERATE
                      </Button>
                    </Label>
                    <div className="relative">
                      <Input
                        id="new-password"
                        type={showPassword ? 'text' : 'password'}
                        value={newMemberPassword}
                        onChange={(e) => setNewMemberPassword(e.target.value)}
                        className="font-mono text-xs pr-10"
                        placeholder="••••••••"
                      />
                      <Button
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-full px-3 text-muted-foreground"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                      </Button>
                    </div>
                  </div>
                </TabsContent>
              </Tabs>
              <div className="grid gap-2">
                <Label htmlFor="role" className="font-mono text-[10px] tracking-widest text-muted-foreground">ROLE</Label>
                <Select value={newMemberRole} onValueChange={setNewMemberRole}>
                  <SelectTrigger className="font-mono text-xs">
                    <SelectValue placeholder="Select a role" />
                  </SelectTrigger>
                  <SelectContent>
                    {roles.map((role: RoleResponse) => (
                      <SelectItem key={role.ID} value={String(role.ID)} className="font-mono text-xs capitalize">
                        {role.Name.replace(/([A-Z])/g, ' $1').trim()}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button 
                onClick={handleAddMember} 
                disabled={!newMemberRole || (activeTab === 'existing' && !newMemberUserId) || (activeTab === 'new' && (!newMemberEmail || !newMemberName))} 
                className="font-mono text-xs uppercase tracking-widest"
              >
                Add User
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card className="border-sidebar-border bg-card/50 backdrop-blur-sm shadow-sm overflow-hidden">
        <CardHeader className="border-b border-sidebar-border bg-muted/20 px-4 py-3">
          <CardTitle className="font-mono text-xs tracking-widest text-muted-foreground flex items-center gap-2">
            <User className="w-3.5 h-3.5" />
            TENANT MEMBERS
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent border-sidebar-border">
                <TableHead className="w-[300px] font-mono text-[10px] tracking-widest text-muted-foreground h-9">USER</TableHead>
                <TableHead className="w-[200px] font-mono text-[10px] tracking-widest text-muted-foreground h-9">ROLE</TableHead>
                <TableHead className="font-mono text-[10px] tracking-widest text-muted-foreground h-9">SUPER ADMIN</TableHead>
                <TableHead className="w-[100px] text-right font-mono text-[10px] tracking-widest text-muted-foreground h-9">ACTIONS</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {members.map((member: TenantMemberResponse) => (
                <TableRow key={member.ID} className="border-sidebar-border hover:bg-muted/30">
                  <TableCell className="font-mono text-xs py-2.5">
                    <div className="font-semibold text-foreground">{member.User?.Name}</div>
                    <div className="text-[10px] text-muted-foreground">{member.User?.Email}</div>
                  </TableCell>
                  <TableCell className="py-2.5">
                    <Badge variant="secondary" className="font-mono text-[10px] tracking-widest uppercase rounded-sm bg-primary/10 text-primary hover:bg-primary/20 border-primary/20 capitalize">
                      {member.Role?.Name?.replace(/([A-Z])/g, ' $1').trim()}
                    </Badge>
                  </TableCell>
                  <TableCell className="py-2.5 font-mono text-xs text-muted-foreground">
                    {member.User?.IsSuperAdmin ? 'Yes' : 'No'}
                  </TableCell>
                  <TableCell className="text-right py-2.5">
                    <Button 
                      variant="ghost" 
                      size="icon" 
                      onClick={() => handleRemove(String(member.UserID))}
                      className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
              {members.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="h-24 text-center font-mono text-xs text-muted-foreground">
                    No members found.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Progress } from '@/components/ui/progress'
import { ScrollArea } from '@/components/ui/scroll-area'
import { piiEntries } from '../data/mock.ts'

const TYPE_STYLES: Record<string, string> = {
  EMAIL:         'text-primary border-primary/30 bg-primary/8',
  PHONE:         'text-violet-400 border-violet-400/30 bg-violet-400/8',
  FULL_NAME:     'text-green-400 border-green-400/30 bg-green-400/8',
  SSN:           'text-red-400 border-red-400/30 bg-red-400/8',
  BANK_ACCT:     'text-amber-400 border-amber-400/30 bg-amber-400/8',
  ADDRESS:       'text-blue-400 border-blue-400/30 bg-blue-400/8',
  DATE_OF_BIRTH: 'text-pink-400 border-pink-400/30 bg-pink-400/8',
}

const typeCounts = piiEntries.reduce<Record<string, number>>((acc, e) => {
  acc[e.type] = (acc[e.type] ?? 0) + 1
  return acc
}, {})

export default function PiiVault() {
  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">PII Vault</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Session vault · Redis-backed · 24h TTL</p>
        </div>
        <Badge variant="outline" className="font-mono text-green-400 border-green-400/30 bg-green-400/6">
          AES-256 · BYO KMS
        </Badge>
      </div>

      {/* Stats + breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <div className="space-y-3">
          {[
            { label: 'ACTIVE TOKENS',      value: piiEntries.length,                                color: 'text-primary' },
            { label: 'ENTITY TYPES',        value: Object.keys(typeCounts).length,                  color: 'text-violet-400' },
            { label: 'VAULT DEREFERENCES',  value: piiEntries.reduce((s, e) => s + e.hits, 0),     color: 'text-green-400' },
          ].map(({ label, value, color }) => (
            <Card key={label}>
              <CardHeader className="pb-2">
                <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">{label}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className={`font-mono text-2xl font-black ${color}`}>{value}</p>
              </CardContent>
            </Card>
          ))}
        </div>

        <Card className="col-span-1 lg:col-span-2">
          <CardHeader className="pb-3">
            <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">ENTITY BREAKDOWN</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(typeCounts).map(([type, count]) => {
              const pct = Math.round((count / piiEntries.length) * 100)
              return (
                <div key={type}>
                  <div className="flex items-center justify-between mb-1.5">
                    <Badge variant="outline" className={`font-mono text-[9px] ${TYPE_STYLES[type] ?? ''}`}>{type}</Badge>
                    <span className="font-mono text-xs text-muted-foreground">{count} · {pct}%</span>
                  </div>
                  <Progress value={pct} className="h-1.5" />
                </div>
              )
            })}
          </CardContent>
        </Card>
      </div>

      {/* Token table */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">ACTIVE TOKENS</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <ScrollArea>
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  {['TOKEN', 'TYPE', 'AGENT', 'MASKED AT', 'TTL', 'DEREFERENCES'].map((h) => (
                    <TableHead key={h} className="font-mono text-[10px] tracking-widest text-muted-foreground/50">{h}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {piiEntries.map((e) => (
                  <TableRow key={e.token} className="border-border">
                    <TableCell>
                      <Badge variant="outline" className="font-mono text-xs text-primary border-primary/20 bg-primary/6">
                        {e.token}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className={`font-mono text-[10px] ${TYPE_STYLES[e.type] ?? ''}`}>{e.type}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{e.agent}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{e.maskedAt}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{e.ttl}</TableCell>
                    <TableCell className={`font-mono text-xs font-bold ${e.hits > 0 ? 'text-violet-400' : 'text-muted-foreground/30'}`}>
                      {e.hits}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>

      {/* Masking flow */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">MASKING FLOW</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 text-xs font-mono overflow-x-auto pb-1 flex-wrap">
            {[
              { label: 'User Input',     style: 'text-muted-foreground border-border bg-muted/30' },
              null,
              { label: 'Detect Entities', style: 'text-primary border-primary/30 bg-primary/8' },
              null,
              { label: 'Assign Tokens',  style: 'text-violet-400 border-violet-400/30 bg-violet-400/8' },
              null,
              { label: 'Store in Vault', style: 'text-green-400 border-green-400/30 bg-green-400/8' },
              null,
              { label: 'Masked → LLM',   style: 'text-primary border-primary/30 bg-primary/8' },
              null,
              { label: 'Deref Output',   style: 'text-violet-400 border-violet-400/30 bg-violet-400/8' },
              null,
              { label: 'User Sees Plain', style: 'text-muted-foreground border-border bg-muted/30' },
            ].map((item, i) =>
              item === null
                ? <span key={i} className="text-muted-foreground/30">→</span>
                : <Badge key={i} variant="outline" className={`font-mono text-[10px] whitespace-nowrap ${item.style}`}>{item.label}</Badge>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

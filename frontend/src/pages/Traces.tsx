import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { STATUS_STYLES } from '@/lib/styles.ts'
import { traces } from '../data/mock.ts'
import type { Trace } from '../data/mock.ts'

const ALL_STATUSES = ['OK', 'GUARDRAIL', 'PII MASKED', 'FALLBACK', 'ERROR']
const ALL_MODELS   = ['gpt-4o', 'claude-3.5', 'gemini-pro']

function StatusBadge({ status }: { status: string }) {
  return (
    <Badge variant="outline" className={`font-mono text-[10px] ${STATUS_STYLES[status] ?? ''}`}>
      {status}
    </Badge>
  )
}

function TraceDetail({ trace }: { trace: Trace }) {
  const total = 30 + parseInt(trace.ttft) + 4
  const spans = [
    { name: 'guardrail_eval', ms: 18,                  offset: 0,                     color: '#f59e0b' },
    { name: 'pii_scan',       ms: 12,                  offset: 18,                    color: '#a78bfa' },
    { name: 'llm_request',    ms: parseInt(trace.ttft), offset: 30,                   color: '#22d3ee' },
    { name: 'pii_unmask',     ms: 4,                   offset: 30 + parseInt(trace.ttft), color: '#a78bfa' },
  ]

  const fields = [
    ['AGENT',         `${trace.agent} · ${trace.agentSub}`],
    ['THREAD',        trace.threadId],
    ['MODEL',         trace.model],
    ['TIMESTAMP',     trace.ts],
    ['INPUT TOKENS',  trace.inputTokens.toLocaleString()],
    ['OUTPUT TOKENS', trace.outputTokens.toLocaleString()],
    ['TOTAL TOKENS',  (trace.inputTokens + trace.outputTokens).toLocaleString()],
    ['COST',          trace.cost],
    ['TTFT',          trace.ttft],
    ['AVG TPS',       trace.tps],
  ] as const

  return (
    <div className="space-y-5 pt-2">
      <StatusBadge status={trace.status} />
      <div className="space-y-0">
        {fields.map(([label, value]) => (
          <div key={label} className="flex justify-between items-center py-2 border-b border-border/50">
            <span className="font-mono text-[10px] tracking-widest text-muted-foreground">{label}</span>
            <span className="font-mono text-xs text-foreground">{value}</span>
          </div>
        ))}
      </div>
      <Separator />
      <div>
        <p className="font-mono text-[10px] tracking-widest text-muted-foreground mb-3">SPAN WATERFALL</p>
        <div className="space-y-2">
          {spans.map((span) => (
            <div key={span.name}>
              <div className="flex justify-between mb-1">
                <span className="font-mono text-[9px] text-muted-foreground">{span.name}</span>
                <span className="font-mono text-[9px]" style={{ color: span.color }}>{span.ms}ms</span>
              </div>
              <div className="h-2.5 rounded-sm overflow-hidden bg-muted">
                <div
                  className="h-full rounded-sm"
                  style={{ marginLeft: `${(span.offset / total) * 100}%`, width: `${(span.ms / total) * 100}%`, background: span.color, opacity: 0.65 }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default function Traces() {
  const [statusFilter, setStatusFilter] = useState<string>('ALL')
  const [modelFilter,  setModelFilter]  = useState<string>('ALL')
  const [selected, setSelected] = useState<Trace | null>(null)

  const filtered = traces.filter((t) =>
    (statusFilter === 'ALL' || t.status === statusFilter) &&
    (modelFilter  === 'ALL' || t.model  === modelFilter)
  )

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Traces</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Flight log · {traces.length} entries</p>
        </div>
        <div className="flex gap-2">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="font-mono text-xs w-36">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All statuses</SelectItem>
              {ALL_STATUSES.map((s) => <SelectItem key={s} value={s}>{s}</SelectItem>)}
            </SelectContent>
          </Select>
          <Select value={modelFilter} onValueChange={setModelFilter}>
            <SelectTrigger className="font-mono text-xs w-36">
              <SelectValue placeholder="All models" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All models</SelectItem>
              {ALL_MODELS.map((m) => <SelectItem key={m} value={m}>{m}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Status strip */}
      <div className="grid grid-cols-5 gap-2">
        {ALL_STATUSES.map((s) => (
          <Card key={s} className={`cursor-pointer transition-colors ${statusFilter === s ? 'ring-1 ring-primary/40' : ''}`} onClick={() => setStatusFilter(statusFilter === s ? 'ALL' : s)}>
            <CardContent className="p-3">
              <Badge variant="outline" className={`font-mono text-[9px] mb-2 ${STATUS_STYLES[s]}`}>{s}</Badge>
              <p className="font-mono text-xl font-black text-foreground">{traces.filter(t => t.status === s).length}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Trace table */}
      <Card>
        <CardHeader className="pb-2 flex-row items-center justify-between space-y-0">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">TRACE LOG</CardTitle>
          <span className="font-mono text-xs text-muted-foreground/40">{filtered.length} results</span>
        </CardHeader>
        <CardContent className="p-0">
          <ScrollArea>
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  {['ID', 'AGENT', 'THREAD', 'MODEL', 'IN', 'OUT', 'COST', 'TTFT', 'TPS', 'STATUS'].map((h) => (
                    <TableHead key={h} className="font-mono text-[10px] tracking-widest text-muted-foreground/50">{h}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((t) => (
                  <TableRow
                    key={t.id}
                    className="border-border cursor-pointer"
                    onClick={() => setSelected(t)}
                  >
                    <TableCell className="font-mono text-xs text-primary">{t.id}</TableCell>
                    <TableCell className="font-mono text-xs">
                      <span className="text-foreground">{t.agent}</span>
                      <span className="ml-1.5 text-muted-foreground/40">{t.agentSub}</span>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.threadId}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.model}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.inputTokens.toLocaleString()}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.outputTokens.toLocaleString()}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.cost}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.ttft}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.tps}</TableCell>
                    <TableCell><StatusBadge status={t.status} /></TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>

      <Sheet open={!!selected} onOpenChange={(open) => !open && setSelected(null)}>
        <SheetContent className="w-96 bg-card border-l border-border overflow-y-auto">
          {selected && (
            <>
              <SheetHeader>
                <SheetTitle className="font-mono text-sm text-foreground">{selected.id}</SheetTitle>
              </SheetHeader>
              <TraceDetail trace={selected} />
            </>
          )}
        </SheetContent>
      </Sheet>
    </div>
  )
}

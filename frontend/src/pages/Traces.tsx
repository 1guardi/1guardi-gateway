import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { STATUS_STYLES } from '@/lib/styles.ts'
import { useTraces, useTraceSpans } from '@/api/traces'
import type { TraceRow, TraceSpan } from '@/api/traces'

// ── span waterfall ────────────────────────────────────────────────────────────

const SPAN_COLORS: Record<string, string> = {
  proxy:                 '#4b5563',
  agent:                 '#6b7280',
  thread:                '#6b7280',
  'guardrail.execution': '#f59e0b',
  'llm.generation':      '#22d3ee',
  'upstream.call':       '#6366f1',
}

interface SpanNode extends TraceSpan {
  children: SpanNode[]
  depth: number
  offsetMs: number
}

function buildSpanTree(spans: TraceSpan[]): SpanNode[] {
  if (!spans.length) return []

  const nodeMap = new Map<string, SpanNode>()
  const traceStartMs = Math.min(...spans.map(s => s.start_time_ms))

  for (const span of spans) {
    nodeMap.set(span.span_id, {
      ...span,
      children: [],
      depth: 0,
      offsetMs: span.start_time_ms - traceStartMs,
    })
  }

  const roots: SpanNode[] = []
  for (const node of nodeMap.values()) {
    const parent = nodeMap.get(node.parent_span_id)
    if (parent) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }

  function assignDepth(node: SpanNode, depth: number) {
    node.depth = depth
    for (const child of node.children) assignDepth(child, depth + 1)
  }
  for (const root of roots) assignDepth(root, 0)

  const ordered: SpanNode[] = []
  function flatten(node: SpanNode) {
    ordered.push(node)
    const sorted = [...node.children].sort((a, b) => a.start_time_ms - b.start_time_ms)
    for (const child of sorted) flatten(child)
  }
  const sortedRoots = [...roots].sort((a, b) => a.start_time_ms - b.start_time_ms)
  for (const root of sortedRoots) flatten(root)

  return ordered
}

function SpanAttributes({ span }: { span: SpanNode }) {
  const attrs = span.attributes
  const prompt = attrs['gen_ai.prompt']
  const completion = attrs['gen_ai.completion']
  const scope = attrs['guardrail.scope']

  const metaKeys = Object.keys(attrs).filter(
    k => k !== 'gen_ai.prompt' && k !== 'gen_ai.completion'
  )

  return (
    <div className="mt-4 space-y-3">
      <p className="font-mono text-[10px] tracking-widest text-muted-foreground">
        {span.span_name.toUpperCase()} ATTRIBUTES
      </p>
      {scope && (
        <div className="flex justify-between py-1.5 border-b border-border/40">
          <span className="font-mono text-[10px] text-muted-foreground">guardrail.scope</span>
          <Badge variant="outline" className="font-mono text-[9px] h-4">
            {scope}
          </Badge>
        </div>
      )}
      {metaKeys.filter(k => k !== 'guardrail.scope').map(k => (
        <div key={k} className="flex justify-between items-start py-1.5 border-b border-border/40 gap-2">
          <span className="font-mono text-[10px] text-muted-foreground shrink-0">{k}</span>
          <span className="font-mono text-[10px] text-foreground text-right break-all">{attrs[k]}</span>
        </div>
      ))}
      {prompt && (
        <div className="space-y-1">
          <p className="font-mono text-[10px] tracking-widest text-muted-foreground">PROMPT</p>
          <div className="rounded bg-muted/40 p-2 max-h-32 overflow-y-auto">
            <pre className="font-mono text-[10px] text-foreground whitespace-pre-wrap break-all">{prompt}</pre>
          </div>
        </div>
      )}
      {completion && (
        <div className="space-y-1">
          <p className="font-mono text-[10px] tracking-widest text-muted-foreground">COMPLETION</p>
          <div className="rounded bg-muted/40 p-2 max-h-32 overflow-y-auto">
            <pre className="font-mono text-[10px] text-foreground whitespace-pre-wrap break-all">{completion}</pre>
          </div>
        </div>
      )}
    </div>
  )
}

function SpanWaterfall({ spans, selectedSpanId, onSelectSpan }: {
  spans: TraceSpan[]
  selectedSpanId: string | null
  onSelectSpan: (span: SpanNode) => void
}) {
  const nodes = buildSpanTree(spans)
  const totalMs = nodes.reduce((max, n) => Math.max(max, n.offsetMs + n.duration_ms), 1)

  return (
    <div className="space-y-0.5">
      {nodes.map(node => {
        const color = SPAN_COLORS[node.span_name] ?? '#6b7280'
        const indent = node.depth * 10
        const isSelected = node.span_id === selectedSpanId
        return (
          <div
            key={node.span_id}
            className={`cursor-pointer rounded px-2 py-1.5 transition-colors ${isSelected ? 'bg-muted ring-1 ring-border' : 'hover:bg-muted/40'}`}
            onClick={() => onSelectSpan(node)}
          >
            <div className="flex items-center gap-1.5 mb-1">
              <div style={{ width: indent, flexShrink: 0 }} />
              {node.depth > 0 && (
                <span className="text-muted-foreground/40 font-mono text-[9px] mr-0.5">└</span>
              )}
              <span className="font-mono text-[10px]" style={{ color }}>{node.span_name}</span>
              <span className="font-mono text-[9px] text-muted-foreground ml-auto">
                {node.duration_ms < 1 ? `${(node.duration_ms * 1000).toFixed(0)}μs` : `${node.duration_ms.toFixed(1)}ms`}
              </span>
            </div>
            <div className="h-1.5 rounded-sm overflow-hidden bg-muted" style={{ marginLeft: indent + (node.depth > 0 ? 12 : 0) }}>
              <div
                className="h-full rounded-sm opacity-60"
                style={{
                  marginLeft: `${(node.offsetMs / totalMs) * 100}%`,
                  width: `${Math.max((node.duration_ms / totalMs) * 100, 0.5)}%`,
                  background: color,
                }}
              />
            </div>
          </div>
        )
      })}
    </div>
  )
}

// ── trace detail panel ────────────────────────────────────────────────────────

function TraceDetail({ trace, tenantId }: { trace: TraceRow; tenantId: string | null }) {
  const [selectedSpan, setSelectedSpan] = useState<SpanNode | null>(null)
  const { data: spans = [], isLoading } = useTraceSpans(tenantId, trace.trace_id)

  const fields: [string, string][] = [
    ['AGENT',         trace.agent_id || '—'],
    ['THREAD',        trace.thread_id || '—'],
    ['MODEL',         trace.model || '—'],
    ['TIMESTAMP',     trace.ts],
    ['INPUT TOKENS',  trace.input_tokens.toLocaleString()],
    ['OUTPUT TOKENS', trace.output_tokens.toLocaleString()],
    ['TOTAL TOKENS',  (trace.input_tokens + trace.output_tokens).toLocaleString()],
    ['COST',          `$${trace.cost.toFixed(4)}`],
    ['TTFT',          `${trace.ttft_ms.toFixed(0)}ms`],
    ['AVG TPS',       trace.tps.toFixed(1)],
    ['DURATION',      `${trace.duration_ms.toFixed(0)}ms`],
  ]

  return (
    <div className="space-y-5 p-6">
      <Badge variant="outline" className={`font-mono text-[10px] ${STATUS_STYLES[trace.status] ?? ''}`}>
        {trace.status}
      </Badge>

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
        {isLoading ? (
          <div className="space-y-2">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="h-8 rounded bg-muted animate-pulse" />
            ))}
          </div>
        ) : spans.length === 0 ? (
          <p className="font-mono text-[10px] text-muted-foreground/50">No span data available</p>
        ) : (
          <SpanWaterfall
            spans={spans}
            selectedSpanId={selectedSpan?.span_id ?? null}
            onSelectSpan={setSelectedSpan}
          />
        )}
      </div>

      {selectedSpan && (
        <>
          <Separator />
          <SpanAttributes span={selectedSpan} />
        </>
      )}
    </div>
  )
}

// ── main page ─────────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: string }) {
  return (
    <Badge variant="outline" className={`font-mono text-[10px] ${STATUS_STYLES[status] ?? ''}`}>
      {status}
    </Badge>
  )
}

const DISPLAY_STATUSES = ['OK', 'GUARDRAIL', 'ERROR']

export default function Traces({
  selectedAgent,
  tenantId,
}: {
  selectedAgent: string
  tenantId: string | null
}) {
  const [statusFilter, setStatusFilter] = useState<string>('ALL')
  const [modelFilter, setModelFilter] = useState<string>('ALL')
  const [selected, setSelected] = useState<TraceRow | null>(null)

  const agentId = selectedAgent !== 'all' ? selectedAgent : undefined
  const { data: traces = [], isLoading, isError } = useTraces(tenantId, agentId)

  const models = [...new Set(traces.map(t => t.model).filter(Boolean))]

  const filtered = traces.filter(t =>
    (statusFilter === 'ALL' || t.status === statusFilter) &&
    (modelFilter === 'ALL' || t.model === modelFilter)
  )

  const countByStatus = DISPLAY_STATUSES.reduce<Record<string, number>>((acc, s) => {
    acc[s] = traces.filter(t => t.status === s).length
    return acc
  }, {})

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Traces</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">
            Flight log · {isLoading ? '…' : `${traces.length} entries`}
          </p>
        </div>
        <div className="flex gap-2">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="font-mono text-xs w-36">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All statuses</SelectItem>
              {DISPLAY_STATUSES.map(s => <SelectItem key={s} value={s}>{s}</SelectItem>)}
            </SelectContent>
          </Select>
          <Select value={modelFilter} onValueChange={setModelFilter}>
            <SelectTrigger className="font-mono text-xs w-40">
              <SelectValue placeholder="All models" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All models</SelectItem>
              {models.map(m => <SelectItem key={m} value={m}>{m}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Status strip */}
      <div className="grid grid-cols-3 gap-2">
        {DISPLAY_STATUSES.map(s => (
          <Card
            key={s}
            className={`cursor-pointer transition-colors ${statusFilter === s ? 'ring-1 ring-primary/40' : ''}`}
            onClick={() => setStatusFilter(statusFilter === s ? 'ALL' : s)}
          >
            <CardContent className="p-3">
              <Badge variant="outline" className={`font-mono text-[9px] mb-2 ${STATUS_STYLES[s] ?? ''}`}>{s}</Badge>
              <p className="font-mono text-xl font-black text-foreground">
                {isLoading ? '—' : countByStatus[s] ?? 0}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Trace table */}
      <Card>
        <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">TRACE LOG</CardTitle>
          <span className="font-mono text-xs text-muted-foreground/40">{filtered.length} results</span>
        </CardHeader>
        <CardContent className="p-0">
          {isError ? (
            <div className="p-6 text-center">
              <p className="font-mono text-xs text-error">Failed to load traces</p>
            </div>
          ) : isLoading ? (
            <div className="p-4 space-y-2">
              {[1, 2, 3, 4, 5].map(i => (
                <div key={i} className="h-10 rounded bg-muted animate-pulse" />
              ))}
            </div>
          ) : (
            <ScrollArea>
              <Table>
                <TableHeader>
                  <TableRow className="border-border hover:bg-transparent">
                    {['TRACE ID', 'AGENT', 'THREAD', 'MODEL', 'IN', 'OUT', 'COST', 'TTFT', 'TPS', 'STATUS'].map(h => (
                      <TableHead key={h} className="font-mono text-[10px] tracking-widest text-muted-foreground/50">{h}</TableHead>
                    ))}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={10} className="text-center py-10">
                        <p className="font-mono text-xs text-muted-foreground/50">No traces found</p>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filtered.map(t => (
                      <TableRow
                        key={t.trace_id}
                        className="border-border cursor-pointer transition-colors hover:bg-muted/30"
                        onClick={() => setSelected(t)}
                      >
                        <TableCell className="font-mono text-xs text-primary">
                          {t.trace_id.slice(0, 8)}
                        </TableCell>
                        <TableCell className="font-mono text-xs text-foreground">
                          {t.agent_id || <span className="text-muted-foreground/50">—</span>}
                        </TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">
                          {t.thread_id
                            ? t.thread_id.slice(0, 8)
                            : <span className="text-muted-foreground/30">—</span>}
                        </TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">{t.model}</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">{t.input_tokens.toLocaleString()}</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">{t.output_tokens.toLocaleString()}</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">${t.cost.toFixed(4)}</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">{t.ttft_ms.toFixed(0)}ms</TableCell>
                        <TableCell className="font-mono text-xs text-muted-foreground">{t.tps.toFixed(1)}</TableCell>
                        <TableCell><StatusBadge status={t.status} /></TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </ScrollArea>
          )}
        </CardContent>
      </Card>

      <Sheet open={!!selected} onOpenChange={open => !open && setSelected(null)}>
        <SheetContent className="w-[420px] bg-card border-l border-border overflow-y-auto p-0 gap-0">
          {selected && (
            <>
              <SheetHeader className="px-6 py-5 border-b border-border/50">
                <SheetTitle className="font-mono text-sm text-foreground">
                  {selected.trace_id.slice(0, 16)}…
                </SheetTitle>
                <p className="font-mono text-[10px] text-muted-foreground">{selected.ts}</p>
              </SheetHeader>
              <TraceDetail trace={selected} tenantId={tenantId} />
            </>
          )}
        </SheetContent>
      </Sheet>
    </div>
  )
}

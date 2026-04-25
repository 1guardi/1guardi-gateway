import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ScrollArea } from '@/components/ui/scroll-area'
import { STATUS_STYLES, CIRCUIT_STYLES } from '@/lib/styles.ts'
import { traces, endpoints, costData } from '../data/mock.ts'

const W = 460, H = 72
const pts = costData.map((v, i) => [i * (W / (costData.length - 1)), (1 - v) * H] as [number, number])
const smooth = pts.map(([x, y], i) => {
  if (i === 0) return `M ${x} ${y}`
  const [px, py] = pts[i - 1]
  return `C ${px + (x - px) / 3} ${py} ${x - (x - px) / 3} ${y} ${x} ${y}`
}).join(' ')

function StatCard({ label, value, delta, good, sub }: { label: string; value: string; delta: string; good: boolean; sub: string }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">{label}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-black font-mono text-foreground mb-1">{value}</p>
        <div className="flex items-center gap-1.5">
          <span className={`text-xs font-mono ${good ? 'text-success' : 'text-warning'}`}>{delta}</span>
          <span className="text-xs font-mono text-muted-foreground/50">{sub}</span>
        </div>
      </CardContent>
    </Card>
  )
}

export default function Overview() {
  const totalCost = traces.reduce((s, t) => s + parseFloat(t.cost.replace('$', '')), 0)

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Tower View</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Last 24 hours · acme-corp</p>
        </div>
        <Badge variant="outline" className="font-mono gap-1.5 text-primary border-primary/30 bg-primary/6">
          <span className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
          LIVE
        </Badge>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard label="TTFT P99"        value="124ms"                              delta="↓ 12%" good={true}  sub="vs yesterday" />
        <StatCard label="COST / 24H"      value={`$${totalCost.toFixed(3)}`}          delta="↑ 8%"  good={false} sub={`${traces.length} traces`} />
        <StatCard label="GUARDRAIL FIRES" value={String(traces.filter(t => t.status === 'GUARDRAIL').length)} delta="↑ 3" good={false} sub="this window" />
        <StatCard label="PII DETECTIONS"  value={String(traces.filter(t => t.status === 'PII MASKED').length)} delta="stable" good={true} sub="masked + vaulted" />
      </div>

      {/* Chart + circuit breakers */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3">
        <Card className="col-span-1 lg:col-span-2">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">COST / HOUR</CardTitle>
              <div className="flex gap-3">
                {[{ l: 'gpt-4o', c: 'bg-primary' }, { l: 'claude-3.5', c: 'bg-violet' }, { l: 'gemini', c: 'bg-warning' }].map(({ l, c }) => (
                  <div key={l} className="flex items-center gap-1">
                    <div className={`w-2 h-2 rounded-full ${c}`} />
                    <span className="text-xs font-mono text-muted-foreground">{l}</span>
                  </div>
                ))}
              </div>
            </div>
            <p className="text-sm font-mono font-bold text-foreground">$0.038 <span className="text-xs font-normal text-muted-foreground">current hour</span></p>
          </CardHeader>
          <CardContent>
            <svg viewBox={`0 0 ${W} ${H + 20}`} className="w-full" preserveAspectRatio="none" style={{ height: '72px' }}>
              <defs>
                <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#22d3ee" stopOpacity="0.15" />
                  <stop offset="100%" stopColor="#22d3ee" stopOpacity="0" />
                </linearGradient>
              </defs>
              {[0, 0.25, 0.5, 0.75, 1].map((v) => (
                <line key={v} x1="0" y1={v * H} x2={W} y2={v * H} stroke="var(--grid-color)" strokeWidth="1" />
              ))}
              <path d={`${smooth} L ${W} ${H} L 0 ${H} Z`} fill="url(#areaGrad)" />
              <path d={smooth} fill="none" stroke="#22d3ee" strokeWidth="1.5" />
              <circle cx={pts[pts.length - 1][0]} cy={pts[pts.length - 1][1]} r="3" fill="#22d3ee" />
              {[0, 6, 12, 18, 23].map((h) => (
                <text key={h} x={h * (W / 23)} y={H + 14} textAnchor="middle" fill="currentColor" fillOpacity="0.2" fontSize="8" fontFamily="monospace">
                  {h === 0 ? '00:00' : h === 23 ? 'now' : `${String(h).padStart(2, '0')}:00`}
                </text>
              ))}
            </svg>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">CIRCUIT BREAKERS</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {endpoints.map((ep) => (
              <div key={ep.id} className="flex items-center justify-between rounded-md px-3 py-2 bg-background/60 border border-border">
                <div>
                  <p className="text-xs font-mono font-bold text-foreground">{ep.label}</p>
                  <p className="text-[10px] font-mono text-muted-foreground/40">p99 · {ep.ttftP99}</p>
                </div>
                <Badge variant="outline" className={`font-mono text-[10px] ${CIRCUIT_STYLES[ep.circuitState]}`}>{ep.circuitState}</Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Recent traces */}
      <Card>
        <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">RECENT TRACES</CardTitle>
          <span className="text-xs font-mono text-muted-foreground/40">showing {traces.length} of 1,847</span>
        </CardHeader>
        <CardContent className="p-0">
          <ScrollArea>
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  {['AGENT', 'MODEL', 'TOKENS', 'COST', 'TTFT', 'STATUS'].map((h) => (
                    <TableHead key={h} className="font-mono text-[10px] tracking-widest text-muted-foreground/50">{h}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {traces.slice(0, 6).map((t) => (
                  <TableRow key={t.id} className="border-border">
                    <TableCell className="font-mono text-xs">
                      <span className="text-foreground">{t.agent}</span>
                      <span className="ml-2 text-muted-foreground/40">{t.agentSub}</span>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.model}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{(t.inputTokens + t.outputTokens).toLocaleString()}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.cost}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">{t.ttft}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className={`font-mono text-[10px] ${STATUS_STYLES[t.status]}`}>{t.status}</Badge>
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

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { CIRCUIT_STYLES, scoreColor, quotaColor } from '@/lib/styles.ts'
import { useRouterEndpoints } from '../api/router.ts'

interface EndpointVM {
  id: string
  label: string
  provider: string
  model: string
  ttftP50: string
  ttftP99: string
  avgTps: string
  errorRate: string
  errorRateNum: number
  quotaUsed: number
  circuitState: 'CLOSED' | 'OPEN' | 'HALF-OPEN'
  score: number
}

function toVM(e: any): EndpointVM {
  return {
    id: String(e.id || e.ID || ''),
    label: e.label || String(e.ID || ''),
    provider: e.provider || e.Provider || 'unknown',
    model: e.model || e.Model || '',
    ttftP50: e.ttft_p50_ms > 0 ? `${Math.round(e.ttft_p50_ms)}ms` : '—',
    ttftP99: e.ttft_p99_ms > 0 ? `${Math.round(e.ttft_p99_ms)}ms` : '—',
    avgTps:  e.avg_tps > 0    ? `${Math.round(e.avg_tps)}`         : '—',
    errorRate: `${(e.error_rate * 100 || 0).toFixed(1)}%`,
    errorRateNum: e.error_rate || 0,
    quotaUsed: e.quota_used || 0,
    circuitState: e.circuit_state || 'CLOSED',
    score: e.score || 0,
  }
}

export default function Router({ selectedAgent, tenantId }: { selectedAgent: string; tenantId: string | null }) {
  const { data: rawEndpoints = [] } = useRouterEndpoints()
  const endpoints = rawEndpoints
    .filter((e) => !tenantId || String(e.tenant_id) === tenantId)
    .map(toVM)

  const grouped = endpoints.reduce((acc, ep) => {
    if (!acc[ep.provider]) acc[ep.provider] = []
    acc[ep.provider].push(ep)
    return acc
  }, {} as Record<string, EndpointVM[]>)

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Router</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">
            ATC · Fallback-first routing · {endpoints.length} endpoints
            {selectedAgent !== 'all' && ` · showing weights for ${selectedAgent}`}
          </p>
        </div>
      </div>

      {/* Circuit state summary */}
      <div className="grid grid-cols-3 gap-3">
        {(['CLOSED', 'HALF-OPEN', 'OPEN'] as const).map((state) => (
          <Card key={state}>
            <CardHeader className="pb-2">
              <CardTitle className="font-mono text-[10px] tracking-widest">
                <Badge variant="outline" className={`font-mono text-[10px] ${CIRCUIT_STYLES[state]}`}>{state}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className={`font-mono text-2xl font-black ${CIRCUIT_STYLES[state].split(' ')[0]}`}>
                {endpoints.filter((e) => e.circuitState === state).length}
              </p>
              <p className="font-mono text-[9px] text-muted-foreground/40 mt-1">endpoints</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Endpoint tables by provider */}
      <div className="space-y-8">
        {Object.entries(grouped).map(([provider, providerEndpoints]) => (
          <div key={provider} className="space-y-3">
            <div className="flex items-center gap-2 px-1">
              <p className="font-mono text-[10px] font-bold tracking-widest text-muted-foreground uppercase">{provider}</p>
              <div className="h-px flex-1 bg-border/50" />
              <Badge variant="outline" className="font-mono text-[9px] text-muted-foreground/50">
                {providerEndpoints.length} MODELS
              </Badge>
            </div>

            <div className="rounded-xl border bg-card overflow-hidden">
              <Table>
                <TableHeader className="bg-muted/30">
                  <TableRow className="hover:bg-transparent border-b border-border/50">
                    <TableHead className="font-mono text-[9px] tracking-widest h-10">MODEL</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10">CIRCUIT</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10 text-right">P50</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10 text-right">P99</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10 text-right">TPS</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10 text-right">ERR %</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10 text-right">QUOTA</TableHead>
                    <TableHead className="font-mono text-[9px] tracking-widest h-10 text-right w-24">SCORE</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {providerEndpoints
                    .sort((a, b) => b.score - a.score)
                    .map((ep) => (
                      <TableRow key={ep.id} className="hover:bg-muted/20 border-border/40">
                        <TableCell>
                          <div className="py-1">
                            <p className="font-bold text-sm text-foreground">{ep.label}</p>
                            <p className="font-mono text-[10px] text-muted-foreground">{ep.model}</p>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className={`font-mono text-[9px] ${CIRCUIT_STYLES[ep.circuitState]}`}>
                            {ep.circuitState}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs text-right text-muted-foreground">{ep.ttftP50}</TableCell>
                        <TableCell className="font-mono text-xs text-right text-muted-foreground">{ep.ttftP99}</TableCell>
                        <TableCell className="font-mono text-xs text-right text-muted-foreground">{ep.avgTps}</TableCell>
                        <TableCell className={`font-mono text-xs text-right font-bold ${ep.errorRateNum > 0.05 ? 'text-error' : 'text-muted-foreground'}`}>
                          {ep.errorRate}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-col items-end gap-1">
                            <span className={`font-mono text-[10px] font-bold ${quotaColor(ep.quotaUsed)}`}>
                              {ep.quotaUsed}%
                            </span>
                            <Progress value={ep.quotaUsed} className="h-1 w-16" />
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-col items-end gap-1">
                            <span className={`font-mono text-sm font-black ${scoreColor(ep.score)}`}>
                              {(ep.score * 100).toFixed(0)}
                            </span>
                            <Progress value={ep.score * 100} className="h-1 w-20" />
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </div>
          </div>
        ))}
      </div>

      {/* Scoring weights */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">
            SCORING WEIGHTS · scope: {selectedAgent === 'all' ? 'GLOBAL DEFAULT' : selectedAgent}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4">
            {[
              { label: 'w₁  TTFT WEIGHT',  value: 0.5, hint: 'interactive agents prioritise TTFT' },
              { label: 'w₂  TPS WEIGHT',   value: 0.3, hint: 'long-form agents prioritise TPS' },
              { label: 'w₃  ERROR WEIGHT', value: 0.2, hint: 'stability floor' },
            ].map(({ label, value, hint }) => (
              <div key={label} className="rounded-lg p-4 bg-muted/30 border border-border/50">
                <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-2">{label}</p>
                <p className="font-mono text-2xl font-black text-foreground">{value}</p>
                <Separator className="my-2" />
                <p className="font-mono text-[9px] text-muted-foreground/50">{hint}</p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}


import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { CIRCUIT_STYLES, scoreColor, quotaColor } from '@/lib/styles.ts'
import { useRouterEndpoints } from '../api/router.ts'

interface EndpointVM {
  id: string
  label: string
  model: string
  ttftP50: string
  ttftP99: string
  avgTps: string
  errorRate: string
  quotaUsed: number
  circuitState: 'CLOSED' | 'OPEN' | 'HALF-OPEN'
  score: number
}

function toVM(e: any): EndpointVM {
  return {
    id: String(e.id || e.ID || ''),
    label: e.label || String(e.ID || ''),
    model: e.model || e.Model || '',
    ttftP50: e.ttft_p50_ms > 0 ? `${Math.round(e.ttft_p50_ms)}ms` : '—',
    ttftP99: e.ttft_p99_ms > 0 ? `${Math.round(e.ttft_p99_ms)}ms` : '—',
    avgTps:  e.avg_tps > 0    ? `${Math.round(e.avg_tps)}`         : '—',
    errorRate: `${(e.error_rate * 100 || 0).toFixed(1)}%`,
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

      {/* Endpoint cards */}
      <div className="space-y-3">
        <p className="font-mono text-[10px] font-bold tracking-widest text-muted-foreground px-1">ENDPOINT HEALTH</p>
        {endpoints.map((ep) => (
          <Card key={ep.id}>
            <CardContent className="p-5">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <div className="flex items-center gap-2.5 mb-1">
                    <div className={`w-2 h-2 rounded-full flex-shrink-0 ${CIRCUIT_STYLES[ep.circuitState].split(' ')[0].replace('text-', 'bg-')}`} />
                    <p className="font-bold text-foreground">{ep.label}</p>
                    <Badge variant="outline" className={`font-mono text-[10px] ${CIRCUIT_STYLES[ep.circuitState]}`}>
                      {ep.circuitState}
                    </Badge>
                  </div>
                  <p className="font-mono text-xs text-muted-foreground ml-4.5">{ep.model}</p>
                </div>
                <div className="text-right">
                  <p className="font-mono text-[10px] text-muted-foreground mb-1">ROUTING SCORE</p>
                  <p className={`font-mono text-2xl font-black ${scoreColor(ep.score)}`}>{(ep.score * 100).toFixed(0)}</p>
                </div>
              </div>

              <div className="grid grid-cols-2 lg:grid-cols-5 gap-4 mb-4">
                {[
                  { label: 'TTFT P50',   value: ep.ttftP50 },
                  { label: 'TTFT P99',   value: ep.ttftP99 },
                  { label: 'AVG TPS',    value: ep.avgTps },
                  { label: 'ERROR RATE', value: ep.errorRate },
                  { label: 'QUOTA USED', value: `${ep.quotaUsed}%` },
                ].map(({ label, value }) => (
                  <div key={label}>
                    <p className="font-mono text-[9px] tracking-widest text-muted-foreground/40 mb-0.5">{label}</p>
                    <p className="font-mono text-sm font-bold text-foreground">{value}</p>
                  </div>
                ))}
              </div>

              <div className="space-y-2">
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="font-mono text-[9px] text-muted-foreground/40">QUOTA CONSUMPTION</span>
                    <span className={`font-mono text-[9px] font-bold ${quotaColor(ep.quotaUsed)}`}>
                      {ep.quotaUsed}%
                    </span>
                  </div>
                  <Progress value={ep.quotaUsed} className="h-1.5" />
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="font-mono text-[9px] text-muted-foreground/40">COMPOSITE SCORE  w₁·(1/TTFT_P99) + w₂·TPS + w₃·(1-err_rate)</span>
                  </div>
                  <Progress value={ep.score * 100} className="h-1.5" />
                </div>
              </div>
            </CardContent>
          </Card>
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
              { label: 'w₁  TTFT WEIGHT',  value: 0.4, hint: 'interactive agents prioritise TTFT' },
              { label: 'w₂  TPS WEIGHT',   value: 0.3, hint: 'long-form agents prioritise TPS' },
              { label: 'w₃  ERROR WEIGHT', value: 0.3, hint: 'stability floor' },
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

import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'

export interface TraceRow {
  trace_id: string
  ts: string
  model: string
  input_tokens: number
  output_tokens: number
  cost: number
  ttft_ms: number
  tps: number
  duration_ms: number
  status: 'OK' | 'ERROR' | 'GUARDRAIL'
  agent_id: string
  thread_id: string
}

export interface TraceSpan {
  span_id: string
  parent_span_id: string
  span_name: string
  duration_ms: number
  start_time_ms: number
  status_code: string
  attributes: Record<string, string>
}

export function useTraces(tenantId: string | null, agentId?: string) {
  return useQuery({
    queryKey: ['traces', tenantId, agentId],
    queryFn: async () => {
      const params = new URLSearchParams({ limit: '200' })
      if (agentId) params.set('agent_id', agentId)
      const { data } = await apiClient.get<TraceRow[]>(
        `/tenants/${tenantId}/traces?${params}`
      )
      return data ?? []
    },
    enabled: !!tenantId,
    refetchInterval: 30_000,
  })
}

export function useTraceSpans(tenantId: string | null, traceId: string | null) {
  return useQuery({
    queryKey: ['trace-spans', tenantId, traceId],
    queryFn: async () => {
      const { data } = await apiClient.get<TraceSpan[]>(
        `/tenants/${tenantId}/traces/${traceId}/spans`
      )
      return data ?? []
    },
    enabled: !!tenantId && !!traceId,
  })
}

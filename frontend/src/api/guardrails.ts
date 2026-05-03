import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface GuardrailCondition {
  type: 'regex' | 'keyword' | 'managed'
  patterns?: string[]
  match_all?: boolean
  rule_id?: string
}

export interface GuardrailRuleResponse {
  // gorm.Model fields (no json tags → Go default serialization, uppercase)
  ID: number
  CreatedAt: string
  // custom fields (explicit lowercase json tags on the Go struct)
  tenant_id: number
  agent_id: number | null
  name: string
  priority: number
  scope: string       // CSV: "input,output"
  direction: string
  condition: string   // JSON string — parse with JSON.parse if needed
  action: string
  mode: string
  managed: boolean
  managed_id: string
  version: string
  enabled: boolean
  fires24h: number    // injected by admin API from ClickHouse
}

export interface GuardrailEvent {
  timestamp: string
  trace_id: string
  rule_id: string
  rule_name: string
  action: string
  reason: string
  scope: string
  agent_id: string
}

export interface CreateRuleRequest {
  name: string
  priority?: number
  scope: string[]
  direction?: string
  condition: GuardrailCondition
  action: string
  mode?: string
  managed?: boolean
  managed_id?: string
  enabled?: boolean
  agent_id?: number
}

export interface UpdateRuleRequest {
  name?: string
  priority?: number
  scope?: string[]
  direction?: string
  condition?: GuardrailCondition
  action?: string
  mode?: string
  enabled?: boolean
  agent_id?: number
}

export function useGuardrailRules(tenantId: string | null) {
  return useQuery({
    queryKey: ['guardrails', tenantId],
    queryFn: async () => {
      const { data } = await apiClient.get<GuardrailRuleResponse[]>(`/tenants/${tenantId}/rules`)
      return data
    },
    enabled: !!tenantId,
  })
}

export function useCreateRule(tenantId: string | null) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (body: CreateRuleRequest) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.post<GuardrailRuleResponse>(`/tenants/${tenantId}/rules`, body)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['guardrails', tenantId] })
    },
  })
}

export function useUpdateRule(tenantId: string | null) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ ruleId, ...body }: UpdateRuleRequest & { ruleId: number }) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.patch<GuardrailRuleResponse>(
        `/tenants/${tenantId}/rules/${ruleId}`,
        body
      )
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['guardrails', tenantId] })
    },
  })
}

export function useGuardrailEvents(tenantId: string | null, ruleId: string | null, limit = 50) {
  return useQuery({
    queryKey: ['guardrail-events', tenantId, ruleId],
    queryFn: async () => {
      const params = new URLSearchParams({ limit: String(limit) })
      if (ruleId) params.set('rule_id', ruleId)
      const { data } = await apiClient.get<GuardrailEvent[]>(
        `/tenants/${tenantId}/guardrail-events?${params}`
      )
      return data
    },
    enabled: !!tenantId,
  })
}

export function useDeleteRule(tenantId: string | null) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (ruleId: number) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      await apiClient.delete(`/tenants/${tenantId}/rules/${ruleId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['guardrails', tenantId] })
    },
  })
}

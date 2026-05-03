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

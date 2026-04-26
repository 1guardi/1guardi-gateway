import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface AgentSummary {
  ID: number
  Name: string
  Description: string
  CreatedAt: string
}

export function useAgents(tenantId: string | null) {
  return useQuery({
    queryKey: ['agents', tenantId],
    queryFn: async () => {
      const { data } = await apiClient.get<AgentSummary[]>(`/tenants/${tenantId}/agents`)
      return data
    },
    enabled: !!tenantId,
  })
}

export function useCreateAgent(tenantId: string | null) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: { name: string; description: string }) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.post<AgentSummary>(`/tenants/${tenantId}/agents`, body)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents', tenantId] })
    },
  })
}

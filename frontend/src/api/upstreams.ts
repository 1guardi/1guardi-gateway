import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface UpstreamResponse {
  ID: number
  CreatedAt: string
  key_id: string
  model: string
  base_url: string
  tenant_id: number
}

export function useUpstreams(tenantId: string | null) {
  return useQuery({
    queryKey: ['upstreams', tenantId],
    queryFn: async () => {
      const { data } = await apiClient.get<UpstreamResponse[]>(`/tenants/${tenantId}/upstreams`)
      return data
    },
    enabled: !!tenantId,
  })
}

export function useCreateUpstream(tenantId: string | null) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: { key_id: string; model: string; base_url: string; api_key: string }) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.post(`/tenants/${tenantId}/upstreams`, body)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['upstreams', tenantId] })
    },
  })
}

export function useDeleteUpstream(tenantId: string | null) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (keyId: string) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      await apiClient.delete(`/tenants/${tenantId}/upstreams/${keyId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['upstreams', tenantId] })
    },
  })
}

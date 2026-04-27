import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface UpstreamResponse {
  ID: number
  CreatedAt: string
  key_id: string
  provider: string
  models: string // Now a comma-separated string from backend
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
    mutationFn: async (body: { key_id: string; provider: string; models: string[]; base_url: string; api_key: string }) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.post(`/tenants/${tenantId}/upstreams`, body)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['upstreams', tenantId] })
    },
  })
}

export function useProviderModels(provider: string, apiKey: string, tenantId?: string | null, upstreamKeyId?: string | null) {
  return useQuery({
    queryKey: ['provider-models', provider, apiKey, tenantId, upstreamKeyId],
    queryFn: async () => {
      if (!provider) return []
      const { data } = await apiClient.get<string[]>(`/providers/${provider}/models`, {
        params: { 
          apiKey,
          tenantID: tenantId,
          upstreamKeyID: upstreamKeyId
        },
      })
      return data
    },
    enabled: !!provider && (!!apiKey || !!upstreamKeyId || provider === 'openai' || provider === 'anthropic' || provider === 'gemini'),
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

export function useUpdateUpstream(tenantId: string | null) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ keyId, body }: { keyId: string; body: { provider: string; models: string[]; base_url: string; api_key?: string } }) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.put(`/tenants/${tenantId}/upstreams/${keyId}`, body)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['upstreams', tenantId] })
    },
  })
}

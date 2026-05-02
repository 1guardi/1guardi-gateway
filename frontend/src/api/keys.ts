import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface APIKeyResponse {
  ID: number
  CreatedAt: string
  Name: string
  Prefix: string
  Suffix: string
  TenantID: number
  AgentID: number | null
  UserID: number | null
  LastUsedAt: string | null
  IsActive: boolean
}

export interface CreateKeyResponse extends APIKeyResponse {
  key: string
}

export function useAPIKeys(tenantId: string | null) {
  return useQuery({
    queryKey: ['keys', tenantId],
    queryFn: async () => {
      const { data } = await apiClient.get<APIKeyResponse[]>(`/tenants/${tenantId}/keys`)
      return data
    },
    enabled: !!tenantId,
  })
}

export function useCreateAPIKey(tenantId: string | null) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: { name: string; agent_id?: number; user_id?: number }) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      const { data } = await apiClient.post<CreateKeyResponse>(`/tenants/${tenantId}/keys`, body)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keys', tenantId] })
    },
  })
}

export function useDeleteAPIKey(tenantId: string | null) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (keyId: string) => {
      if (!tenantId) throw new Error('Tenant ID is required')
      await apiClient.delete(`/tenants/${tenantId}/keys/${keyId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keys', tenantId] })
    },
  })
}

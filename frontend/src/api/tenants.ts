import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface TenantResponse {
  ID: number
  Name: string
  Description: string
  CreatedAt: string
}

export function useTenants() {
  return useQuery({
    queryKey: ['tenants'],
    queryFn: async () => {
      const { data } = await apiClient.get<TenantResponse[]>('/tenants')
      return data
    },
  })
}

export function useCreateTenant() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: { name: string; description?: string }) => {
      const { data } = await apiClient.post<TenantResponse>('/tenants', req)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
    },
  })
}

// useCreateOrganization is the self-service tenant creation used at onboarding
// (a user belonging to no tenant) and for adding further organizations.
export function useCreateOrganization() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: { name: string; description?: string }) => {
      const { data } = await apiClient.post<TenantResponse>('/onboarding/tenant', req)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries()
    },
  })
}

export function useDeleteTenant() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (tenantId: number) => {
      await apiClient.delete(`/tenants/${tenantId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
    },
  })
}

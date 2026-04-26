import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'

export interface TenantResponse {
  ID: number
  Name: string
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

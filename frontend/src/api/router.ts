import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'

export interface RouteEndpoint {
  id: string
  label: string
  provider: string
  model: string
  tenant_id: number
  ttft_p50_ms: number
  ttft_p99_ms: number
  avg_tps: number
  error_rate: number
  quota_used: number
  circuit_state: string
  score: number
}

export function useRouterEndpoints() {
  return useQuery({
    queryKey: ['router-endpoints'],
    queryFn: async () => {
      const { data } = await apiClient.get<RouteEndpoint[]>('/router/endpoints')
      return data
    },
  })
}

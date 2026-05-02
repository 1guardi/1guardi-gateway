import { useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: string
}

export function useLogin(onSuccess: () => void) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: LoginRequest) => {
      const { data } = await apiClient.post<LoginResponse>('/auth/login', req)
      return data
    },
    onSuccess: (data) => {
      localStorage.setItem('admin_token', data.token)
      queryClient.invalidateQueries()
      onSuccess()
    },
  })
}

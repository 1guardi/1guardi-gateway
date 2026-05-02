import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client.ts'

export interface RoleResponse {
  ID: number
  Name: string
  Description: string
}

export interface UserResponse {
  ID: number
  Name: string
  Email: string
  IsSuperAdmin: boolean
}

export interface TenantMemberResponse {
  ID: number
  UserID: number
  User: UserResponse
  TenantID: number
  RoleID: number
  Role: RoleResponse
  CreatedAt: string
}

export function useMembers(tenantId: string | null) {
  return useQuery({
    queryKey: ['tenants', tenantId, 'members'],
    queryFn: async () => {
      const { data } = await apiClient.get<TenantMemberResponse[]>(`/tenants/${tenantId}/members`)
      return data
    },
    enabled: !!tenantId,
  })
}

export function useAddMember(tenantId: string | null) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (req: { user_id?: number; email?: string; name?: string; password?: string; role_id: number }) => {
      const { data } = await apiClient.post<TenantMemberResponse>(`/tenants/${tenantId}/members`, req)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants', tenantId, 'members'] })
    },
  })
}

export function useRemoveMember(tenantId: string | null) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (userId: string) => {
      await apiClient.delete(`/tenants/${tenantId}/members/${userId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants', tenantId, 'members'] })
    },
  })
}

export function useRoles() {
  return useQuery({
    queryKey: ['roles'],
    queryFn: async () => {
      const { data } = await apiClient.get<RoleResponse[]>('/roles')
      return data
    },
  })
}

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const { data } = await apiClient.get<UserResponse[]>('/users')
      return data
    },
  })
}

import axios from 'axios'
import { getAdminToken, dispatchLogout } from './auth-storage'

const API_URL = import.meta.env.VITE_API_URL || ''

export const apiClient = axios.create({
  baseURL: `${API_URL}/api/v1`,
})

apiClient.interceptors.request.use((config) => {
  const token = getAdminToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      dispatchLogout()
    }
    return Promise.reject(error)
  },
)

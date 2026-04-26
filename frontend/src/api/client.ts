import axios from 'axios'

const API_URL = import.meta.env.VITE_API_URL || ''

export const apiClient = axios.create({
  baseURL: `${API_URL}/api/v1`,
})

apiClient.interceptors.request.use((config) => {
  // TODO: Replace with real auth token logic
  const token = 'mock-auth-token'
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

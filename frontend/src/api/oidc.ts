import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'
import { setAdminToken } from './auth-storage'

const API_URL = import.meta.env.VITE_API_URL || ''

export interface OIDCProvider {
  name: string
  label: string
}

/** useOIDCProviders lists the SSO providers enabled on the gateway. */
export function useOIDCProviders() {
  return useQuery({
    queryKey: ['oidc-providers'],
    queryFn: async () => {
      const { data } = await apiClient.get<OIDCProvider[] | null>('/auth/oidc/providers')
      return data ?? []
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

/** oidcLoginUrl is the full-page navigation target that starts the SSO flow. */
export function oidcLoginUrl(provider: string) {
  return `${API_URL}/api/v1/auth/oidc/${encodeURIComponent(provider)}/login`
}

export interface AuthCallbackResult {
  handled: boolean
  ok: boolean
  error?: string
}

/**
 * consumeAuthCallback handles the OIDC redirect landing at /auth/callback.
 * The gateway returns the JWT (or an error) in the URL fragment. On success
 * it stores the token; either way it scrubs the fragment from the URL so a
 * refresh cannot replay it. Returns handled=false when not on the callback path.
 */
export function consumeAuthCallback(): AuthCallbackResult {
  if (!window.location.pathname.endsWith('/auth/callback')) {
    return { handled: false, ok: false }
  }
  const raw = window.location.hash.startsWith('#')
    ? window.location.hash.slice(1)
    : window.location.hash
  const params = new URLSearchParams(raw)
  const token = params.get('token')
  const error = params.get('error')

  // Scrub the token/error from the address bar before doing anything else.
  window.history.replaceState({}, '', '/')

  if (token) {
    setAdminToken(token)
    return { handled: true, ok: true }
  }
  return { handled: true, ok: false, error: error || 'sso_failed' }
}

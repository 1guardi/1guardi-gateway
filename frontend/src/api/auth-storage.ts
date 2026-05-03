export const ADMIN_TOKEN_KEY = 'admin_token'
export const ADMIN_LOGOUT_EVENT = 'admin-logout'

export function getAdminToken() {
  return localStorage.getItem(ADMIN_TOKEN_KEY)
}

export function setAdminToken(token: string) {
  localStorage.setItem(ADMIN_TOKEN_KEY, token)
}

export function clearAdminToken() {
  localStorage.removeItem(ADMIN_TOKEN_KEY)
}

export function dispatchLogout() {
  clearAdminToken()
  window.dispatchEvent(new Event(ADMIN_LOGOUT_EVENT))
}

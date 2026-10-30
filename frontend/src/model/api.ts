import { sessionCoordinator } from './session'

export type Profile = {
  userID: string
  displayName: string
  firstName: string
  lastName: string
  email: string
}

export type Role = 'owner' | 'editor' | 'viewer'

export type Document = {
  id: string
  ownerUserID: string
  title: string
  content: string
  version: number
  role: Role
  createdAt: string
  updatedAt: string
}

export type Permission = {
  userID: string
  email: string
  displayName: string
  role: Exclude<Role, 'owner'>
  updatedAt: string
}

export type Registration = {
  login: string
  password: string
  displayName: string
  firstName: string
  lastName: string
  email: string
  rememberMe: boolean
}

export class ApiError extends Error {
  constructor(public status: number, message: string, public fields?: Record<string, string>) {
    super(message)
  }
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: 'include',
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({ message: 'Ошибка запроса' })) as { message?: string; fields?: Record<string, string> }
    if (response.status === 401) sessionCoordinator.publish('logged-out')
    const fieldMessage = body.fields ? Object.entries(body.fields).map(([field, reason]) => `${field}: ${reason}`).join('; ') : ''
    throw new ApiError(response.status, fieldMessage || body.message || 'Ошибка запроса', body.fields)
  }
  return response.status === 204 ? undefined as T : response.json() as Promise<T>
}

export const api = {
  register: (data: Registration) => request<{ authenticated: boolean }>('/api/public/v1/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (login: string, password: string, rememberMe: boolean) => request('/api/public/v1/auth/login', { method: 'POST', body: JSON.stringify({ login, password, rememberMe }) }),
  logout: () => request('/api/public/v1/auth/logout', { method: 'POST' }),
  changePassword: (currentPassword: string, newPassword: string) => request('/api/public/v1/auth/change-password', { method: 'POST', body: JSON.stringify({ currentPassword, newPassword }) }),
  me: () => request<Profile>('/api/bff/v1/me'),
  updateMe: (profile: Omit<Profile, 'userID'>) => request<Profile>('/api/bff/v1/me', { method: 'PUT', body: JSON.stringify(profile) }),
  documents: () => request<Document[]>('/api/editor/v1/documents'),
  document: (id: string) => request<Document>(`/api/editor/v1/documents/${id}`),
  createDocument: (title = 'Без названия') => request<Document>('/api/editor/v1/documents', { method: 'POST', body: JSON.stringify({ title, content: '' }) }),
  updateDocument: (document: Pick<Document, 'id' | 'title' | 'content' | 'version'>) => request<Document>(`/api/editor/v1/documents/${document.id}`, { method: 'PUT', body: JSON.stringify({ title: document.title, content: document.content, version: document.version }) }),
  deleteDocument: (id: string) => request(`/api/editor/v1/documents/${id}`, { method: 'DELETE' }),
  permissions: (id: string) => request<Permission[]>(`/api/editor/v1/documents/${id}/permissions`),
  grant: (id: string, email: string, role: 'viewer' | 'editor') => request(`/api/editor/v1/documents/${id}/permissions`, { method: 'PUT', body: JSON.stringify({ email, role }) }),
  revoke: (id: string, userID: string) => request(`/api/editor/v1/documents/${id}/permissions/${userID}`, { method: 'DELETE' }),
}

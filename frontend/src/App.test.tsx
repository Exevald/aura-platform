import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import App from './App'
import { afterEach, expect, test, vi } from 'vitest'

const profile = { userID: '7c8e21f3-93f4-43f5-9b42-6171247d495c', displayName: 'Ada', firstName: 'Ada', lastName: 'Lovelace', email: 'ada@example.com' }

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

test('renders login screen when there is no active session', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({ message: 'authentication required' }, 401)))
  render(<MemoryRouter initialEntries={['/login']}><App/></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Войти' })).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: 'Нет аккаунта? Зарегистрироваться' }))
  expect(screen.getByLabelText('Пароль')).toHaveAttribute('minlength', '8')
})

test('restores an existing session instead of showing login', async () => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input)
    if (url === '/api/bff/v1/me') return Promise.resolve(jsonResponse(profile))
    if (['/api/usergroups/v1/groups','/api/access/v1/modules','/api/access/v1/group-module-accesses'].includes(url)) return Promise.resolve(jsonResponse([]))
    return Promise.resolve(jsonResponse({ message: 'not found' }, 404))
  }))
  render(<MemoryRouter initialEntries={['/login']}><App/></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Добро пожаловать, Ada' })).toBeInTheDocument()
})

test('detects a cross-subdomain login when the login tab receives focus', async () => {
  let authenticated = false
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input)
    if (url === '/api/bff/v1/me') {
      return Promise.resolve(authenticated ? jsonResponse(profile) : jsonResponse({ message: 'authentication required' }, 401))
    }
    if (['/api/usergroups/v1/groups','/api/access/v1/modules','/api/access/v1/group-module-accesses'].includes(url)) return Promise.resolve(jsonResponse([]))
    return Promise.resolve(jsonResponse({ message: 'not found' }, 404))
  }))
  render(<MemoryRouter initialEntries={['/login']}><App/></MemoryRouter>)
  await screen.findByRole('heading', { name: 'Войти' })

  authenticated = true
  window.dispatchEvent(new Event('focus'))

  expect(await screen.findByRole('heading', { name: 'Добро пожаловать, Ada' })).toBeInTheDocument()
})

test('revalidates the session when the product tab receives focus', async () => {
  const fetchMock = vi.fn((input: RequestInfo | URL) => {
    const url = String(input)
    if (url === '/api/bff/v1/me') return Promise.resolve(jsonResponse(profile))
    if (url === '/api/editor/v1/documents') return Promise.resolve(jsonResponse([]))
    return Promise.resolve(jsonResponse({ message: 'not found' }, 404))
  })
  vi.stubGlobal('fetch', fetchMock)
  render(<MemoryRouter initialEntries={['/documents']}><App/></MemoryRouter>)
  await screen.findByRole('heading', { name: 'Документы' })
  const beforeFocus = fetchMock.mock.calls.filter(([url]) => String(url) === '/api/bff/v1/me').length

  window.dispatchEvent(new Event('focus'))

  await waitFor(() => expect(fetchMock.mock.calls.filter(([url]) => String(url) === '/api/bff/v1/me')).toHaveLength(beforeFocus + 1))
})

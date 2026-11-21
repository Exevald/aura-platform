import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react'
import {MemoryRouter} from 'react-router-dom'
import App from './App'
import {afterEach, beforeEach, expect, test, vi} from 'vitest'

const profile = {
    userID: '7c8e21f3-93f4-43f5-9b42-6171247d495c',
    displayName: 'Ada',
    firstName: 'Ada',
    lastName: 'Lovelace',
    email: 'ada@example.com'
}
const workspace = {id: profile.userID, kind: 'personal', name: 'Ada', role: 'owner', members: []}

function jsonResponse(body: unknown, status = 200) {
    return new Response(JSON.stringify(body), {status, headers: {'Content-Type': 'application/json'}})
}

beforeEach(() => {
    const values = new Map<string, string>()
    vi.stubGlobal('localStorage', {
        getItem: (key: string) => values.get(key) ?? null,
        setItem: (key: string, value: string) => {
            values.set(key, value)
        },
        removeItem: (key: string) => {
            values.delete(key)
        },
    })
})

afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
})

test('renders login screen when there is no active session', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({message: 'authentication required'}, 401)))
    render(<MemoryRouter initialEntries={['/login']}><App/></MemoryRouter>)
    expect(await screen.findByRole('heading', {name: 'Войти'})).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', {name: 'Нет аккаунта? Зарегистрироваться'}))
    expect(screen.getByLabelText('Пароль')).toHaveAttribute('minlength', '8')
})

test('redirects only once when a product session expires', async () => {
    const assign = vi.fn()
    vi.stubGlobal('location', {
        hostname: 'aura-platform.documents',
        protocol: 'http:',
        port: '',
        href: 'http://aura-platform.documents/documents',
        pathname: '/documents',
        assign,
    })
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({message: 'authentication required'}, 401)))

    render(<MemoryRouter initialEntries={['/documents']}><App/></MemoryRouter>)

    await waitFor(() => expect(assign).toHaveBeenCalledTimes(1))
    expect(assign.mock.calls[0][0]).toContain('http://aura-platform.auth/login?return=')
})

test('restores an existing session instead of showing login', async () => {
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url === '/api/bff/v1/me') return Promise.resolve(jsonResponse(profile))
        if (url === '/api/workspaces/v1/workspaces') return Promise.resolve(jsonResponse([workspace]))
        if (['/api/usergroups/v1/groups', '/api/access/v1/modules', '/api/access/v1/group-module-accesses'].includes(url)) return Promise.resolve(jsonResponse([]))
        return Promise.resolve(jsonResponse({message: 'not found'}, 404))
    }))
    render(<MemoryRouter initialEntries={['/login']}><App/></MemoryRouter>)
    expect(await screen.findByRole('heading', {name: 'Добро пожаловать, Ada'})).toBeInTheDocument()
})

test('detects a cross-subdomain login when the login tab receives focus', async () => {
    let authenticated = false
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url === '/api/bff/v1/me') {
            return Promise.resolve(authenticated ? jsonResponse(profile) : jsonResponse({message: 'authentication required'}, 401))
        }
        if (url === '/api/workspaces/v1/workspaces') return Promise.resolve(jsonResponse([workspace]))
        if (['/api/usergroups/v1/groups', '/api/access/v1/modules', '/api/access/v1/group-module-accesses'].includes(url)) return Promise.resolve(jsonResponse([]))
        return Promise.resolve(jsonResponse({message: 'not found'}, 404))
    }))
    render(<MemoryRouter initialEntries={['/login']}><App/></MemoryRouter>)
    await screen.findByRole('heading', {name: 'Войти'})

    authenticated = true
    window.dispatchEvent(new Event('focus'))

    expect(await screen.findByRole('heading', {name: 'Добро пожаловать, Ada'})).toBeInTheDocument()
})

test('revalidates the session when the product tab receives focus', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url === '/api/bff/v1/me') return Promise.resolve(jsonResponse(profile))
        if (url === '/api/workspaces/v1/workspaces') return Promise.resolve(jsonResponse([workspace]))
        if (url === '/api/editor/v1/documents') return Promise.resolve(jsonResponse([]))
        return Promise.resolve(jsonResponse({message: 'not found'}, 404))
    })
    vi.stubGlobal('fetch', fetchMock)
    render(<MemoryRouter initialEntries={['/documents']}><App/></MemoryRouter>)
    await screen.findByRole('heading', {name: 'Документы'})
    const beforeFocus = fetchMock.mock.calls.filter(([url]) => String(url) === '/api/bff/v1/me').length

    window.dispatchEvent(new Event('focus'))

    await waitFor(() => expect(fetchMock.mock.calls.filter(([url]) => String(url) === '/api/bff/v1/me')).toHaveLength(beforeFocus + 1))
})

test('replaces another account workspace before requesting documents', async () => {
    window.localStorage.setItem('aura-workspace-id', '11111111-1111-4111-8111-111111111111')
    let releaseWorkspaces!: (response: Response) => void
    const workspacesResponse = new Promise<Response>(resolve => {
        releaseWorkspaces = resolve
    })
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input)
        if (url === '/api/bff/v1/me') return Promise.resolve(jsonResponse(profile))
        if (url === '/api/workspaces/v1/workspaces') return workspacesResponse
        if (url === '/api/editor/v1/documents') {
            const headers = init?.headers as Record<string, string>
            return Promise.resolve(headers['X-Workspace-ID'] === workspace.id
                ? jsonResponse([])
                : jsonResponse({message: 'workspace access denied'}, 403))
        }
        return Promise.resolve(jsonResponse({message: 'not found'}, 404))
    })
    vi.stubGlobal('fetch', fetchMock)

    render(<MemoryRouter initialEntries={['/documents']}><App/></MemoryRouter>)
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith('/api/workspaces/v1/workspaces', expect.anything()))
    expect(fetchMock.mock.calls.some(([url]) => String(url) === '/api/editor/v1/documents')).toBe(false)

    releaseWorkspaces(jsonResponse([workspace]))
    expect(await screen.findByRole('heading', {name: 'Документы'})).toBeInTheDocument()
    await waitFor(() => expect(fetchMock.mock.calls.some(([url]) => String(url) === '/api/editor/v1/documents')).toBe(true))
    expect(window.localStorage.getItem('aura-workspace-id')).toBe(workspace.id)
    expect(screen.queryByText('workspace access denied')).not.toBeInTheDocument()
})

test('clears the selected workspace on logout', async () => {
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url === '/api/bff/v1/me') return Promise.resolve(jsonResponse(profile))
        if (url === '/api/workspaces/v1/workspaces') return Promise.resolve(jsonResponse([workspace]))
        if (url === '/api/public/v1/auth/logout') return Promise.resolve(new Response(null, {status: 204}))
        return Promise.resolve(jsonResponse({message: 'not found'}, 404))
    }))

    render(<MemoryRouter initialEntries={['/overview']}><App/></MemoryRouter>)
    await screen.findByRole('heading', {name: 'Добро пожаловать, Ada'})
    expect(window.localStorage.getItem('aura-workspace-id')).toBe(workspace.id)

    fireEvent.click(screen.getByRole('button', {name: 'Выйти'}))
    await waitFor(() => expect(window.localStorage.getItem('aura-workspace-id')).toBeNull())
})

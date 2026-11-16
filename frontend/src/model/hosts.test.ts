import {expect, test} from 'vitest'
import {productForPath, safeReturnURL} from './hosts'

test('maps product paths to their local domains', () => {
    expect(productForPath('/documents/123')).toBe('documents')
    expect(productForPath('/calendar')).toBe('calendar')
    expect(productForPath('/planner')).toBe('planner')
    expect(productForPath('/profile')).toBe('home')
})

test('SSO return URLs stay on the expected product hosts and port', () => {
    const port = location.port ? `:${location.port}` : ''
    expect(safeReturnURL(`http://aura-platform.documents${port}/documents/123`).hostname).toBe('aura-platform.documents')
    expect(safeReturnURL('http://evil.example/documents').hostname).toBe('aura-platform.home')
    expect(safeReturnURL('http://aura-platform.documents:9999/documents').hostname).toBe('aura-platform.home')
})

import {randomHex} from './random'

export const productPaths: Record<string, string> = {
    home: '/overview',
    documents: '/documents',
    calendar: '/calendar',
    planner: '/planner',
    shortlinks: '/shortlinks',
    publications: '/publications',
    notifications: '/notifications',
}

export function productHost(product: string): string {
    return `aura-platform.${product}`
}

export function productURL(product: string, path = productPaths[product]): string {
    return `${location.protocol}//${productHost(product)}${location.port ? `:${location.port}` : ''}${path}`
}

export function productForPath(path: string): string {
    if (path.startsWith('/documents')) return 'documents'
    if (path.startsWith('/calendar')) return 'calendar'
    if (path.startsWith('/planner')) return 'planner'
    if (path.startsWith('/shortlinks') || path.startsWith('/s/')) return 'shortlinks'
    if (path.startsWith('/publications')) return 'publications'
    if (path.startsWith('/notifications')) return 'notifications'
    return 'home'
}

export function isAuraHost(): boolean {
    return /^aura-platform\.(home|auth|documents|calendar|planner|shortlinks|publications|notifications)$/.test(location.hostname)
}

export function loginURL(returnTo: string): string {
    const target = new URL(returnTo)
    const state = randomHex(16)
    sessionStorage.setItem('aura-sso-state', state)
    target.searchParams.set('sso_state', state)
    return productURL('auth', `/login?return=${encodeURIComponent(target.toString())}`)
}

export function navigationURL(path: string): string {
    if (!isAuraHost()) return path
    return productURL(productForPath(path), path)
}

export function safeReturnURL(raw: string | null): URL {
    try {
        const url = new URL(raw ?? '', location.href)
        if (url.protocol === location.protocol && url.port === location.port && !url.username && !url.password &&
            Object.keys(productPaths).some(name => url.hostname === productHost(name))) {
            return url
        }
    } catch { /* invalid return URL */
    }
    return new URL(productURL('home'))
}

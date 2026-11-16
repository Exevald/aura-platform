import {useEffect, useRef, useState} from 'react'
import {StateView} from '../views/StateView'
import {loginURL, productPaths} from '../model/hosts'

export function SSOCallbackPresenter() {
    const started = useRef(false)
    const [failed, setFailed] = useState(() => {
        const params = new URLSearchParams(location.search)
        const next = params.get('next') ?? '/overview'
        const target = new URL(next, location.origin)
        return !params.get('code') || target.origin !== location.origin ||
            !target.searchParams.get('sso_state') || target.searchParams.get('sso_state') !== sessionStorage.getItem('aura-sso-state')
    })
    useEffect(() => {
        if (started.current) return
        started.current = true
        const params = new URLSearchParams(location.search)
        const code = params.get('code')
        const next = params.get('next') ?? '/overview'
        history.replaceState(null, '', '/sso/callback')
        const target = new URL(next, location.origin)
        if (!code || target.origin !== location.origin ||
            !target.searchParams.get('sso_state') || target.searchParams.get('sso_state') !== sessionStorage.getItem('aura-sso-state')) {
            return
        }
        sessionStorage.removeItem('aura-sso-state')
        target.searchParams.delete('sso_state')
        fetch('/api/public/v1/auth/sso/consume', {
            method: 'POST', credentials: 'include',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({code}),
        }).then(response => {
            if (!response.ok) throw new Error('SSO exchange failed')
            location.replace(target.pathname + target.search + target.hash)
        }).catch(() => setFailed(true))
    }, [])
    return <StateView title={failed ? 'Не удалось войти' : 'Выполняем вход…'}
                      action={failed ? 'Повторить вход' : undefined}
                      onAction={failed ? () => location.replace(loginURL(location.origin + (productPaths[location.hostname.split('.')[1]] ?? '/overview'))) : undefined}/>
}

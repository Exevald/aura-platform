import {useCallback, useEffect, useRef, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import {api, ApiError} from '../model/api'
import {sessionCoordinator} from '../model/session'
import {type AuthSubmission, AuthView} from '../views/AuthView'
import {StateView} from '../views/StateView'
import {errorMessage} from './error'
import {isAuraHost, loginURL, safeReturnURL} from '../model/hosts'

export function AuthPresenter() {
    const navigate = useNavigate()
    const [registering, setRegistering] = useState(false)
    const [error, setError] = useState('')
    const [busy, setBusy] = useState(false)
    const [checkingSession, setCheckingSession] = useState(true)
    const redirecting = useRef(false)

    const finishSignIn = useCallback(async () => {
        if (location.hostname !== 'aura-platform.auth') {
            navigate('/overview', {replace: true})
            return
        }
        if (redirecting.current) return
        redirecting.current = true
        const returnTo = new URLSearchParams(location.search).get('return')
        if (!returnTo) {
            location.replace(`${location.protocol}//aura-platform.home${location.port ? `:${location.port}` : ''}/`)
            return
        }
        const target = safeReturnURL(returnTo)
        if (!target.searchParams.has('sso_state')) {
            location.replace(`${location.protocol}//aura-platform.home${location.port ? `:${location.port}` : ''}/`)
            return
        }
        try {
            const response = await fetch('/api/public/v1/auth/sso/ticket', {
                method: 'POST', credentials: 'include',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({targetHost: target.hostname}),
            })
            if (!response.ok) throw new Error('Не удалось открыть продукт')
            const {code} = await response.json() as {code: string}
            const next = target.pathname + target.search + target.hash
            target.pathname = '/sso/callback'
            target.search = new URLSearchParams({code, next}).toString()
            location.replace(target.toString())
        } catch (reason) {
            redirecting.current = false
            throw reason
        }
    }, [navigate])

    useEffect(() => {
        if (isAuraHost() && location.hostname !== 'aura-platform.auth') {
            location.replace(loginURL(location.href))
            return
        }
        let active = true
        const unsubscribe = sessionCoordinator.subscribe((event) => {
            if (event === 'authenticated') void finishSignIn().catch((reason: unknown) => setError(errorMessage(reason)))
        })
        api.me().then(() => finishSignIn()).catch((requestError) => {
            if (!active) return
            if (!(requestError instanceof ApiError && requestError.status === 401)) setError(errorMessage(requestError))
            setCheckingSession(false)
        })
        const revalidate = () => {
            api.me().then(() => {
                if (active) void finishSignIn().catch((reason: unknown) => setError(errorMessage(reason)))
            }).catch((requestError) => {
                if (!active) return
                if (!(requestError instanceof ApiError && requestError.status === 401)) setError(errorMessage(requestError))
            })
        }
        const revalidateVisible = () => {
            if (document.visibilityState === 'visible') revalidate()
        }
        window.addEventListener('focus', revalidate)
        document.addEventListener('visibilitychange', revalidateVisible)
        return () => {
            active = false
            unsubscribe()
            window.removeEventListener('focus', revalidate)
            document.removeEventListener('visibilitychange', revalidateVisible)
        }
    }, [finishSignIn])

    async function submit(values: AuthSubmission) {
        setBusy(true)
        setError('')
        try {
            if (registering) await api.register(values)
            else await api.login(values.login, values.password, values.rememberMe)
            sessionCoordinator.publish('authenticated')
        } catch (requestError) {
            setError(errorMessage(requestError))
        } finally {
            setBusy(false)
        }
    }

    function toggleMode() {
        setRegistering((value) => !value)
        setError('')
    }

    if (checkingSession) return <StateView title="Проверяем сессию…"/>
    return <AuthView registering={registering} busy={busy} error={error} onSubmit={submit} onToggleMode={toggleMode}/>
}

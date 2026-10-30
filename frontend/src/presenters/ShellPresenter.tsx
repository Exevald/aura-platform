import {useCallback, useEffect, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import {api, ApiError, type Profile} from '../model/api'
import {sessionCoordinator} from '../model/session'
import {ShellView} from '../views/ShellView'
import {StateView} from '../views/StateView'

export function ShellPresenter() {
    const navigate = useNavigate()
    const [profile, setProfile] = useState<Profile | null>(null)
    const [failed, setFailed] = useState(false)

    const loadProfile = useCallback(async () => {
        try {
            setProfile(await api.me())
            setFailed(false)
        } catch (error) {
            if (error instanceof ApiError && error.status === 401) navigate('/login')
            else setFailed(true)
        }
    }, [navigate])

    useEffect(() => {
        api.me().then((currentProfile) => {
            setProfile(currentProfile)
            setFailed(false)
        }).catch((error) => {
            if (error instanceof ApiError && error.status === 401) navigate('/login')
            else setFailed(true)
        })
    }, [navigate])

    useEffect(() => {
        const unsubscribe = sessionCoordinator.subscribe((event) => {
            if (event === 'logged-out') navigate('/login', {replace: true})
            else void loadProfile()
        })
        const revalidate = () => void loadProfile()
        const revalidateVisible = () => {
            if (document.visibilityState === 'visible') void loadProfile()
        }
        window.addEventListener('focus', revalidate)
        document.addEventListener('visibilitychange', revalidateVisible)
        return () => {
            unsubscribe()
            window.removeEventListener('focus', revalidate)
            document.removeEventListener('visibilitychange', revalidateVisible)
        }
    }, [loadProfile, navigate])

    async function logout() {
        try {
            await api.logout()
            sessionCoordinator.publish('logged-out')
            navigate('/login')
        } catch {
            setFailed(true)
        }
    }

    if (failed) return <StateView title="Сервис временно недоступен" action="Обновить"
                                  onAction={() => location.reload()}/>
    if (!profile) return <StateView title="Загружаем Aura…"/>
    return <ShellView profile={profile} setProfile={setProfile} onLogout={logout}/>
}

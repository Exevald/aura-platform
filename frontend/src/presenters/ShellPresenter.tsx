import {useCallback, useEffect, useRef, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import {api, ApiError, type Profile, type Workspace} from '../model/api'
import {sessionCoordinator} from '../model/session'
import {ShellView} from '../views/ShellView'
import {StateView} from '../views/StateView'
import {isAuraHost, loginURL} from '../model/hosts'

export function ShellPresenter() {
    const navigate = useNavigate()
    const [profile, setProfile] = useState<Profile | null>(null)
    const [failed, setFailed] = useState(false)
    const [workspaces, setWorkspaces] = useState<Workspace[]>([])
    const [workspaceID, setWorkspaceID] = useState('')
    const redirecting = useRef(false)

    const redirectToLogin = useCallback(() => {
        if (redirecting.current) return
        redirecting.current = true
        if (isAuraHost() && location.hostname !== 'aura-platform.auth') location.assign(loginURL(location.href))
        else navigate('/login', {replace: true})
    }, [navigate])

    const loadSession = useCallback(async () => {
        try {
            const currentProfile = await api.me()
            const items = await api.workspaces()
            const savedID = window.localStorage.getItem('aura-workspace-id')
            const selectedID = items.find(item => item.id === savedID)?.id ?? items[0]?.id
            if (!selectedID) {
                setFailed(true)
                return
            }
            window.localStorage.setItem('aura-workspace-id', selectedID)
            setWorkspaceID(selectedID)
            setWorkspaces(items)
            setProfile(currentProfile)
            setFailed(false)
        } catch (error) {
            if (error instanceof ApiError && error.status === 401) redirectToLogin()
            else setFailed(true)
        }
    }, [redirectToLogin])

    useEffect(() => {
        void Promise.resolve().then(loadSession)
    }, [loadSession])

    useEffect(() => {
        const unsubscribe = sessionCoordinator.subscribe((event) => {
            if (event === 'logged-out') {
                window.localStorage.removeItem('aura-workspace-id')
                redirectToLogin()
            } else {
                setProfile(null)
                void loadSession()
            }
        })
        const revalidate = () => void loadSession()
        const revalidateVisible = () => {
            if (document.visibilityState === 'visible') void loadSession()
        }
        window.addEventListener('focus', revalidate)
        document.addEventListener('visibilitychange', revalidateVisible)
        return () => {
            unsubscribe()
            window.removeEventListener('focus', revalidate)
            document.removeEventListener('visibilitychange', revalidateVisible)
        }
    }, [loadSession, redirectToLogin])

    async function logout() {
        try {
            await api.logout()
            sessionCoordinator.publish('logged-out')
        } catch {
            setFailed(true)
        }
    }

    if (failed) return <StateView title="Сервис временно недоступен" action="Обновить"
                                  onAction={() => location.reload()}/>
    if (!profile) return <StateView title="Загружаем Aura…"/>

    function selectWorkspace(id: string) {
        window.localStorage.setItem('aura-workspace-id', id)
        setWorkspaceID(id)
        location.reload()
    }

    return <ShellView profile={profile} setProfile={setProfile} onLogout={logout} workspaces={workspaces}
                      workspaceID={workspaceID} onSelectWorkspace={selectWorkspace}/>
}

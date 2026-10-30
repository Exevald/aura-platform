import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ApiError, api } from '../model/api'
import { sessionCoordinator } from '../model/session'
import { AuthView, type AuthSubmission } from '../views/AuthView'
import { StateView } from '../views/StateView'
import { errorMessage } from './error'

export function AuthPresenter() {
  const navigate = useNavigate()
  const [registering, setRegistering] = useState(false)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const [checkingSession, setCheckingSession] = useState(true)

  useEffect(() => {
    let active = true
    const unsubscribe = sessionCoordinator.subscribe((event) => {
      if (event === 'authenticated') navigate('/documents', { replace: true })
    })
    api.me().then(() => navigate('/documents', { replace: true })).catch((requestError) => {
      if (!active) return
      if (!(requestError instanceof ApiError && requestError.status === 401)) setError(errorMessage(requestError))
      setCheckingSession(false)
    })
    const revalidate = () => {
      api.me().then(() => {
        if (active) navigate('/documents', { replace: true })
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
  }, [navigate])

  async function submit(values: AuthSubmission) {
    setBusy(true)
    setError('')
    try {
      if (registering) await api.register(values)
      else await api.login(values.login, values.password, values.rememberMe)
      sessionCoordinator.publish('authenticated')
      navigate('/documents')
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

import {useState} from 'react'
import {useNavigate, useOutletContext} from 'react-router-dom'
import {api, type Profile} from '../model/api'
import {sessionCoordinator} from '../model/session'
import {ProfileView} from '../views/ProfileView'
import type {ShellContext} from '../views/ShellView'
import {errorMessage} from './error'

export function ProfilePresenter() {
    const {profile, setProfile} = useOutletContext<ShellContext>()
    const navigate = useNavigate()
    const [notice, setNotice] = useState('')
    const [error, setError] = useState('')

    async function save(values: Omit<Profile, 'userID'>) {
        try {
            const updated = await api.updateMe(values)
            setProfile(updated)
            setNotice('Профиль сохранён')
            setError('')
        } catch (requestError) {
            setError(errorMessage(requestError))
        }
    }

    async function changePassword(currentPassword: string, newPassword: string) {
        try {
            await api.changePassword(currentPassword, newPassword)
            sessionCoordinator.publish('logged-out')
            navigate('/login')
        } catch (requestError) {
            setError(errorMessage(requestError))
        }
    }

    return <ProfileView profile={profile} notice={notice} error={error} onSave={save}
                        onChangePassword={changePassword}/>
}

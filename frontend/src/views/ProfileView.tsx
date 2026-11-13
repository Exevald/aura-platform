import type {FormEvent} from 'react'
import {KeyRound, Save, UserRound} from 'lucide-react'
import type {Profile} from '../model/api'

type Props = {
    profile: Profile
    notice: string
    error: string
    onSave: (profile: Omit<Profile, 'userID'>) => void
    onChangePassword: (currentPassword: string, newPassword: string) => void
}

export function ProfileView({profile, notice, error, onSave, onChangePassword}: Props) {
    function save(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const data = new FormData(event.currentTarget)
        onSave({
            displayName: String(data.get('displayName')),
            firstName: String(data.get('firstName')),
            lastName: String(data.get('lastName')),
            email: String(data.get('email'))
        })
    }

    function password(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const data = new FormData(event.currentTarget)
        onChangePassword(String(data.get('currentPassword')), String(data.get('newPassword')))
    }

    return <>
        <div className="page-title">
            <div><h1>Личные данные</h1><p>Управляйте профилем и безопасностью аккаунта</p></div>
        </div>
        <div className="settings-grid">
            <form className="settings-card form-stack" onSubmit={save}>
                <div className="card-heading"><span className="card-heading-icon"><UserRound size={20}/></span>
                    <div><h2>Профиль</h2><p>Основная информация об аккаунте</p></div>
                </div>
                <label>Отображаемое имя<input
                    name="displayName" defaultValue={profile.displayName} required/></label>
                <div className="form-row"><label>Имя<input name="firstName" defaultValue={profile.firstName} required/></label><label>Фамилия<input
                    name="lastName" defaultValue={profile.lastName} required/></label></div>
                <label>Email<input name="email" type="email" defaultValue={profile.email} required/></label>
                <button className="primary"><Save size={17}/> Сохранить</button>
                {notice && <div className="success">{notice}</div>}</form>
            <form className="settings-card form-stack" onSubmit={password}>
                <div className="card-heading"><span className="card-heading-icon violet"><KeyRound size={20}/></span>
                    <div><h2>Смена пароля</h2><p>Безопасность входа в Aura</p></div>
                </div>
                <label>Текущий
                    пароль<input name="currentPassword" type="password" required/></label><label>Новый пароль<input
                name="newPassword" type="password" minLength={8} required/><small>Не менее 8 символов</small></label>
                <button className="secondary">Изменить пароль</button>
                {error && <div className="error">{error}</div>}</form>
        </div>
    </>
}

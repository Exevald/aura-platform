import type {FormEvent} from 'react'
import type {Registration} from '../model/api'

export type AuthSubmission = Registration

type Props = {
    registering: boolean
    busy: boolean
    error: string
    onSubmit: (values: AuthSubmission) => void
    onToggleMode: () => void
}

export function AuthView({registering, busy, error, onSubmit, onToggleMode}: Props) {
    function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        const data = new FormData(event.currentTarget)
        onSubmit({
            login: String(data.get('login')),
            password: String(data.get('password')),
            displayName: String(data.get('displayName') ?? ''),
            firstName: String(data.get('firstName') ?? ''),
            lastName: String(data.get('lastName') ?? ''),
            email: String(data.get('email') ?? ''),
            rememberMe: data.get('rememberMe') === 'on',
        })
    }

    return <main className="auth-page">
        <section className="auth-card">
            <div className="brand big"><span>A</span>Aura Docs</div>
            <h1>{registering ? 'Создать аккаунт' : 'Войти'}</h1><p className="muted">Документы и заметки в одном
            спокойном рабочем пространстве.</p>
            <form onSubmit={submit} className="form-stack">
                {registering && <><label>Отображаемое имя<input name="displayName" required maxLength={255}/></label>
                    <div className="form-row"><label>Имя<input name="firstName" required/></label><label>Фамилия<input
                        name="lastName" required/></label></div>
                    <label>Email<input name="email" type="email" required/></label></>}
                <label>Логин<input name="login" autoComplete="username" required/></label><label>Пароль<input
                name="password" type="password" autoComplete={registering ? 'new-password' : 'current-password'}
                minLength={registering ? 8 : 1} required/></label>
                <label className="check"><input name="rememberMe" type="checkbox"/> Оставаться в
                    системе</label>{error && <div className="error">{error}</div>}
                <button className="primary"
                        disabled={busy}>{busy ? 'Подождите…' : registering ? 'Зарегистрироваться' : 'Войти'}</button>
            </form>
            <button className="link-button"
                    onClick={onToggleMode}>{registering ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Зарегистрироваться'}</button>
        </section>
    </main>
}

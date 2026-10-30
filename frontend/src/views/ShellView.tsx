import {Link, Outlet} from 'react-router-dom'
import type {Profile} from '../model/api'

export type ShellContext = { profile: Profile; setProfile: (profile: Profile) => void }

type Props = ShellContext & { onLogout: () => void }

export function ShellView({profile, setProfile, onLogout}: Props) {
    return <div className="shell">
        <header><Link className="brand" to="/documents"><span>A</span>Aura Docs</Link>
            <div className="search">⌕ <span>Поиск в документах</span></div>
            <button className="avatar" title="Выйти"
                    onClick={onLogout}>{profile.displayName.slice(0, 1).toUpperCase()}</button>
        </header>
        <aside><Link className="new-button" to="/documents/new">＋ Новый документ</Link>
            <nav><Link to="/documents">▤ Документы</Link><Link to="/profile">♙ Профиль</Link></nav>
            <div className="aside-user"><strong>{profile.displayName}</strong><small>{profile.email}</small></div>
        </aside>
        <section className="workspace"><Outlet context={{profile, setProfile} satisfies ShellContext}/></section>
    </div>
}

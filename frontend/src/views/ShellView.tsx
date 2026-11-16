import {useEffect, useState} from 'react'
import {
    Bell,
    CalendarDays,
    CheckSquare,
    FileText,
    LayoutDashboard,
    Link2,
    LogOut,
    Menu,
    Moon,
    Sun,
    UserRound,
    Users,
    X
} from 'lucide-react'
import {NavLink, Outlet} from 'react-router-dom'
import type {Profile, Workspace} from '../model/api'
import {isAuraHost, navigationURL} from '../model/hosts'

export type ShellContext = { profile: Profile; setProfile: (profile: Profile) => void }

type Props = ShellContext & {
    onLogout: () => void;
    workspaces?: Workspace[];
    workspaceID?: string;
    onSelectWorkspace?: (id: string) => void
}

export function ShellView({
                              profile,
                              setProfile,
                              onLogout,
                              workspaces = [],
                              workspaceID = '',
                              onSelectWorkspace
                          }: Props) {
    const [menuOpen, setMenuOpen] = useState(false)
    const [theme, setTheme] = useState<'light' | 'dark'>(() => {
        const storage = window.localStorage
        const saved = typeof storage?.getItem === 'function' ? storage.getItem('aura-theme') : null
        if (saved === 'light' || saved === 'dark') return saved
        return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    })

    useEffect(() => {
        document.documentElement.dataset.theme = theme
        const storage = window.localStorage
        if (typeof storage?.setItem === 'function') storage.setItem('aura-theme', theme)
    }, [theme])

    const navItems = [
        {to: '/overview', label: 'Обзор', icon: LayoutDashboard},
        {to: '/documents', label: 'Документы', icon: FileText},
        {to: '/shortlinks', label: 'Shortlink', icon: Link2},
        {to: '/calendar', label: 'Календарь', icon: CalendarDays},
        {to: '/planner', label: 'Задачи', icon: CheckSquare},
        {to: '/notifications', label: 'Уведомления', icon: Bell},
        {to: '/groups', label: 'Группы', icon: Users},
        {to: '/profile', label: 'Профиль', icon: UserRound},
    ]

    return <div className="shell">
        <header className="mobile-header">
            <a className="brand" href={navigationURL('/overview')}><span>A</span><span>Aura</span></a>
            <button className="icon-button" aria-label={menuOpen ? 'Закрыть меню' : 'Открыть меню'}
                    onClick={() => setMenuOpen(value => !value)}>
                {menuOpen ? <X size={20}/> : <Menu size={20}/>}
            </button>
        </header>
        {menuOpen && <button className="nav-backdrop" aria-label="Закрыть меню" onClick={() => setMenuOpen(false)}/>}
        <aside className={menuOpen ? 'open' : ''}>
            <a className="brand sidebar-brand" href={navigationURL('/overview')}
                  onClick={() => setMenuOpen(false)}><span>A</span><span>Aura</span></a>
            {workspaces.length > 0 && <select aria-label="Рабочее пространство" value={workspaceID}
                                              onChange={event => onSelectWorkspace?.(event.target.value)}>
                {workspaces.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
            </select>}
            <nav aria-label="Основная навигация">
                {navItems.map(({to, label, icon: Icon}) => isAuraHost()
                    ? <a key={to} href={navigationURL(to)} className={location.pathname === to || location.pathname.startsWith(to + '/') ? 'active' : undefined}
                         onClick={() => setMenuOpen(false)}>
                        <Icon size={19}/><span>{label}</span>
                    </a>
                    : <NavLink key={to} to={to} onClick={() => setMenuOpen(false)}>
                        <Icon size={19}/><span>{label}</span>
                    </NavLink>)}
            </nav>
            <div className="sidebar-footer">
                <button className="theme-toggle"
                        onClick={() => setTheme(value => value === 'light' ? 'dark' : 'light')}>
                    {theme === 'light' ? <Moon size={18}/> :
                        <Sun size={18}/>}<span>{theme === 'light' ? 'Тёмная тема' : 'Светлая тема'}</span>
                </button>
                <div className="aside-user">
                    <a className="avatar" href={navigationURL('/profile')}>{profile.displayName.slice(0, 1).toUpperCase()}</a>
                    <div><strong>{profile.displayName}</strong><small>{profile.email}</small></div>
                    <button className="logout-button" title="Выйти" aria-label="Выйти" onClick={onLogout}><LogOut
                        size={18}/></button>
                </div>
            </div>
        </aside>
        <main className="workspace"><Outlet context={{profile, setProfile} satisfies ShellContext}/></main>
    </div>
}

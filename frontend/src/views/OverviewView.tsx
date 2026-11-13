import {ArrowUpRight, FileText, Link2, ShieldCheck, Users} from 'lucide-react'
import {Link, useOutletContext} from 'react-router-dom'
import type {ShellContext} from './ShellView'

const modules = [
    {
        to: '/documents',
        eyebrow: 'Документы',
        title: 'Пишите и делитесь идеями',
        detail: 'Markdown-редактор с предпросмотром и точечной выдачей прав.',
        action: 'Открыть документы',
        icon: FileText,
        tone: 'blue',
    },
    {
        to: '/shortlinks',
        eyebrow: 'Shortlink',
        title: 'Сокращайте длинные адреса',
        detail: 'Компактные ссылки для команды с быстрым копированием.',
        action: 'Открыть shortlink',
        icon: Link2,
        tone: 'teal',
    },
    {
        to: '/groups',
        eyebrow: 'Доступы',
        title: 'Управляйте доступом один раз',
        detail: 'Назначайте модули группам — Aura учтёт все группы пользователя.',
        action: 'Настроить группы',
        icon: Users,
        tone: 'violet',
    },
]

export function OverviewView() {
    const {profile} = useOutletContext<ShellContext>()
    const firstName = profile.firstName || profile.displayName

    return <div className="overview-page">
        <section className="overview-hero">
            <div>
                <span className="eyebrow">Рабочее пространство</span>
                <h1>Добро пожаловать, {firstName}</h1>
                <p>Документы, короткие ссылки и доступы команды — в одном спокойном пространстве.</p>
            </div>
            <div className="hero-orbit" aria-hidden="true">
                <span><FileText size={22}/></span>
                <span><Link2 size={22}/></span>
                <span><Users size={22}/></span>
                <div className="brand-mark">A</div>
            </div>
        </section>

        <div className="section-heading">
            <div><span className="eyebrow">Модули Aura</span><h2>Продолжить работу</h2></div>
            <Link className="text-link" to="/groups">Управление доступами <ArrowUpRight size={16}/></Link>
        </div>
        <div className="module-grid">
            {modules.map(({to, eyebrow, title, detail, action, icon: Icon, tone}) =>
                <Link className={`module-card ${tone}`} to={to} key={to}>
                    <div className="module-icon"><Icon size={24}/></div>
                    <span className="eyebrow">{eyebrow}</span>
                    <h3>{title}</h3>
                    <p>{detail}</p>
                    <strong>{action}<ArrowUpRight size={17}/></strong>
                </Link>)}
        </div>

        <section className="access-callout">
            <div className="callout-icon"><ShieldCheck size={24}/></div>
            <div><h3>Доступ складывается из всех групп</h3><p>Если модуль отозван в одной группе, он останется доступен
                пользователю через любую другую группу, где всё ещё включён.</p></div>
            <Link className="secondary button-link" to="/groups">Проверить группы</Link>
        </section>
    </div>
}

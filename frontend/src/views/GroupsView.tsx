import {type FormEvent, useState} from 'react'
import {Boxes, Plus, Save, Trash2, UserPlus, Users} from 'lucide-react'
import type {GroupModuleAccess, ProductModule, UserGroup} from '../model/api'

type Props = {
    groups: UserGroup[]; modules: ProductModule[]; accesses: GroupModuleAccess[]; error: string
    onCreate: (name: string) => void; onRename: (id: string, name: string) => void; onDelete: (id: string) => void
    onAdd: (id: string, email: string) => void; onRemove: (id: string, userID: string) => void
    onSetModules: (groupID: string, moduleIDs: string[]) => void
}

function ModulePicker({modules, selected, onSave}: {
    modules: ProductModule[];
    selected: string[];
    onSave: (ids: string[]) => void
}) {
    const selectedIDs = [...new Set(selected)]
    const [value, setValue] = useState(selectedIDs)

    function toggle(id: string) {
        setValue(items => items.includes(id) ? items.filter(item => item !== id) : [...items, id])
    }

    const changed = value.length !== selectedIDs.length || value.some(id => !selectedIDs.includes(id))
    return <div className="module-picker">
        <div className="subsection-heading"><span><Boxes size={18}/></span>
            <div><strong>Доступные модули</strong>
                <small className="muted">Владелец группы получает доступ автоматически</small>
            </div>
        </div>
        <div className="module-checks">{modules.map(module => <label className="module-check" key={module.id}>
            <input aria-label={module.name} type="checkbox" checked={value.includes(module.id)}
                   onChange={() => toggle(module.id)}/><span><strong>{module.name}</strong><small>{module.code === 'editor' ? 'Создание и совместная работа' : module.code === 'shortlink' ? 'Компактные ссылки и переходы' : module.code === 'calendar' ? 'События и напоминания' : 'Проекты и задачи'}</small></span>
        </label>)}
        </div>
        <button className="secondary" onClick={() => onSave([...new Set(value)])} disabled={!changed}><Save size={16}/> Сохранить
            доступы
        </button>
    </div>
}

function GroupNameEditor({group, onRename}: { group: UserGroup; onRename: (id: string, name: string) => void }) {
    const [name, setName] = useState(group.name)
    const normalizedName = name.trim()
    const changed = normalizedName !== group.name
    return <form className="group-name-form" onSubmit={event => {
        event.preventDefault();
        if (changed && normalizedName) onRename(group.id, normalizedName)
    }}>
        <input aria-label={`Название группы ${group.name}`} value={name} maxLength={120}
               onChange={event => setName(event.target.value)}/>
        <button className="secondary" disabled={!changed || !normalizedName}><Save size={15}/> Сохранить название
        </button>
    </form>
}

export function GroupsView({
                               groups,
                               modules,
                               accesses,
                               error,
                               onCreate,
                               onRename,
                               onDelete,
                               onAdd,
                               onRemove,
                               onSetModules
                           }: Props) {
    function create(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const form = e.currentTarget;
        onCreate(String(new FormData(form).get('name')));
        form.reset()
    }

    return <div className="groups-page">
        <div className="page-title">
            <div><span className="eyebrow">Команда и доступ</span><h1>Группы</h1><p>Управляйте участниками и модулями,
                которые им доступны</p></div>
        </div>
        <form className="group-create" onSubmit={create}><span className="group-create-icon"><Users
            size={20}/></span><input name="name" placeholder="Название новой группы" aria-label="Название новой группы"
                                     required maxLength={120}/>
            <button className="primary"><Plus size={18}/> Создать</button>
        </form>
        <p className="access-hint">Владелец получает доступ к модулям группы автоматически. Для остальных пользователей
            доступ сохраняется, пока модуль назначен хотя бы одной из их групп.</p>
        {error && <div className="error">{error}</div>}
        {groups.length === 0 &&
            <div className="state compact"><span className="state-icon"><Users size={27}/></span><h2>Групп пока нет</h2>
                <p>Создайте группу, добавьте участников и выберите доступные им модули.</p></div>}
        <div className="group-grid">{groups.map(group => {
            const selected = [...new Set(accesses.find(access => access.groupID === group.id)?.moduleIDs ?? [])]
            return <section className="settings-card" key={group.id}>
                <div className="group-card-header"><GroupNameEditor key={group.name} group={group} onRename={onRename}/>
                    <button className="icon-button danger" aria-label={`Удалить группу ${group.name}`}
                            title="Удалить группу" onClick={() => onDelete(group.id)}><Trash2 size={18}/></button>
                </div>
                <ModulePicker key={`${group.id}:${selected.join(',')}`} modules={modules} selected={selected}
                              onSave={ids => onSetModules(group.id, ids)}/>
                <h3 className="section-label"><Users size={17}/> Участники <span>{group.members.length}</span></h3>
                <form className="share-form" onSubmit={e => {
                    e.preventDefault();
                    const form = e.currentTarget;
                    onAdd(group.id, String(new FormData(form).get('email')));
                    form.reset()
                }}>
                    <input name="email" type="email" placeholder="Email участника" required/>
                    <button className="secondary"><UserPlus size={16}/> Добавить</button>
                </form>
                <div className="people">{group.members.length === 0 &&
                    <p className="muted">В группе пока нет участников</p>}{group.members.map(member => <div
                    className="person" key={member.userID}>
                    <span className="avatar small-avatar">{member.displayName.slice(0, 1)}</span>
                    <div><strong>{member.displayName}</strong><small>{member.email}</small></div>
                    <button onClick={() => onRemove(group.id, member.userID)}>Удалить</button>
                </div>)}</div>
            </section>
        })}</div>
    </div>
}

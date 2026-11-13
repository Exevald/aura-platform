import type {FormEvent, MouseEvent} from 'react'
import {UserPlus, X} from 'lucide-react'
import type {Permission} from '../model/api'

type Role = 'viewer' | 'editor'
type Props = {
    items: Permission[];
    error: string;
    onClose: () => void;
    onGrant: (email: string, role: Role) => void;
    onRevoke: (userID: string) => void
}

export function ShareView({items, error, onClose, onGrant, onRevoke}: Props) {
    function individual(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const form = e.currentTarget, data = new FormData(form);
        onGrant(String(data.get('email')), data.get('role') as Role);
        form.reset()
    }

    function stop(e: MouseEvent) {
        e.stopPropagation()
    }

    return <div className="dialog-backdrop" onMouseDown={onClose}>
        <section className="dialog" onMouseDown={stop}>
            <div className="dialog-title"><h2>Поделиться документом</h2>
                <button aria-label="Закрыть" onClick={onClose}><X size={21}/></button>
            </div>
            <h3>Отдельный пользователь</h3>
            <form className="share-form" onSubmit={individual}><input name="email" type="email"
                                                                      placeholder="Email пользователя"
                                                                      required/><RoleSelect/>
                <button className="primary"><UserPlus size={17}/> Добавить</button>
            </form>
            {error && <div className="error">{error}</div>}
            <div className="people"><h3>Индивидуальный доступ</h3>{items.length === 0 &&
                <p className="muted">Пока доступ есть только у владельца</p>}{items.map(p => <div className="person"
                                                                                                  key={p.userID}><span
                className="avatar small-avatar">{p.displayName.slice(0, 1)}</span>
                <div><strong>{p.displayName}</strong><small>{p.email}</small></div>
                <span>{p.role === 'editor' ? 'Редактор' : 'Читатель'}</span>
                <button onClick={() => onRevoke(p.userID)}>Удалить</button>
            </div>)}</div>
        </section>
    </div>
}

function RoleSelect() {
    return <select name="role">
        <option value="viewer">Читатель</option>
        <option value="editor">Редактор</option>
    </select>
}

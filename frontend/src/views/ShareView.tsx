import type { FormEvent, MouseEvent } from 'react'
import type { Permission } from '../model/api'

type Props = {
  items: Permission[]
  error: string
  onClose: () => void
  onGrant: (email: string, role: 'viewer' | 'editor') => void
  onRevoke: (userID: string) => void
}

export function ShareView({ items, error, onClose, onGrant, onRevoke }: Props) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = event.currentTarget
    const data = new FormData(form)
    onGrant(String(data.get('email')), data.get('role') as 'viewer' | 'editor')
    form.reset()
  }
  function stop(event: MouseEvent) { event.stopPropagation() }

  return <div className="dialog-backdrop" onMouseDown={onClose}><section className="dialog" onMouseDown={stop}><div className="dialog-title"><h2>Поделиться документом</h2><button onClick={onClose}>×</button></div><form className="share-form" onSubmit={submit}><input name="email" type="email" placeholder="Email пользователя" required/><select name="role"><option value="viewer">Читатель</option><option value="editor">Редактор</option></select><button className="primary">Добавить</button></form>{error && <div className="error">{error}</div>}<div className="people"><h3>У кого есть доступ</h3>{items.length === 0 && <p className="muted">Пока доступ есть только у владельца</p>}{items.map((permission) => <div className="person" key={permission.userID}><span className="avatar small-avatar">{permission.displayName.slice(0,1)}</span><div><strong>{permission.displayName}</strong><small>{permission.email}</small></div><span>{permission.role === 'editor' ? 'Редактор' : 'Читатель'}</span><button onClick={() => onRevoke(permission.userID)}>Удалить</button></div>)}</div></section></div>
}

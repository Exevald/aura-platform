import {type FormEvent, useState} from 'react'
import {Check, Copy, Link2, Plus, Trash2} from 'lucide-react'
import type {ShortLink} from '../model/api'

export function ShortlinkView({links, error, onCreate, onDelete}: {
    links: ShortLink[];
    error: string;
    onCreate: (url: string) => void;
    onDelete: (id: string) => void
}) {
    const [copiedID, setCopiedID] = useState('')

    function submit(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const form = e.currentTarget;
        onCreate(String(new FormData(form).get('targetURL')));
        form.reset()
    }

    async function copy(id: string, path: string) {
        try {
            await navigator.clipboard.writeText(new URL(path, location.origin).toString());
            setCopiedID(id)
        } catch {
            setCopiedID('')
        }
    }

    return <div className="shortlink-page">
        <div className="page-title">
            <div><span className="eyebrow">Shortlink</span><h1>Короткие ссылки</h1><p>Превращайте длинные URL в
                аккуратные адреса Aura</p></div>
        </div>
        <section className="shortlink-create">
            <div className="shortlink-create-icon"><Link2 size={23}/></div>
            <div><h2>Новая ссылка</h2><p>Вставьте полный адрес — короткую ссылку можно сразу скопировать.</p></div>
            <form className="share-form" onSubmit={submit}><input name="targetURL" type="url"
                                                                  placeholder="https://example.com/long-address"
                                                                  required maxLength={2048}/>
                <button className="primary"><Plus size={18}/> Сократить</button>
            </form>
        </section>
        {error && <div className="error">{error}</div>}
        <div className="link-list">{links.length === 0 &&
            <div className="state compact"><span className="state-icon"><Link2 size={27}/></span><h2>Ссылок пока
                нет</h2><p>Вставьте длинный URL выше, чтобы получить короткий адрес.</p></div>}{links.map(link => <div
            className="person link-row" key={link.id}>
            <span className="link-row-icon"><Link2 size={19}/></span>
            <div className="link-details"><strong><a href={link.shortPath} target="_blank"
                                                     rel="noreferrer">{location.origin}{link.shortPath}</a></strong><small
                title={link.targetURL}>{link.targetURL}</small></div>
            <button className="copy-button secondary"
                    onClick={() => void copy(link.id, link.shortPath)}>{copiedID === link.id ? <Check size={16}/> :
                <Copy size={16}/>} {copiedID === link.id ? 'Скопировано' : 'Копировать'}</button>
            <button className="delete-button" aria-label="Удалить ссылку" onClick={() => onDelete(link.id)}><Trash2
                size={18}/></button>
        </div>)}</div>
    </div>
}

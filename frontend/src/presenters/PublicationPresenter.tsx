import {useEffect, useState} from 'react'
import {Check, Copy, Globe2, Plus, Trash2} from 'lucide-react'
import {api, type Document, type Publication} from '../model/api'
import {copyText} from '../model/clipboard'
import {randomHex} from '../model/random'
import {errorMessage} from './error'

export function PublicationPresenter() {
    const [documents, setDocuments] = useState<Document[]>([])
    const [releases, setReleases] = useState<Publication[]>([])
    const [selected, setSelected] = useState('')
    const [error, setError] = useState('')
    const [busy, setBusy] = useState(false)
    const [copied, setCopied] = useState('')

    const refresh = async () => setReleases(await api.publications())
    useEffect(() => {
        void Promise.all([api.documents(), api.publications()]).then(([docs, items]) => {
            setDocuments(docs.filter(document => document.role === 'owner'))
            setReleases(items)
        }).catch(e => setError(errorMessage(e)))
        const timer = window.setInterval(() => void refresh().catch(() => {
        }), 3000)
        return () => window.clearInterval(timer)
    }, [])

    async function publish() {
        const document = documents.find(item => item.id === selected)
        if (!document) return
        setBusy(true)
        setError('')
        try {
            const key = randomHex(16)
            const release = await api.publishDocument(document, key)
            setReleases(items => [release, ...items.filter(item => item.id !== release.id)])
        } catch (e) {
            setError(errorMessage(e))
        } finally {
            setBusy(false)
        }
    }

    async function unpublish(id: string) {
        if (!confirm('Снять публикацию? Ссылка перестанет работать.')) return
        setError('')
        try {
            await api.unpublish(id)
            await refresh()
        } catch (e) {
            setError(errorMessage(e))
        }
    }

    async function copy(item: Publication) {
        try {
            await copyText(new URL(item.shortPath, location.origin).toString())
            setCopied(item.id)
            setError('')
        } catch {
            setCopied('')
            setError('Не удалось скопировать ссылку')
        }
    }

    const labels: Record<string, string> = {
        pending: 'Публикуется',
        active: 'Опубликовано',
        closing: 'Снимается',
        closed: 'Снято',
        failed: 'Ошибка'
    }
    return <div className="shortlink-page">
        <div className="page-title">
            <div><span className="eyebrow">Publication</span><h1>Публикации</h1><p>Создайте общедоступную копию
                документа с вложениями и короткой ссылкой.</p></div>
        </div>
        <section className="shortlink-create">
            <div className="shortlink-create-icon"><Globe2 size={23}/></div>
            <div><h2>Новая публикация</h2><p>Публикуется сохранённая версия документа. Изменения в редакторе не меняют
                опубликованную копию.</p></div>
            <div className="share-form">
                <select aria-label="Документ" value={selected} onChange={event => setSelected(event.target.value)}>
                    <option value="">Выберите документ</option>
                    {documents.map(document => <option key={document.id} value={document.id}>{document.title}</option>)}
                </select>
                <button className="primary" disabled={!selected || busy} onClick={() => void publish()}><Plus
                    size={18}/> Опубликовать
                </button>
            </div>
        </section>
        {error && <div className="error">{error}</div>}
        <div className="link-list">
            {releases.length === 0 &&
                <div className="state compact"><span className="state-icon"><Globe2 size={27}/></span><h2>Публикаций
                    пока нет</h2></div>}
            {releases.map(item => <div className="person link-row" key={item.id}>
                <span className="link-row-icon"><Globe2 size={19}/></span>
                <div className="link-details">
                    <strong>{item.title || documents.find(document => document.id === item.documentID)?.title || 'Документ'}</strong>
                    <small>{labels[item.status] || item.status}{item.deliveryStatus === 'failed' ? ' · Уведомления не доставлены' : ''}{item.failure ? ` · ${item.failure}` : ''}</small>
                    {item.status === 'active' && item.shortPath &&
                        <a href={item.shortPath} target="_blank" rel="noreferrer">{location.origin}{item.shortPath}</a>}
                </div>
                {item.status === 'active' &&
                    <button className="copy-button secondary" onClick={() => void copy(item)}>{copied === item.id ?
                        <Check size={16}/> :
                        <Copy size={16}/>} {copied === item.id ? 'Скопировано' : 'Копировать'}</button>}
                {(item.status === 'active' || item.status === 'pending') &&
                    <button className="delete-button" aria-label="Снять публикацию"
                            onClick={() => void unpublish(item.id)}><Trash2 size={18}/></button>}
            </div>)}
        </div>
    </div>
}

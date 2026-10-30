import type {ReactNode} from 'react'
import ReactMarkdown from 'react-markdown'
import type {Document} from '../model/api'

type Props = {
    document: Document
    status: string
    sharing: boolean
    onBack: () => void
    onChange: (patch: Partial<Pick<Document, 'title' | 'content'>>) => void
    onOpenSharing: () => void
    onCloseSharing: () => void
    shareDialog?: ReactNode
}

export function EditorView({
                               document,
                               status,
                               sharing,
                               onBack,
                               onChange,
                               onOpenSharing,
                               onCloseSharing,
                               shareDialog
                           }: Props) {
    const editable = document.role !== 'viewer'
    return <div className="editor-page">
        <div className="editor-bar">
            <button className="icon-button" onClick={onBack}>←</button>
            <input className="title-input" value={document.title} readOnly={!editable}
                   onChange={(event) => onChange({title: event.target.value})}/><span
            className={`save-status ${status.startsWith('Конфликт') ? 'danger' : ''}`}>{document.role === 'viewer' ? 'Только просмотр' : status}</span>{document.role === 'owner' &&
            <button className="primary small" onClick={onOpenSharing}>Поделиться</button>}</div>
        <div className="editor-split"><textarea aria-label="Markdown" value={document.content} readOnly={!editable}
                                                onChange={(event) => onChange({content: event.target.value})}
                                                placeholder="# Начните писать…"/>
            <article className="markdown">
                <ReactMarkdown>{document.content || '*Предпросмотр появится здесь*'}</ReactMarkdown></article>
        </div>
        {sharing && <div onKeyDown={(event) => event.key === 'Escape' && onCloseSharing()}>{shareDialog}</div>}
    </div>
}

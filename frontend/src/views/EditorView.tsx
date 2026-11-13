import {type ReactNode, useState} from 'react'
import {ArrowLeft, Eye, PenLine, Share2} from 'lucide-react'
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
    const [mobileMode, setMobileMode] = useState<'write' | 'preview'>('write')
    return <div className="editor-page">
        <div className="editor-bar">
            <button className="icon-button" aria-label="К документам" onClick={onBack}><ArrowLeft size={19}/></button>
            <input className="title-input" value={document.title} readOnly={!editable}
                   onChange={(event) => onChange({title: event.target.value})}/><span
            className={`save-status ${status.startsWith('Конфликт') ? 'danger' : ''}`}>{document.role === 'viewer' ? 'Только просмотр' : status}</span>{document.role === 'owner' &&
            <button className="primary small" onClick={onOpenSharing}><Share2 size={16}/> Поделиться</button>}</div>
        <div className="editor-mobile-tabs" role="tablist" aria-label="Режим редактора">
            <button className={mobileMode === 'write' ? 'active' : ''} onClick={() => setMobileMode('write')}><PenLine
                size={16}/> Редактор
            </button>
            <button className={mobileMode === 'preview' ? 'active' : ''} onClick={() => setMobileMode('preview')}><Eye
                size={16}/> Просмотр
            </button>
        </div>
        <div className={`editor-split mobile-${mobileMode}`}><textarea aria-label="Markdown" value={document.content}
                                                                       readOnly={!editable}
                                                                       onChange={(event) => onChange({content: event.target.value})}
                                                                       placeholder="# Начните писать…"/>
            <article className="markdown">
                <ReactMarkdown>{document.content || '*Предпросмотр появится здесь*'}</ReactMarkdown></article>
        </div>
        {sharing && <div onKeyDown={(event) => event.key === 'Escape' && onCloseSharing()}>{shareDialog}</div>}
    </div>
}

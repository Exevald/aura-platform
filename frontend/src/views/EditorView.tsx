import {type ChangeEvent, type DragEvent, type ReactNode, useEffect, useRef, useState} from 'react'
import {ArrowLeft, Eye, Paperclip, PenLine, Share2} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import type {Attachment, Document} from '../model/api'
import {AttachmentImage, AttachmentLink} from '../components/AttachmentPreview'

type Props = {
    document: Document
    status: string
    sharing: boolean
    onBack: () => void
    onChange: (patch: Partial<Pick<Document, 'title' | 'content'>>) => void
    onUpload: (file: File) => Promise<Attachment>
    onOpenSharing: () => void
    onCloseSharing: () => void
    shareDialog?: ReactNode
}

type Insertion = { start: number; end: number }
const imageTypes = new Set(['image/png', 'image/jpeg', 'image/gif', 'image/webp'])

function moveOffset(offset: number, changedAt: number, removed: number, added: number) {
    if (offset <= changedAt) return offset
    if (offset >= changedAt + removed) return offset + added - removed
    return changedAt + added
}

function escapeLabel(value: string) {
    return value.replace(/[\\[\]]/g, '\\$&')
}

export function EditorView({
                               document,
                               status,
                               sharing,
                               onBack,
                               onChange,
                               onUpload,
                               onOpenSharing,
                               onCloseSharing,
                               shareDialog
                           }: Props) {
    const editable = document.role !== 'viewer'
    const [mobileMode, setMobileMode] = useState<'write' | 'preview'>('write')
    const [dragging, setDragging] = useState(false)
    const [uploading, setUploading] = useState(false)
    const [uploadError, setUploadError] = useState('')
    const textarea = useRef<HTMLTextAreaElement>(null)
    const imageInput = useRef<HTMLInputElement>(null)
    const fileInput = useRef<HTMLInputElement>(null)
    const insertAt = useRef<Insertion | null>(null)
    const content = useRef(document.content)
    useEffect(() => {
        content.current = document.content
    }, [document.content])

    function rememberSelection() {
        const field = textarea.current
        insertAt.current = field ? {
            start: field.selectionStart,
            end: field.selectionEnd
        } : {start: content.current.length, end: content.current.length}
    }

    function changeContent(next: string) {
        const previous = content.current
        if (insertAt.current && previous !== next) {
            let prefix = 0
            while (prefix < previous.length && prefix < next.length && previous[prefix] === next[prefix]) prefix++
            let suffix = 0
            while (suffix < previous.length - prefix && suffix < next.length - prefix && previous[previous.length - 1 - suffix] === next[next.length - 1 - suffix]) suffix++
            const removed = previous.length - prefix - suffix
            const added = next.length - prefix - suffix
            insertAt.current = {
                start: moveOffset(insertAt.current.start, prefix, removed, added),
                end: moveOffset(insertAt.current.end, prefix, removed, added),
            }
        }
        content.current = next
        onChange({content: next})
    }

    async function uploadFiles(files: File[]) {
        if (!editable || uploading || files.length === 0) return
        if (!insertAt.current) rememberSelection()
        setUploading(true)
        setUploadError('')
        for (const [index, file] of files.entries()) {
            try {
                const attachment = await onUpload(file)
                const path = `/api/editor/v1/documents/${document.id}/attachments/${attachment.FileID}/content`
                const label = escapeLabel(attachment.FileName)
                const markdown = `${index > 0 ? '\n' : ''}${imageTypes.has(attachment.ContentType.toLowerCase().split(';')[0].trim()) ? '!' : ''}[${label}](${path})`
                const range = insertAt.current ?? {start: content.current.length, end: content.current.length}
                const next = content.current.slice(0, range.start) + markdown + content.current.slice(range.end)
                insertAt.current = null
                changeContent(next)
                const cursor = range.start + markdown.length
                insertAt.current = {start: cursor, end: cursor}
                textarea.current?.focus()
                textarea.current?.setSelectionRange(cursor, cursor)
            } catch (error) {
                setUploadError(`Не удалось загрузить «${file.name}»: ${error instanceof Error ? error.message : 'Ошибка запроса'}`)
                break
            }
        }
        insertAt.current = null
        setUploading(false)
    }

    function chooseFiles(event: ChangeEvent<HTMLInputElement>) {
        const files = Array.from(event.target.files ?? [])
        event.target.value = ''
        void uploadFiles(files)
    }

    function dropFiles(event: DragEvent<HTMLTextAreaElement>) {
        if (!event.dataTransfer.files.length) return
        event.preventDefault()
        setDragging(false)
        if (!editable) return
        rememberSelection()
        void uploadFiles(Array.from(event.dataTransfer.files))
    }

    return <div className="editor-page">
        <div className="editor-bar">
            <button className="icon-button" aria-label="К документам" onClick={onBack}><ArrowLeft size={19}/></button>
            <input className="title-input" value={document.title} readOnly={!editable}
                   onChange={(event) => onChange({title: event.target.value})}/>
            <span
                className={`save-status ${status.startsWith('Конфликт') ? 'danger' : ''}`}>{editable ? status : 'Только просмотр'}</span>
            {editable && <details className="editor-insert-menu">
                <summary aria-label="Вставить вложение"><Paperclip size={16}/> Вставить</summary>
                <div className="editor-insert-options">
                    <button type="button" disabled={uploading} onClick={(event) => {
                        rememberSelection();
                        event.currentTarget.closest('details')?.removeAttribute('open');
                        imageInput.current?.click()
                    }}>Изображение
                    </button>
                    <button type="button" disabled={uploading} onClick={(event) => {
                        rememberSelection();
                        event.currentTarget.closest('details')?.removeAttribute('open');
                        fileInput.current?.click()
                    }}>Файл
                    </button>
                </div>
            </details>}
            {document.role === 'owner' &&
                <button className="primary small" onClick={onOpenSharing}><Share2 size={16}/> Поделиться</button>}
        </div>
        {editable && <>
            <input ref={imageInput} type="file" accept="image/png,image/jpeg,image/gif,image/webp" multiple hidden
                   aria-label="Выбрать изображения" onChange={chooseFiles}/>
            <input ref={fileInput} type="file" multiple hidden aria-label="Выбрать файлы" onChange={chooseFiles}/>
        </>}
        {(uploading || uploadError) && <div className={`editor-upload-status ${uploadError ? 'danger' : ''}`}
                                            role="status">{uploadError || 'Загрузка вложений…'}</div>}
        <div className="editor-mobile-tabs" role="tablist" aria-label="Режим редактора">
            <button className={mobileMode === 'write' ? 'active' : ''} onClick={() => setMobileMode('write')}><PenLine
                size={16}/> Редактор
            </button>
            <button className={mobileMode === 'preview' ? 'active' : ''} onClick={() => setMobileMode('preview')}><Eye
                size={16}/> Просмотр
            </button>
        </div>
        <div className={`editor-split mobile-${mobileMode} ${dragging ? 'editor-dragging' : ''}`}>
            <textarea ref={textarea} aria-label="Markdown" value={document.content} readOnly={!editable}
                      onChange={(event) => changeContent(event.target.value)}
                      onDragOver={(event) => {
                          if (Array.from(event.dataTransfer.types).includes('Files')) {
                              event.preventDefault();
                              if (editable) setDragging(true)
                          }
                      }}
                      onDragLeave={() => setDragging(false)} onDrop={dropFiles}
                      placeholder="# Начните писать…"/>
            <article className="markdown"><ReactMarkdown components={{
                img: AttachmentImage,
                a: AttachmentLink
            }}>{document.content || '*Предпросмотр появится здесь*'}</ReactMarkdown></article>
        </div>
        {sharing && <div onKeyDown={(event) => event.key === 'Escape' && onCloseSharing()}>{shareDialog}</div>}
    </div>
}

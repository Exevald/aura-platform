import {type MouseEvent, type ReactNode, useEffect, useState} from 'react'
import {api} from '../model/api'

const attachmentPath = /^\/api\/editor\/v1\/documents\/([0-9a-f-]{36})\/attachments\/([0-9a-f-]{36})\/content$/i

export function AttachmentImage({src, alt}: { src?: string; alt?: string }) {
    const match = src?.match(attachmentPath)
    const documentID = match?.[1]
    const fileID = match?.[2]
    const [objectURL, setObjectURL] = useState<{ src: string; value: string }>()
    const [error, setError] = useState<string>()

    useEffect(() => {
        if (!documentID || !fileID || !src) return
        let active = true
        let createdURL: string | undefined
        api.downloadAttachment(documentID, fileID).then((blob) => {
            createdURL = URL.createObjectURL(blob)
            if (active) setObjectURL({src, value: createdURL})
            else URL.revokeObjectURL(createdURL)
        }).catch(() => {
            if (active) setError(src)
        })
        return () => {
            active = false
            if (createdURL) URL.revokeObjectURL(createdURL)
        }
    }, [documentID, fileID, src])

    if (!match) return <img src={src} alt={alt}/>
    if (error === src) return <span role="status">Не удалось загрузить изображение: {alt}</span>
    return objectURL && objectURL.src === src ? <img src={objectURL.value} alt={alt}/> :
        <span role="status">Загрузка изображения…</span>
}

export function AttachmentLink({href, children}: { href?: string; children?: ReactNode }) {
    const match = href?.match(attachmentPath)
    const [error, setError] = useState(false)

    async function download(event: MouseEvent<HTMLAnchorElement>) {
        if (!match) return
        event.preventDefault()
        setError(false)
        try {
            const blob = await api.downloadAttachment(match[1], match[2])
            const objectURL = URL.createObjectURL(blob)
            const link = document.createElement('a')
            link.href = objectURL
            link.download = typeof children === 'string' ? children : 'attachment'
            link.click()
            setTimeout(() => URL.revokeObjectURL(objectURL), 0)
        } catch {
            setError(true)
        }
    }

    return <><a href={href} onClick={(event) => {
        void download(event)
    }}>{children}</a>{error && <span role="status"> Не удалось загрузить вложение</span>}</>
}

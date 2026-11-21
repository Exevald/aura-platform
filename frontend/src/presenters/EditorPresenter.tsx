import {useEffect, useRef, useState} from 'react'
import {useNavigate, useParams} from 'react-router-dom'
import {api, ApiError, type Document} from '../model/api'
import {EditorView} from '../views/EditorView'
import {StateView} from '../views/StateView'
import {errorMessage} from './error'
import {SharePresenter} from './SharePresenter'

export function EditorPresenter() {
    const {id = ''} = useParams()
    const navigate = useNavigate()
    const [document, setDocument] = useState<Document | null>(null)
    const [dirty, setDirty] = useState(false)
    const [status, setStatus] = useState('')
    const [error, setError] = useState('')
    const [sharing, setSharing] = useState(false)
    const [saveTick, setSaveTick] = useState(0)
    const revision = useRef(0)
    const saving = useRef(false)

    useEffect(() => {
        api.document(id).then(setDocument).catch((requestError) => setError(errorMessage(requestError)))
    }, [id])
    useEffect(() => {
        if (!document || !dirty || document.role === 'viewer' || saving.current) return
        const timer = setTimeout(() => {
            saving.current = true
            const savedRevision = revision.current
            api.updateDocument(document).then((saved) => {
                const changed = revision.current !== savedRevision
                setDocument((current) => current?.id === saved.id ? {...current, version: saved.version} : current)
                setDirty(changed)
                setStatus(changed ? 'Сохранение…' : 'Сохранено')
            }).catch((requestError) => {
                setStatus(requestError instanceof ApiError && requestError.status === 409 ? 'Конфликт версии — перезагрузите документ' : errorMessage(requestError))
                setDirty(false)
            }).finally(() => {
                saving.current = false
                setSaveTick((value) => value + 1)
            })
        }, 800)
        return () => clearTimeout(timer)
    }, [document, dirty, saveTick])

    function change(patch: Partial<Pick<Document, 'title' | 'content'>>) {
        revision.current += 1
        setDocument((value) => value ? {...value, ...patch} : value)
        setStatus('Сохранение…')
        setDirty(true)
    }

    if (error) return <StateView title={error} action="К документам" onAction={() => navigate('/documents')}/>
    if (!document) return <StateView title="Открываем документ…"/>
    return <EditorView document={document} status={status} sharing={sharing} onBack={() => navigate('/documents')}
                       onChange={change} onUpload={(file) => api.uploadAttachment(document.id, file)}
                       onOpenSharing={() => setSharing(true)} onCloseSharing={() => setSharing(false)}
                       shareDialog={<SharePresenter documentID={document.id} onClose={() => setSharing(false)}/>}/>
}

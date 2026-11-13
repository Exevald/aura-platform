import {useEffect, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import {api, type Document} from '../model/api'
import {DocumentsView} from '../views/DocumentsView'
import {StateView} from '../views/StateView'
import {errorMessage} from './error'

export function DocumentsPresenter() {
    const navigate = useNavigate()
    const [documents, setDocuments] = useState<Document[] | null>(null)
    const [error, setError] = useState('')

    useEffect(() => {
        api.documents().then(setDocuments).catch((requestError) => setError(errorMessage(requestError)))
    }, [])

    async function create() {
        try {
            const document = await api.createDocument()
            navigate(`/documents/${document.id}`)
        } catch (requestError) {
            setError(errorMessage(requestError))
        }
    }

    async function remove(document: Document) {
        if (document.role !== 'owner' || !confirm(`Удалить «${document.title}»?`)) return
        try {
            await api.deleteDocument(document.id)
            setDocuments((items) => items?.filter((item) => item.id !== document.id) ?? null)
        } catch (requestError) {
            setError(errorMessage(requestError))
        }
    }

    return <DocumentsView documents={documents} error={error} onCreate={create}
                          onOpen={(id) => navigate(`/documents/${id}`)} onRemove={remove}/>
}

export function NewDocumentPresenter() {
    const navigate = useNavigate()
    useEffect(() => {
        api.createDocument().then((document) => navigate(`/documents/${document.id}`, {replace: true})).catch(() => navigate('/documents', {replace: true}))
    }, [navigate])
    return <StateView title="Создаём документ…"/>
}

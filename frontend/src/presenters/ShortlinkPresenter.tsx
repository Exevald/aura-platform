import {useEffect, useState} from 'react'
import {api, type ShortLink} from '../model/api'
import {ShortlinkView} from '../views/ShortlinkView'
import {errorMessage} from './error'

export function ShortlinkPresenter() {
    const [links, setLinks] = useState<ShortLink[]>([])
    const [error, setError] = useState('')
    const load = () => api.shortLinks().then(setLinks)
    useEffect(() => {
        void load().catch(e => setError(errorMessage(e)))
    }, [])

    async function create(targetURL: string) {
        try {
            setError('');
            const link = await api.createShortLink(targetURL);
            setLinks(items => [link, ...items])
        } catch (e) {
            setError(errorMessage(e))
        }
    }

    async function remove(id: string) {
        try {
            setError('');
            await api.deleteShortLink(id);
            setLinks(items => items.filter(item => item.id !== id))
        } catch (e) {
            setError(errorMessage(e))
        }
    }

    return <ShortlinkView links={links} error={error} onCreate={create} onDelete={remove}/>
}

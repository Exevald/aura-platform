import {useCallback, useEffect, useState} from 'react'
import {api, type Permission} from '../model/api'
import {ShareView} from '../views/ShareView'
import {errorMessage} from './error'

export function SharePresenter({documentID, onClose}: { documentID: string; onClose: () => void }) {
    const [items, setItems] = useState<Permission[]>([])
    const [error, setError] = useState('')
    const load = useCallback(() => api.permissions(documentID).then(setItems).catch((requestError) => setError(errorMessage(requestError))), [documentID])

    useEffect(() => {
        void load()
    }, [load])

    async function grant(email: string, role: 'viewer' | 'editor') {
        try {
            await api.grant(documentID, email, role)
            await load()
        } catch (requestError) {
            setError(errorMessage(requestError))
        }
    }

    async function revoke(userID: string) {
        try {
            await api.revoke(documentID, userID)
            await load()
        } catch (requestError) {
            setError(errorMessage(requestError))
        }
    }

    return <ShareView items={items} error={error} onClose={onClose} onGrant={grant} onRevoke={revoke}/>
}

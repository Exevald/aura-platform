import {useCallback, useEffect, useState} from 'react'
import {api, type Permission} from '../model/api'
import {ShareView} from '../views/ShareView'
import {errorMessage} from './error'

export function SharePresenter({documentID, onClose}: { documentID: string; onClose: () => void }) {
    const [items, setItems] = useState<Permission[]>([])
    const [error, setError] = useState('')
    const load = useCallback(async () => {
        setItems(await api.permissions(documentID))
    }, [documentID])
    useEffect(() => {
        void api.permissions(documentID)
            .then(setItems)
            .catch(e => setError(errorMessage(e)))
    }, [documentID])

    async function run(action: () => Promise<unknown>) {
        try {
            setError('');
            await action();
            await load()
        } catch (e) {
            setError(errorMessage(e))
        }
    }

    return <ShareView items={items} error={error} onClose={onClose}
                      onGrant={(email, role) => run(() => api.grant(documentID, email, role))}
                      onRevoke={(userID) => run(() => api.revoke(documentID, userID))}/>
}

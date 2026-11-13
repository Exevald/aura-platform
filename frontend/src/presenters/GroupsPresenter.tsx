import {useCallback, useEffect, useState} from 'react'
import {api, type GroupModuleAccess, type ProductModule, type UserGroup} from '../model/api'
import {GroupsView} from '../views/GroupsView'
import {errorMessage} from './error'

export function GroupsPresenter() {
    const [groups, setGroups] = useState<UserGroup[]>([])
    const [modules, setModules] = useState<ProductModule[]>([])
    const [accesses, setAccesses] = useState<GroupModuleAccess[]>([])
    const [error, setError] = useState('')
    const load = useCallback(async () => {
        const [availableGroups, availableModules, currentAccesses] = await Promise.all([api.groups(), api.modules(), api.groupModuleAccesses()])
        setGroups(availableGroups);
        setModules(availableModules);
        setAccesses(currentAccesses)
    }, [])
    useEffect(() => {
        void Promise.all([api.groups(), api.modules(), api.groupModuleAccesses()])
            .then(([availableGroups, availableModules, currentAccesses]) => {
                setGroups(availableGroups);
                setModules(availableModules);
                setAccesses(currentAccesses)
            })
            .catch(e => setError(errorMessage(e)))
    }, [])

    async function run(action: () => Promise<unknown>) {
        try {
            setError('');
            await action();
            await load()
        } catch (e) {
            setError(errorMessage(e))
        }
    }

    return <GroupsView groups={groups} modules={modules} accesses={accesses} error={error}
                       onCreate={(name) => run(() => api.createGroup(name))}
                       onRename={(id, name) => run(() => api.updateGroup(id, name))}
                       onDelete={(id) => run(() => api.deleteGroup(id))}
                       onAdd={(id, email) => run(() => api.addGroupMember(id, email))}
                       onRemove={(id, userID) => run(() => api.removeGroupMember(id, userID))}
                       onSetModules={(groupID, moduleIDs) => run(() => api.setGroupModules(groupID, moduleIDs))}/>
}

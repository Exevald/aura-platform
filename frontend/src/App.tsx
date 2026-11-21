import {Navigate, Route, Routes} from 'react-router-dom'
import {AuthPresenter} from './presenters/AuthPresenter'
import {DocumentsPresenter, NewDocumentPresenter} from './presenters/DocumentsPresenter'
import {EditorPresenter} from './presenters/EditorPresenter'
import {ProfilePresenter} from './presenters/ProfilePresenter'
import {ShellPresenter} from './presenters/ShellPresenter'
import {GroupsPresenter} from './presenters/GroupsPresenter'
import {ShortlinkPresenter} from './presenters/ShortlinkPresenter'
import {PublicationPresenter} from './presenters/PublicationPresenter'
import {OverviewView} from './views/OverviewView'
import {CalendarPresenter} from './presenters/CalendarPresenter'
import {PlannerPresenter} from './presenters/PlannerPresenter'
import {NotificationsPresenter} from './presenters/NotificationsPresenter'
import {SSOCallbackPresenter} from './presenters/SSOCallbackPresenter'
import {isAuraHost, productPaths} from './model/hosts'

export default function App() {
    const defaultPath = isAuraHost() ? productPaths[location.hostname.split('.')[1]] ?? '/login' : '/overview'
    return <Routes>
        <Route path="/" element={<Navigate to={defaultPath} replace/>}/>
        <Route path="/sso/callback" element={<SSOCallbackPresenter/>}/>
        <Route path="/login" element={<AuthPresenter/>}/>
        <Route element={<ShellPresenter/>}>
            <Route path="/overview" element={<OverviewView/>}/>
            <Route path="/documents" element={<DocumentsPresenter/>}/>
            <Route path="/documents/new" element={<NewDocumentPresenter/>}/>
            <Route path="/documents/:id" element={<EditorPresenter/>}/>
            <Route path="/profile" element={<ProfilePresenter/>}/>
            <Route path="/groups" element={<GroupsPresenter/>}/>
            <Route path="/shortlinks" element={<ShortlinkPresenter/>}/>
            <Route path="/publications" element={<PublicationPresenter/>}/>
            <Route path="/calendar" element={<CalendarPresenter/>}/>
            <Route path="/planner" element={<PlannerPresenter/>}/>
            <Route path="/notifications" element={<NotificationsPresenter/>}/>
        </Route>
        <Route path="*" element={<Navigate to={defaultPath} replace/>}/>
    </Routes>
}

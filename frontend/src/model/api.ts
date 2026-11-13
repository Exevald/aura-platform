import {sessionCoordinator} from './session'

export type Profile = {
    userID: string
    displayName: string
    firstName: string
    lastName: string
    email: string
}

export type Role = 'owner' | 'editor' | 'viewer'

export type Document = {
    id: string
    ownerUserID: string
    title: string
    content: string
    version: number
    role: Role
    createdAt: string
    updatedAt: string
}

export type Permission = {
    userID: string
    email: string
    displayName: string
    role: Exclude<Role, 'owner'>
    updatedAt: string
}

export type GroupMember = { userID: string; email: string; displayName: string }
export type UserGroup = {
    id: string;
    ownerUserID: string;
    name: string;
    members: GroupMember[];
    createdAt: string;
    updatedAt: string
}
export type ProductModule = { id: string; code: 'editor' | 'shortlink' | 'calendar' | 'planner'; name: string }
export type GroupModuleAccess = { groupID: string; moduleIDs: string[] }
export type ShortLink = { id: string; code: string; targetURL: string; shortPath: string; createdAt: string }
export type WorkspaceRole = 'owner' | 'editor' | 'viewer'
export type Workspace = {
    id: string;
    kind: 'personal' | 'team';
    name: string;
    role: WorkspaceRole;
    members: Array<{ userID: string; role: WorkspaceRole; createdAt: string }>;
    createdAt: string;
    updatedAt: string
}
export type Notification = {
    id: string;
    workspaceID?: string;
    kind: string;
    title: string;
    body: string;
    read: boolean;
    createdAt: string
}
export type Calendar = { id:string; workspaceID:string; ownerUserID:string; name:string; timezone:string; createdAt:string; updatedAt:string }
export type CalendarEvent = { id:string; calendarID:string; title:string; description:string; startsAt:string; endsAt:string; timezone:string; rrule?:string; attendees:Array<{userID?:string;email:string;status:'pending'|'accepted'|'declined'|'tentative'}>; reminders:Array<{minutesBefore:number;inApp:boolean;email:boolean}>; version:number; createdAt:string; updatedAt:string }
export type Project = { id:string; workspaceID:string; name:string; createdAt:string; updatedAt:string }
export type Task = { id:string; projectID:string; title:string; description:string; status:'todo'|'in_progress'|'done'; priority:'low'|'normal'|'high'; assigneeUserID?:string; dueAt?:string; syncToCalendar:boolean; calendarID?:string; calendarEventID?:string; timezone?:string; version:number; createdAt:string; updatedAt:string }

export type Registration = {
    login: string
    password: string
    displayName: string
    firstName: string
    lastName: string
    email: string
    rememberMe: boolean
}

export class ApiError extends Error {
    constructor(public status: number, message: string, public fields?: Record<string, string>) {
        super(message)
    }
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
    const storage = window.localStorage
    const workspaceID = typeof storage?.getItem === 'function' ? storage.getItem('aura-workspace-id') : null
    const response = await fetch(url, {
        credentials: 'include',
        ...init,
        headers: {'Content-Type': 'application/json', ...(workspaceID ? {'X-Workspace-ID': workspaceID} : {}), ...init?.headers},
    })
    if (!response.ok) {
        const body = await response.json().catch(() => ({message: 'Ошибка запроса'})) as {
            message?: string;
            fields?: Record<string, string>
        }
        if (response.status === 401) sessionCoordinator.publish('logged-out')
        const fieldMessage = body.fields ? Object.entries(body.fields).map(([field, reason]) => `${field}: ${reason}`).join('; ') : ''
        throw new ApiError(response.status, fieldMessage || body.message || 'Ошибка запроса', body.fields)
    }
    return response.status === 204 ? undefined as T : response.json() as Promise<T>
}

export const api = {
    register: (data: Registration) => request<{
        authenticated: boolean
    }>('/api/public/v1/registerer/register', {method: 'POST', body: JSON.stringify(data)}),
    login: (login: string, password: string, rememberMe: boolean) => request('/api/public/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({login, password, rememberMe})
    }),
    logout: () => request('/api/public/v1/auth/logout', {method: 'POST'}),
    changePassword: (currentPassword: string, newPassword: string) => request('/api/public/v1/auth/change-password', {
        method: 'POST',
        body: JSON.stringify({currentPassword, newPassword})
    }),
    me: () => request<Profile>('/api/bff/v1/me'),
    updateMe: (profile: Omit<Profile, 'userID'>) => request<Profile>('/api/bff/v1/me', {
        method: 'PUT',
        body: JSON.stringify(profile)
    }),
    documents: () => request<Document[]>('/api/editor/v1/documents'),
    document: (id: string) => request<Document>(`/api/editor/v1/documents/${id}`),
    createDocument: (title = 'Без названия') => request<Document>('/api/editor/v1/documents', {
        method: 'POST',
        body: JSON.stringify({title, content: ''})
    }),
    updateDocument: (document: Pick<Document, 'id' | 'title' | 'content' | 'version'>) => request<Document>(`/api/editor/v1/documents/${document.id}`, {
        method: 'PUT',
        body: JSON.stringify({title: document.title, content: document.content, version: document.version})
    }),
    deleteDocument: (id: string) => request(`/api/editor/v1/documents/${id}`, {method: 'DELETE'}),
    permissions: (id: string) => request<Permission[]>(`/api/editor/v1/documents/${id}/permissions`),
    grant: (id: string, email: string, role: 'viewer' | 'editor') => request(`/api/editor/v1/documents/${id}/permissions`, {
        method: 'PUT',
        body: JSON.stringify({email, role})
    }),
    revoke: (id: string, userID: string) => request(`/api/editor/v1/documents/${id}/permissions/${userID}`, {method: 'DELETE'}),
    groups: () => request<UserGroup[]>('/api/usergroups/v1/groups'),
    createGroup: (name: string) => request<UserGroup>('/api/usergroups/v1/groups', {
        method: 'POST',
        body: JSON.stringify({name})
    }),
    updateGroup: (id: string, name: string) => request<UserGroup>(`/api/usergroups/v1/groups/${id}`, {
        method: 'PUT',
        body: JSON.stringify({name})
    }),
    deleteGroup: (id: string) => request(`/api/usergroups/v1/groups/${id}`, {method: 'DELETE'}),
    addGroupMember: (id: string, email: string) => request<UserGroup>(`/api/usergroups/v1/groups/${id}/members`, {
        method: 'POST',
        body: JSON.stringify({email})
    }),
    removeGroupMember: (id: string, userID: string) => request(`/api/usergroups/v1/groups/${id}/members/${userID}`, {method: 'DELETE'}),
    modules: () => request<ProductModule[]>('/api/access/v1/modules'),
    groupModuleAccesses: () => request<GroupModuleAccess[]>('/api/access/v1/group-module-accesses'),
    setGroupModules: (groupID: string, moduleIDs: string[]) => request<GroupModuleAccess>(`/api/access/v1/groups/${groupID}/modules`, {
        method: 'PUT',
        body: JSON.stringify({moduleIDs})
    }),
    shortLinks: () => request<ShortLink[]>('/api/shortlink/v1/links'),
    createShortLink: (targetURL: string) => request<ShortLink>('/api/shortlink/v1/links', {
        method: 'POST',
        body: JSON.stringify({targetURL})
    }),
    deleteShortLink: (id: string) => request(`/api/shortlink/v1/links/${id}`, {method: 'DELETE'}),
    workspaces: () => request<Workspace[]>('/api/workspaces/v1/workspaces'),
    createWorkspace: (name: string) => request<Workspace>('/api/workspaces/v1/workspaces', {
        method: 'POST',
        body: JSON.stringify({name})
    }),
    notifications: () => request<Notification[]>('/api/notifications/v1/notifications'),
    markNotificationRead: (id: string) => request(`/api/notifications/v1/notifications/${id}/read`, {method: 'PUT'}),
    notificationPreferences: () => request<{ inApp: boolean; email: boolean }>('/api/notifications/v1/preferences'),
    saveNotificationPreferences: (value: { inApp: boolean; email: boolean }) => request<{
        inApp: boolean;
        email: boolean
    }>('/api/notifications/v1/preferences', {method: 'PUT', body: JSON.stringify(value)}),
    calendars: () => request<Calendar[]>('/api/calendar/v1/calendars'),
    createCalendar: (name:string, timezone:string) => request<Calendar>('/api/calendar/v1/calendars',{method:'POST',body:JSON.stringify({name,timezone})}),
    deleteCalendar: (id:string) => request(`/api/calendar/v1/calendars/${id}`,{method:'DELETE'}),
    calendarEvents: (calendarID:string) => request<CalendarEvent[]>(`/api/calendar/v1/calendars/${calendarID}/events`),
    createCalendarEvent: (calendarID:string,value:Pick<CalendarEvent,'title'|'description'|'startsAt'|'endsAt'|'timezone'|'attendees'|'reminders'> & {rrule?:string}) => request<CalendarEvent>(`/api/calendar/v1/calendars/${calendarID}/events`,{method:'POST',body:JSON.stringify(value)}),
    deleteCalendarEvent: (id:string) => request(`/api/calendar/v1/events/${id}`,{method:'DELETE'}),
    projects: () => request<Project[]>('/api/planner/v1/projects'),
    createProject: (name:string) => request<Project>('/api/planner/v1/projects',{method:'POST',body:JSON.stringify({name})}),
    deleteProject: (id:string) => request(`/api/planner/v1/projects/${id}`,{method:'DELETE'}),
    tasks: (projectID:string) => request<Task[]>(`/api/planner/v1/projects/${projectID}/tasks`),
    createTask: (projectID:string,value:Pick<Task,'title'|'description'|'status'|'priority'|'syncToCalendar'> & Partial<Pick<Task,'assigneeUserID'|'dueAt'|'calendarID'|'timezone'>>) => request<Task>(`/api/planner/v1/projects/${projectID}/tasks`,{method:'POST',body:JSON.stringify(value)}),
    updateTask: (value:Task) => request<Task>(`/api/planner/v1/tasks/${value.id}`,{method:'PUT',body:JSON.stringify({title:value.title,description:value.description,status:value.status,priority:value.priority,assigneeUserID:value.assigneeUserID,dueAt:value.dueAt,syncToCalendar:value.syncToCalendar,calendarID:value.calendarID,timezone:value.timezone,version:value.version})}),
    deleteTask: (id:string) => request(`/api/planner/v1/tasks/${id}`,{method:'DELETE'}),
}

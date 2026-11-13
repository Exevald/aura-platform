import {cleanup, fireEvent, render, screen} from '@testing-library/react'
import {afterEach, expect, test, vi} from 'vitest'
import {GroupsView} from './GroupsView'
import type {ProductModule, UserGroup} from '../model/api'

const group: UserGroup = {
    id: '5aade9af-0ed9-49ca-849f-1e9cf70be9a7',
    ownerUserID: '4bdd56b3-32be-4b28-8d9c-0baf8ab92f1f',
    name: 'Редакторы',
    members: [{userID: '964d0e17-a89e-4781-8e95-e627277157c2', email: 'member@example.com', displayName: 'Member'}],
    createdAt: '2026-11-09T19:00:00Z',
    updatedAt: '2026-11-09T19:00:00Z',
}

const modules: ProductModule[] = [
    {id: 'b8547c76-51fd-4f7c-9c68-8f30c84f61cb', code: 'editor', name: 'Редактор документов'},
    {id: 'd5b2367d-af92-4fa8-bc65-93b41f37a397', code: 'shortlink', name: 'Короткие ссылки'},
]

afterEach(cleanup)

test('saves a group name explicitly instead of sending a request on every keystroke', () => {
    const onRename = vi.fn()
    render(<GroupsView groups={[group]} modules={modules} accesses={[]} error="" onCreate={vi.fn()} onRename={onRename}
                       onDelete={vi.fn()} onAdd={vi.fn()} onRemove={vi.fn()} onSetModules={vi.fn()}/>)

    fireEvent.change(screen.getByLabelText('Название группы Редакторы'), {target: {value: 'Команда редакторов'}})
    expect(onRename).not.toHaveBeenCalled()
    fireEvent.click(screen.getByRole('button', {name: 'Сохранить название'}))
    expect(onRename).toHaveBeenCalledWith(group.id, 'Команда редакторов')
})

test('enables module saving only after the selection changes', () => {
    const onSetModules = vi.fn()
    render(<GroupsView groups={[group]} modules={modules} accesses={[]} error="" onCreate={vi.fn()} onRename={vi.fn()}
                       onDelete={vi.fn()} onAdd={vi.fn()} onRemove={vi.fn()} onSetModules={onSetModules}/>)

    const save = screen.getByRole('button', {name: 'Сохранить доступы'})
    expect(save).toBeDisabled()
    fireEvent.click(screen.getByLabelText('Короткие ссылки'))
    expect(save).toBeEnabled()
    fireEvent.click(save)
    expect(onSetModules).toHaveBeenCalledWith(group.id, [modules[1].id])
})

test('allows the owner to assign a module before adding members', () => {
    const onSetModules = vi.fn()
    render(<GroupsView groups={[{...group, members: []}]} modules={modules} accesses={[]} error="" onCreate={vi.fn()}
                       onRename={vi.fn()}
                       onDelete={vi.fn()} onAdd={vi.fn()} onRemove={vi.fn()} onSetModules={onSetModules}/>)

    expect(screen.getByText('Владелец группы получает доступ автоматически')).toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('Короткие ссылки'))
    const save = screen.getByRole('button', {name: 'Сохранить доступы'})
    expect(save).toBeEnabled()
    fireEvent.click(save)
    expect(onSetModules).toHaveBeenCalledWith(group.id, [modules[1].id])
})

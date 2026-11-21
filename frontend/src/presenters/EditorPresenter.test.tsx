import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react'
import {MemoryRouter, Route, Routes} from 'react-router-dom'
import {afterEach, expect, test, vi} from 'vitest'
import type {Document} from '../model/api'
import {EditorPresenter} from './EditorPresenter'

const mocks = vi.hoisted(() => ({document: vi.fn(), updateDocument: vi.fn(), uploadAttachment: vi.fn()}))
vi.mock('../model/api', async (importOriginal) => {
    const actual = await importOriginal<typeof import('../model/api')>()
    return {...actual, api: mocks}
})

afterEach(() => {
    cleanup();
    vi.clearAllMocks()
})

test('keeps later edits when an earlier save finishes', async () => {
    const original: Document = {
        id: '11111111-1111-4111-8111-111111111111', ownerUserID: '', title: 'Документ',
        content: 'Начало', version: 1, role: 'editor', createdAt: '', updatedAt: '',
    }
    let finishFirst!: (value: Document) => void
    mocks.document.mockResolvedValue(original)
    mocks.updateDocument.mockImplementationOnce(() => new Promise<Document>((resolve) => {
        finishFirst = resolve
    }))
        .mockImplementationOnce(async (value: Document) => ({...value, version: 3}))

    render(<MemoryRouter initialEntries={[`/documents/${original.id}`]}>
        <Routes><Route path="/documents/:id" element={<EditorPresenter/>}/></Routes>
    </MemoryRouter>)

    const editor = await screen.findByRole('textbox', {name: 'Markdown'})
    fireEvent.change(editor, {target: {value: 'Первая правка'}})
    await waitFor(() => expect(mocks.updateDocument).toHaveBeenCalledTimes(1), {timeout: 2000})
    fireEvent.change(editor, {target: {value: 'Вторая правка'}})
    finishFirst({...original, content: 'Первая правка', version: 2})

    await waitFor(() => expect(mocks.updateDocument).toHaveBeenCalledTimes(2), {timeout: 2000})
    expect(mocks.updateDocument.mock.calls[1][0]).toMatchObject({content: 'Вторая правка', version: 2})
    expect(editor).toHaveValue('Вторая правка')
})

import {useState} from 'react'
import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react'
import {afterEach, expect, test, vi} from 'vitest'
import type {Attachment, Document} from '../model/api'
import {EditorView} from './EditorView'

const baseDocument: Document = {
    id: '11111111-1111-4111-8111-111111111111', ownerUserID: '', title: 'Документ',
    content: 'До после', version: 1, role: 'editor', createdAt: '', updatedAt: '',
}

function attachment(file: File, fileID: string): Attachment {
    return {FileID: fileID, FileName: file.name, ContentType: file.type, SizeBytes: file.size}
}

function Harness({onUpload, role = 'editor'}: {
    onUpload: (file: File) => Promise<Attachment>;
    role?: Document['role']
}) {
    const [document, setDocument] = useState({...baseDocument, role})
    return <EditorView document={document} status="" sharing={false} onBack={() => {
    }}
                       onChange={(patch) => setDocument((value) => ({...value, ...patch}))} onUpload={onUpload}
                       onOpenSharing={() => {
                       }} onCloseSharing={() => {
    }}/>
}

afterEach(() => cleanup())

test('menu uploads an image and keeps edits made while uploading', async () => {
    let finishUpload!: (value: Attachment) => void
    const onUpload = vi.fn(() => new Promise<Attachment>((resolve) => {
        finishUpload = resolve
    }))
    render(<Harness onUpload={onUpload}/>)
    const editor = screen.getByRole('textbox', {name: 'Markdown'}) as HTMLTextAreaElement
    editor.setSelectionRange(3, 3)
    fireEvent.click(screen.getByText('Вставить'))
    fireEvent.click(screen.getByRole('button', {name: 'Изображение'}))
    const file = new File(['image'], 'photo.png', {type: 'image/png'})
    fireEvent.change(screen.getByLabelText('Выбрать изображения'), {target: {files: [file]}})
    expect(onUpload).toHaveBeenCalledWith(file)
    fireEvent.change(editor, {target: {value: 'До красивое после'}})
    finishUpload(attachment(file, 'file-1'))
    await waitFor(() => expect(editor.value).toBe('До ![photo.png](/api/editor/v1/documents/11111111-1111-4111-8111-111111111111/attachments/file-1/content)красивое после'))
    expect(screen.getByRole('img', {name: 'photo.png'})).toBeInTheDocument()
})

test('drop uploads multiple files and inserts links in order', async () => {
    const onUpload = vi.fn(async (file: File) => attachment(file, file.name))
    render(<Harness onUpload={onUpload}/>)
    const editor = screen.getByRole('textbox', {name: 'Markdown'}) as HTMLTextAreaElement
    editor.setSelectionRange(3, 3)
    const files = [new File(['a'], 'first.txt', {type: 'text/plain'}), new File(['b'], 'second.svg', {type: 'image/svg+xml'})]
    fireEvent.drop(editor, {dataTransfer: {files, types: ['Files']}})
    await waitFor(() => expect(onUpload).toHaveBeenCalledTimes(2))
    await waitFor(() => expect(editor.value).toContain('[first.txt](/api/editor/v1/documents/11111111-1111-4111-8111-111111111111/attachments/first.txt/content)\n[second.svg]'))
    expect(editor.value).not.toContain('![second.svg]')
})

test('failed upload leaves the document unchanged and viewers cannot upload', async () => {
    const onUpload = vi.fn().mockRejectedValue(new Error('Ошибка сети'))
    render(<Harness onUpload={onUpload}/>)
    const editor = screen.getByRole('textbox', {name: 'Markdown'}) as HTMLTextAreaElement
    fireEvent.drop(editor, {dataTransfer: {files: [new File(['a'], 'a.txt')], types: ['Files']}})
    expect(await screen.findByRole('status')).toHaveTextContent('Ошибка сети')
    expect(editor.value).toBe('До после')
    cleanup()
    render(<Harness onUpload={onUpload} role="viewer"/>)
    expect(screen.queryByText('Вставить')).not.toBeInTheDocument()
    const readOnlyEditor = screen.getByRole('textbox', {name: 'Markdown'})
    expect(readOnlyEditor).toHaveAttribute('readonly')
    fireEvent.drop(readOnlyEditor, {dataTransfer: {files: [new File(['b'], 'b.txt')], types: ['Files']}})
    expect(onUpload).toHaveBeenCalledTimes(1)
})

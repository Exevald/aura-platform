import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react'
import {afterEach, expect, test, vi} from 'vitest'
import {api} from '../model/api'
import {AttachmentImage, AttachmentLink} from './AttachmentPreview'

const documentID = '11111111-1111-4111-8111-111111111111'
const fileID = '22222222-2222-4222-8222-222222222222'
const path = `/api/editor/v1/documents/${documentID}/attachments/${fileID}/content`

afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    Reflect.deleteProperty(URL, 'createObjectURL')
    Reflect.deleteProperty(URL, 'revokeObjectURL')
})

test('download request includes the selected workspace', async () => {
    const file = new Blob(['image'], {type: 'image/png'})
    const fetchMock = vi.fn().mockResolvedValue({ok: true, status: 200, blob: async () => file})
    vi.stubGlobal('fetch', fetchMock)
    vi.stubGlobal('localStorage', {getItem: () => '33333333-3333-4333-8333-333333333333'})

    expect(await api.downloadAttachment(documentID, fileID)).toBe(file)
    expect(fetchMock).toHaveBeenCalledWith(path, expect.objectContaining({
        credentials: 'include',
        headers: expect.objectContaining({'X-Workspace-ID': '33333333-3333-4333-8333-333333333333'}),
    }))
})

test('loads an editor image through the authenticated API', async () => {
    const download = vi.spyOn(api, 'downloadAttachment').mockResolvedValue(new Blob(['image'], {type: 'image/png'}))
    Object.defineProperty(URL, 'createObjectURL', {configurable: true, value: vi.fn(() => 'blob:preview')})
    Object.defineProperty(URL, 'revokeObjectURL', {configurable: true, value: vi.fn()})

    const {unmount} = render(<AttachmentImage src={path} alt="photo.png"/>)
    await waitFor(() => expect(screen.getByRole('img', {name: 'photo.png'})).toHaveAttribute('src', 'blob:preview'))
    expect(download).toHaveBeenCalledWith(documentID, fileID)
    unmount()
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:preview')
})

test('downloads an attachment link through the authenticated API', async () => {
    const download = vi.spyOn(api, 'downloadAttachment').mockResolvedValue(new Blob(['file']))
    Object.defineProperty(URL, 'createObjectURL', {configurable: true, value: vi.fn(() => 'blob:download')})
    Object.defineProperty(URL, 'revokeObjectURL', {configurable: true, value: vi.fn()})
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {
    })

    render(<AttachmentLink href={path}>notes.txt</AttachmentLink>)
    fireEvent.click(screen.getByRole('link', {name: 'notes.txt'}))
    await waitFor(() => expect(click).toHaveBeenCalledTimes(1))
    expect(download).toHaveBeenCalledWith(documentID, fileID)
})

import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react'
import {afterEach, expect, test, vi} from 'vitest'
import {PublicationPresenter} from './PublicationPresenter'

const mocks = vi.hoisted(() => ({
    documents: vi.fn(),
    publications: vi.fn(),
    publishDocument: vi.fn(),
}))

vi.mock('../model/api', () => ({api: mocks}))

const execCommandDescriptor = Object.getOwnPropertyDescriptor(document, 'execCommand')

afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
    vi.clearAllMocks()
    if (execCommandDescriptor) {
        Object.defineProperty(document, 'execCommand', execCommandDescriptor)
    } else {
        Reflect.deleteProperty(document, 'execCommand')
    }
})

test('publishes on an HTTP origin without crypto.randomUUID', async () => {
    const document = {id: '11111111-1111-4111-8111-111111111111', title: 'Документ', version: 1, role: 'owner'}
    mocks.documents.mockResolvedValue([document])
    mocks.publications.mockResolvedValue([])
    mocks.publishDocument.mockResolvedValue({
        id: '22222222-2222-4222-8222-222222222222',
        documentID: document.id,
        status: 'pending'
    })
    vi.stubGlobal('crypto', {
        getRandomValues: (bytes: Uint8Array) => bytes.fill(7),
        randomUUID: undefined,
    })

    render(<PublicationPresenter/>)
    fireEvent.change(await screen.findByLabelText('Документ'), {target: {value: document.id}})
    fireEvent.click(screen.getByRole('button', {name: 'Опубликовать'}))

    await waitFor(() => expect(mocks.publishDocument).toHaveBeenCalledWith(document, '07'.repeat(16)))
    expect(await screen.findByText('Публикуется')).toBeInTheDocument()
})

test('copies a publication link when Clipboard API is unavailable', async () => {
    const release = {
        id: '22222222-2222-4222-8222-222222222222',
        documentID: '11111111-1111-4111-8111-111111111111',
        title: 'Документ',
        status: 'active',
        shortPath: '/s/published',
    }
    mocks.documents.mockResolvedValue([])
    mocks.publications.mockResolvedValue([release])
    vi.stubGlobal('navigator', {clipboard: undefined})

    let copiedValue = ''
    Object.defineProperty(document, 'execCommand', {
        configurable: true,
        value: vi.fn(() => {
            copiedValue = document.querySelector('textarea')?.value ?? ''
            return true
        }),
    })

    render(<PublicationPresenter/>)
    fireEvent.click(await screen.findByRole('button', {name: 'Копировать'}))

    await waitFor(() => expect(screen.getByRole('button', {name: 'Скопировано'})).toBeInTheDocument())
    expect(copiedValue).toBe(new URL(release.shortPath, location.origin).toString())
    expect(document.querySelector('textarea')).toBeNull()
})

test('copies a publication link when Clipboard API rejects access', async () => {
    mocks.documents.mockResolvedValue([])
    mocks.publications.mockResolvedValue([{
        id: '22222222-2222-4222-8222-222222222222',
        documentID: '11111111-1111-4111-8111-111111111111',
        status: 'active',
        shortPath: '/s/published',
    }])
    vi.stubGlobal('navigator', {clipboard: {writeText: vi.fn().mockRejectedValue(new Error('denied'))}})
    const fallback = vi.fn().mockReturnValue(true)
    Object.defineProperty(document, 'execCommand', {configurable: true, value: fallback})

    render(<PublicationPresenter/>)
    fireEvent.click(await screen.findByRole('button', {name: 'Копировать'}))

    await waitFor(() => expect(screen.getByRole('button', {name: 'Скопировано'})).toBeInTheDocument())
    expect(fallback).toHaveBeenCalledWith('copy')
})

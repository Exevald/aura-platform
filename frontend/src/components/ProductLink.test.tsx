import {cleanup, fireEvent, render, screen} from '@testing-library/react'
import {MemoryRouter, useLocation} from 'react-router-dom'
import {afterEach, expect, test, vi} from 'vitest'
import {ProductLink} from './ProductLink'

afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
})

test('navigates inside the current product without reloading the page', () => {
    vi.stubGlobal('location', {hostname: 'aura-platform.home', protocol: 'http:', port: ''})

    function CurrentPath() {
        return <span>{useLocation().pathname}</span>
    }

    render(<MemoryRouter initialEntries={['/overview']}>
        <ProductLink to="/groups">Группы</ProductLink>
        <CurrentPath/>
    </MemoryRouter>)

    fireEvent.click(screen.getByRole('link', {name: 'Группы'}))
    expect(screen.getByText('/groups')).toBeInTheDocument()
})

test('uses a full navigation when switching products', () => {
    vi.stubGlobal('location', {hostname: 'aura-platform.home', protocol: 'http:', port: ''})
    render(<MemoryRouter><ProductLink to="/documents">Документы</ProductLink></MemoryRouter>)
    expect(screen.getByRole('link', {name: 'Документы'})).toHaveAttribute('href', 'http://aura-platform.documents/documents')
})

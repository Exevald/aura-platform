import type {AnchorHTMLAttributes, ReactNode} from 'react'
import {Link} from 'react-router-dom'
import {isAuraHost, navigationURL, productForPath, productHost} from '../model/hosts'

type Props = Pick<AnchorHTMLAttributes<HTMLAnchorElement>, 'className' | 'onClick' | 'aria-label'> & {
    to: string
    children: ReactNode
}

export function ProductLink({to, children, ...props}: Props) {
    if (!isAuraHost() || productHost(productForPath(to)) === location.hostname) {
        return <Link to={to} {...props}>{children}</Link>
    }
    return <a href={navigationURL(to)} {...props}>{children}</a>
}

import {FileText} from 'lucide-react'

type Props = { title: string; detail?: string; action?: string; onAction?: () => void }

export function StateView({title, detail, action, onAction}: Props) {
    return <div className="state">
        <div className="state-icon"><FileText size={27}/></div>
        <h2>{title}</h2>{detail && <p>{detail}</p>}{action &&
        <button className="primary" onClick={onAction}>{action}</button>}</div>
}

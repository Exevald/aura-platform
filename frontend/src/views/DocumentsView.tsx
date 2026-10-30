import type { Document } from '../model/api'
import { StateView } from './StateView'

type Props = {
  documents: Document[] | null
  error: string
  onCreate: () => void
  onOpen: (id: string) => void
  onRemove: (document: Document) => void
}

export function DocumentsView({ documents, error, onCreate, onOpen, onRemove }: Props) {
  if (error) return <StateView title={error}/>
  return <><div className="page-title"><div><h1>Документы</h1><p>Ваши файлы и документы, которыми поделились с вами</p></div><button className="primary" onClick={onCreate}>＋ Создать</button></div>
    {!documents ? <StateView title="Загружаем документы…"/> : documents.length === 0 ? <StateView title="Здесь пока пусто" detail="Создайте первый Markdown-документ" action="Создать документ" onAction={onCreate}/> : <div className="document-grid">{documents.map((document) => <article className="document-card" key={document.id}><button className="card-open" onClick={() => onOpen(document.id)}><div className="doc-preview"><strong>#</strong><span>{document.content.slice(0, 120) || 'Пустой документ'}</span></div><h3>{document.title}</h3><p>{document.role === 'owner' ? 'Мой документ' : document.role === 'editor' ? 'Можно редактировать' : 'Только просмотр'} · {new Date(document.updatedAt).toLocaleDateString('ru')}</p></button>{document.role === 'owner' && <button className="card-menu" aria-label="Удалить" onClick={() => onRemove(document)}>⋮</button>}</article>)}</div>}
  </>
}

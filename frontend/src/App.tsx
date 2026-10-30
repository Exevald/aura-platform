import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthPresenter } from './presenters/AuthPresenter'
import { DocumentsPresenter, NewDocumentPresenter } from './presenters/DocumentsPresenter'
import { EditorPresenter } from './presenters/EditorPresenter'
import { ProfilePresenter } from './presenters/ProfilePresenter'
import { ShellPresenter } from './presenters/ShellPresenter'

export default function App() {
  return <Routes>
    <Route path="/login" element={<AuthPresenter/>}/>
    <Route element={<ShellPresenter/>}>
      <Route path="/documents" element={<DocumentsPresenter/>}/>
      <Route path="/documents/new" element={<NewDocumentPresenter/>}/>
      <Route path="/documents/:id" element={<EditorPresenter/>}/>
      <Route path="/profile" element={<ProfilePresenter/>}/>
    </Route>
    <Route path="*" element={<Navigate to="/documents" replace/>}/>
  </Routes>
}

import {useEffect, useState} from 'react'
import {api, type Project, type Task} from '../model/api'
import {PlannerView} from '../views/ProductViews'

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : 'Ошибка загрузки'
}

export function PlannerPresenter() {
  const [projects, setProjects] = useState<Project[]>([])
  const [selected, setSelected] = useState('')
  const [tasks, setTasks] = useState<Task[]>([])
  const [error, setError] = useState('')

  const refresh = async () => {
    try {
      const values = await api.projects()
      setProjects(values)
      setSelected(current => current || values[0]?.id || '')
    } catch (cause) {
      setError(errorMessage(cause))
    }
  }

  useEffect(() => {
    void api.projects()
      .then(values => {
        setProjects(values)
        setSelected(current => current || values[0]?.id || '')
      })
      .catch(cause => setError(errorMessage(cause)))
  }, [])

  useEffect(() => {
    if (!selected) return
    void api.tasks(selected)
      .then(setTasks)
      .catch(cause => setError(errorMessage(cause)))
  }, [selected])

  return <PlannerView
    projects={projects}
    selected={selected}
    tasks={tasks}
    error={error}
    onSelect={setSelected}
    onCreateProject={async name => {
      const project = await api.createProject(name)
      setProjects(current => [...current, project])
      setSelected(project.id)
    }}
    onDeleteProject={async id => {
      await api.deleteProject(id)
      setSelected('')
      await refresh()
    }}
    onCreateTask={async title => {
      const task = await api.createTask(selected, {
        title,
        description: '',
        status: 'todo',
        priority: 'normal',
        syncToCalendar: false,
      })
      setTasks(current => [...current, task])
    }}
    onToggleTask={async task => {
      const updated = await api.updateTask({...task, status: task.status === 'done' ? 'todo' : 'done'})
      setTasks(current => current.map(item => item.id === updated.id ? updated : item))
    }}
    onDeleteTask={async id => {
      await api.deleteTask(id)
      setTasks(current => current.filter(task => task.id !== id))
    }}
  />
}

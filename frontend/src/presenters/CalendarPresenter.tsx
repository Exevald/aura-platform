import {useEffect, useState} from 'react'
import {api, type Calendar, type CalendarEvent} from '../model/api'
import {CalendarView} from '../views/ProductViews'

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : 'Ошибка загрузки'
}

export function CalendarPresenter() {
  const [calendars, setCalendars] = useState<Calendar[]>([])
  const [selected, setSelected] = useState('')
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [error, setError] = useState('')

  const refresh = async () => {
    try {
      const values = await api.calendars()
      setCalendars(values)
      setSelected(current => current || values[0]?.id || '')
      setError('')
    } catch (cause) {
      setError(errorMessage(cause))
    }
  }

  useEffect(() => {
    void api.calendars()
      .then(values => {
        setCalendars(values)
        setSelected(current => current || values[0]?.id || '')
        setError('')
      })
      .catch(cause => setError(errorMessage(cause)))
  }, [])

  useEffect(() => {
    if (!selected) return
    void api.calendarEvents(selected)
      .then(setEvents)
      .catch(cause => setError(errorMessage(cause)))
  }, [selected])

  return <CalendarView
    calendars={calendars}
    selected={selected}
    events={events}
    error={error}
    onSelect={setSelected}
    onCreateCalendar={async name => {
      const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
      const calendar = await api.createCalendar(name, timezone)
      setCalendars(current => [...current, calendar])
      setSelected(calendar.id)
    }}
    onDeleteCalendar={async id => {
      await api.deleteCalendar(id)
      setSelected('')
      await refresh()
    }}
    onCreateEvent={async value => {
      const event = await api.createCalendarEvent(selected, value)
      setEvents(current => [...current, event])
    }}
    onDeleteEvent={async id => {
      await api.deleteCalendarEvent(id)
      setEvents(current => current.filter(event => event.id !== id))
    }}
  />
}

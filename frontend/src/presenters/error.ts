export function errorMessage(error: unknown) {
  if (!(error instanceof Error)) return 'Что-то пошло не так'
  if (error.message === 'module access denied') return 'У вас пока нет доступа к этому модулю. Попросите владельца группы включить его в настройках доступов.'
  if (error.message === 'authentication required') return 'Сессия завершилась. Войдите снова.'
  return error.message
}

export async function copyText(value: string): Promise<void> {
    try {
        if (navigator.clipboard?.writeText) {
            await navigator.clipboard.writeText(value)
            return
        }
    } catch {
        // The browser may reject Clipboard API access even when it is present.
    }

    const field = document.createElement('textarea')
    field.value = value
    field.readOnly = true
    field.style.position = 'fixed'
    field.style.opacity = '0'
    document.body.append(field)

    try {
        field.select()
        if (!document.execCommand('copy')) throw new Error('copy failed')
    } finally {
        field.remove()
    }
}

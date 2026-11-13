export type SessionEvent = 'authenticated' | 'logged-out'

type SessionListener = (event: SessionEvent) => void
type SessionMessage = { type: SessionEvent }

export class SessionCoordinator {
    private readonly listeners = new Set<SessionListener>()
    private readonly onMessage = (event: MessageEvent<SessionMessage>) => {
        if (event.data?.type === 'authenticated' || event.data?.type === 'logged-out') {
            this.notify(event.data.type)
        }
    }

    constructor(private readonly channel?: BroadcastChannel) {
        channel?.addEventListener('message', this.onMessage)
    }

    publish(event: SessionEvent) {
        this.notify(event)
        this.channel?.postMessage({type: event} satisfies SessionMessage)
    }

    subscribe(listener: SessionListener) {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    destroy() {
        this.channel?.removeEventListener('message', this.onMessage)
        this.channel?.close()
        this.listeners.clear()
    }

    private notify(event: SessionEvent) {
        for (const listener of this.listeners) listener(event)
    }
}

function openBrowserChannel() {
    return typeof BroadcastChannel === 'undefined' ? undefined : new BroadcastChannel('aura-session')
}

export const sessionCoordinator = new SessionCoordinator(openBrowserChannel())

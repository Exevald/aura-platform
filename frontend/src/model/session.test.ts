import { expect, test, vi } from 'vitest'
import { SessionCoordinator, type SessionEvent } from './session'

class FakeBroadcastChannel {
  sent: unknown[] = []
  listener?: (event: MessageEvent<{ type: SessionEvent }>) => void
  close = vi.fn()

  addEventListener(_type: string, listener: EventListenerOrEventListenerObject) {
    this.listener = listener as (event: MessageEvent<{ type: SessionEvent }>) => void
  }

  removeEventListener() {
    this.listener = undefined
  }

  postMessage(message: unknown) {
    this.sent.push(message)
  }

  receive(type: SessionEvent) {
    this.listener?.({ data: { type } } as MessageEvent<{ type: SessionEvent }>)
  }
}

test('publishes locally and to other tabs without exposing a session id', () => {
  const channel = new FakeBroadcastChannel()
  const coordinator = new SessionCoordinator(channel as unknown as BroadcastChannel)
  const listener = vi.fn()
  coordinator.subscribe(listener)

  coordinator.publish('authenticated')

  expect(listener).toHaveBeenCalledWith('authenticated')
  expect(channel.sent).toEqual([{ type: 'authenticated' }])
  expect(JSON.stringify(channel.sent)).not.toContain('session')
  coordinator.destroy()
})

test('delivers logout received from another tab', () => {
  const channel = new FakeBroadcastChannel()
  const coordinator = new SessionCoordinator(channel as unknown as BroadcastChannel)
  const listener = vi.fn()
  coordinator.subscribe(listener)

  channel.receive('logged-out')

  expect(listener).toHaveBeenCalledWith('logged-out')
  coordinator.destroy()
})

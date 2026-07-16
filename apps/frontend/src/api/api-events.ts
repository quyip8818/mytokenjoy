/**
 * Lightweight typed event bus for API client lifecycle events.
 * Replaces global mutable handler setters with a subscribe/emit pattern.
 *
 * Benefits:
 * - Multiple subscribers allowed (no handler overwriting)
 * - No mount-order dependency
 * - Subscribers auto-cleanup via returned unsubscribe fn
 */

type UnauthorizedListener = () => void
type ForbiddenListener = (path: string) => void
type AuthzRevisionListener = (revision: number) => void

interface ApiEventMap {
  unauthorized: UnauthorizedListener
  forbidden: ForbiddenListener
  authzRevision: AuthzRevisionListener
}

type ListenerFor<K extends keyof ApiEventMap> = ApiEventMap[K]

class ApiEventBus {
  private listeners: { [K in keyof ApiEventMap]: Set<ListenerFor<K>> } = {
    unauthorized: new Set(),
    forbidden: new Set(),
    authzRevision: new Set(),
  }

  on<K extends keyof ApiEventMap>(event: K, listener: ListenerFor<K>): () => void {
    const set = this.listeners[event] as Set<ListenerFor<K>>
    set.add(listener)
    return () => {
      set.delete(listener)
    }
  }

  emit<K extends keyof ApiEventMap>(event: K, ...args: Parameters<ListenerFor<K>>): void {
    const set = this.listeners[event] as Set<(...a: Parameters<ListenerFor<K>>) => void>
    for (const listener of set) {
      listener(...args)
    }
  }
}

export const apiEvents = new ApiEventBus()

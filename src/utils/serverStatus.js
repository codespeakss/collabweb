// Simple connection status monitor using SSE and browser online/offline events
// Exposes: startServerStatusMonitor, stopServerStatusMonitor, onStatusChange, getStatus

const state = {
  connected: false, // SSE connected
  online: typeof navigator !== 'undefined' ? navigator.onLine : true,
  lastEventTs: 0,
}

const listeners = new Set()
let es = null
let retryTimer = null
let retryDelay = 2000 // start with 2s, up to 30s

function notify() {
  for (const cb of listeners) {
    try { cb({ ...state }) } catch (_) {}
  }
}

function scheduleReconnect() {
  if (retryTimer) return
  retryTimer = setTimeout(() => {
    retryTimer = null
    // exponential backoff with cap
    retryDelay = Math.min(retryDelay * 1.6, 30000)
    openSSE()
  }, retryDelay)
}

function openSSE() {
  // Close previous if any
  if (es) {
    try { es.close() } catch (_) {}
    es = null
  }
  try {
    es = new EventSource('/api/v1/health/stream')
  } catch (_) {
    state.connected = false
    notify()
    scheduleReconnect()
    return
  }

  es.addEventListener('open', () => {
    state.connected = true
    state.lastEventTs = Date.now()
    retryDelay = 2000 // reset backoff on success
    notify()
  })

  // custom ping events
  es.addEventListener('ping', () => {
    state.connected = true
    state.lastEventTs = Date.now()
    notify()
  })

  // any message is fine
  es.onmessage = () => {
    state.connected = true
    state.lastEventTs = Date.now()
    notify()
  }

  es.onerror = () => {
    state.connected = false
    notify()
    // browser will auto-reconnect; but we also attempt explicit reopen with backoff
    try { es.close() } catch (_) {}
    es = null
    scheduleReconnect()
  }
}

function handleOnline() {
  state.online = true
  notify()
  // try reconnect immediately when back online
  retryDelay = 2000
  openSSE()
}
function handleOffline() {
  state.online = false
  state.connected = false
  notify()
}

export function startServerStatusMonitor() {
  if (typeof window === 'undefined') return
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
  openSSE()
}

export function stopServerStatusMonitor() {
  if (typeof window === 'undefined') return
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
  if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
  if (es) {
    try { es.close() } catch (_) {}
    es = null
  }
}

export function onStatusChange(cb) {
  if (typeof cb === 'function') listeners.add(cb)
  return () => listeners.delete(cb)
}

export function getStatus() {
  return { ...state }
}

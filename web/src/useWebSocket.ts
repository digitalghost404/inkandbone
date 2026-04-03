import { useEffect, useRef } from 'react'

export function useWebSocket(url: string, onMessage: (data: unknown) => void): void {
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  useEffect(() => {
    let ws: WebSocket
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let cancelled = false

    function connect() {
      ws = new WebSocket(url)

      ws.onmessage = (e) => {
        try {
          onMessageRef.current(JSON.parse(e.data as string))
        } catch {
          // ignore malformed messages
        }
      }

      ws.onclose = () => {
        if (!cancelled) {
          reconnectTimer = setTimeout(connect, 2000)
        }
      }
    }

    connect()

    return () => {
      cancelled = true
      if (reconnectTimer !== null) clearTimeout(reconnectTimer)
      ws.close()
    }
  }, [url])
}

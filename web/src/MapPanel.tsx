import { useEffect, useState } from 'react'
import type { CampaignMap, MapPin } from './api'
import { fetchMaps, fetchMapPins } from './api'

function isMapPinAddedEvent(e: unknown): e is { type: string; payload: { map_id: number } } {
  return (
    typeof e === 'object' &&
    e !== null &&
    (e as Record<string, unknown>)['type'] === 'map_pin_added' &&
    typeof (e as Record<string, unknown>)['payload'] === 'object' &&
    (e as Record<string, { map_id: unknown }>)['payload']['map_id'] !== undefined
  )
}

function isMapCreatedEvent(e: unknown): boolean {
  return typeof e === 'object' && e !== null && (e as Record<string, unknown>)['type'] === 'map_created'
}

interface MapPanelProps {
  campaignId: number | null
  lastEvent: unknown
  onActiveMapChange?: (mapId: number | null, imagePath: string | null) => void
}

export function MapPanel({ campaignId, lastEvent, onActiveMapChange }: MapPanelProps) {
  const [maps, setMaps] = useState<CampaignMap[]>([])
  const [activeMapIdx, setActiveMapIdx] = useState(0)
  const [pins, setPins] = useState<MapPin[]>([])
  const [selectedPin, setSelectedPin] = useState<MapPin | null>(null)

  function loadMaps(goToLast = false) {
    if (campaignId === null) return
    fetchMaps(campaignId).then((m) => {
      setMaps(m)
      if (goToLast && m.length > 0) {
        setActiveMapIdx(m.length - 1)
      }
    }).catch(console.error)
  }

  useEffect(() => {
    setSelectedPin(null)
    setActiveMapIdx(0)
    loadMaps()
  }, [campaignId]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (isMapCreatedEvent(lastEvent)) {
      loadMaps(true)
    }
  }, [lastEvent]) // eslint-disable-line react-hooks/exhaustive-deps

  const activeMap = maps[activeMapIdx] ?? null

  useEffect(() => {
    if (!activeMap) {
      setPins([])
      onActiveMapChange?.(null, null)
      return
    }
    fetchMapPins(activeMap.id).then(setPins).catch(console.error)
    onActiveMapChange?.(activeMap.id, activeMap.image_path)
  }, [activeMap?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (isMapPinAddedEvent(lastEvent) && activeMap && (lastEvent as { payload: { map_id: number } }).payload.map_id === activeMap.id) {
      fetchMapPins(activeMap.id).then(setPins).catch(console.error)
    }
  }, [lastEvent, activeMap?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  if (campaignId === null) return null

  if (maps.length === 0) {
    return <p>No map uploaded.</p>
  }

  return (
    <div style={{ position: 'relative', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {maps.length > 1 && (
        <div className="map-tab-bar">
          {maps.map((m, i) => (
            <button
              key={m.id}
              className={`map-tab-btn${i === activeMapIdx ? ' active' : ''}`}
              onClick={() => { setActiveMapIdx(i); setSelectedPin(null) }}
            >
              {m.name}
            </button>
          ))}
        </div>
      )}
      {activeMap && (
        <div style={{ position: 'relative', flex: 1 }}>
          <img
            src={`/api/files/${activeMap.image_path}`}
            alt={activeMap.name}
            style={{ width: '100%', display: 'block' }}
          />
          {pins.map((pin) => (
            <button
              key={pin.id}
              className="map-pin-btn"
              style={{
                position: 'absolute',
                left: `${pin.x * 100}%`,
                top: `${pin.y * 100}%`,
                transform: 'translate(-50%,-50%)',
                background: pin.color || 'var(--gold)',
                color: '#000',
                border: 'none',
                borderRadius: '50%',
                width: '18px',
                height: '18px',
                fontSize: '9px',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontWeight: 700,
              }}
              title={pin.note || pin.label}
              onClick={() => setSelectedPin(selectedPin?.id === pin.id ? null : pin)}
            >
              ✦
            </button>
          ))}
          {selectedPin && (
            <div className="map-pin-tooltip">
              <strong>{selectedPin.label}</strong>
              {selectedPin.note && <p>{selectedPin.note}</p>}
              <button className="map-pin-tooltip-close" onClick={() => setSelectedPin(null)}>×</button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

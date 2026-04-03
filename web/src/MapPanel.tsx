import { useEffect, useState } from 'react'
import { CampaignMap, MapPin, fetchMaps, fetchMapPins } from './api'

function isMapPinAddedEvent(e: unknown): e is { type: string; payload: { map_id: number } } {
  return (
    typeof e === 'object' &&
    e !== null &&
    (e as any).type === 'map_pin_added' &&
    typeof (e as any).payload?.map_id === 'number'
  )
}

interface MapPanelProps {
  campaignId: number | null
  lastEvent: unknown
}

export function MapPanel({ campaignId, lastEvent }: MapPanelProps) {
  const [map, setMap] = useState<CampaignMap | null>(null)
  const [pins, setPins] = useState<MapPin[]>([])
  const [selectedPin, setSelectedPin] = useState<MapPin | null>(null)

  useEffect(() => {
    setSelectedPin(null)
    if (campaignId === null) return
    fetchMaps(campaignId).then(maps => setMap(maps[0] ?? null)).catch(console.error)
  }, [campaignId])

  useEffect(() => {
    if (!map) return
    fetchMapPins(map.id).then(setPins).catch(console.error)
  }, [map])

  useEffect(() => {
    if (isMapPinAddedEvent(lastEvent) && map && lastEvent.payload.map_id === map.id) {
      fetchMapPins(map.id).then(setPins).catch(console.error)
    }
  }, [lastEvent, map])

  if (campaignId === null) return null

  if (!map) {
    return <p>No map uploaded.</p>
  }

  return (
    <div style={{ position: 'relative' }}>
      <img
        src={`/api/files/${map.image_path}`}
        alt="Campaign map"
        style={{ width: '100%', display: 'block' }}
      />
      {pins.map((pin) => (
        <button
          key={pin.id}
          style={{
            position: 'absolute',
            left: `${pin.x * 100}%`,
            top: `${pin.y * 100}%`,
            transform: 'translate(-50%,-50%)',
          }}
          title={pin.note || pin.label}
          onClick={() => setSelectedPin(selectedPin?.id === pin.id ? null : pin)}
        >
          {pin.label}
        </button>
      ))}
      {selectedPin && (
        <div>
          <strong>{selectedPin.label}</strong>
          {selectedPin.note && <p>{selectedPin.note}</p>}
        </div>
      )}
    </div>
  )
}

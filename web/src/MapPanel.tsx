import { useEffect, useState } from 'react'
import { CampaignMap, MapPin, fetchMaps, fetchMapPins } from './api'

interface MapPanelProps {
  campaignId: number | null
  lastEvent: unknown
}

export function MapPanel({ campaignId, lastEvent }: MapPanelProps) {
  const [map, setMap] = useState<CampaignMap | null>(null)
  const [pins, setPins] = useState<MapPin[]>([])
  const [selectedPin, setSelectedPin] = useState<MapPin | null>(null)

  useEffect(() => {
    if (campaignId === null) return
    fetchMaps(campaignId).then((maps) => {
      if (maps.length > 0) {
        setMap(maps[0])
      } else {
        setMap(null)
      }
    })
  }, [campaignId])

  useEffect(() => {
    if (!map) return
    fetchMapPins(map.id).then(setPins)
  }, [map])

  useEffect(() => {
    if (!lastEvent || !map) return
    const event = lastEvent as { type: string; payload: { map_id: number } }
    if (event.type === 'map_pin_added' && event.payload.map_id === map.id) {
      fetchMapPins(map.id).then(setPins)
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

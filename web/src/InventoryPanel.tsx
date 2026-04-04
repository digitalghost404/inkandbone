import { useState, useEffect } from 'react'
import { fetchItems, createItem, patchItem, deleteItem } from './api'
import type { Item } from './types'

interface InventoryPanelProps {
  characterId: number | null
  lastEvent: unknown
}

export function InventoryPanel({ characterId, lastEvent }: InventoryPanelProps) {
  const [items, setItems] = useState<Item[]>([])
  const [addName, setAddName] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (characterId === null) return
    fetchItems(characterId).then(setItems).catch(() => setItems([]))
  }, [characterId])

  useEffect(() => {
    const ev = lastEvent as { type?: string } | null
    if (ev?.type === 'item_updated' && characterId !== null) {
      fetchItems(characterId).then(setItems).catch(() => {})
    }
  }, [lastEvent, characterId])

  async function handleAdd() {
    if (!addName.trim() || characterId === null) return
    setSaving(true)
    try {
      const item = await createItem(characterId, addName.trim(), '', 1)
      setItems((prev) => [...prev, item])
      setAddName('')
    } catch (err) {
      console.error(err)
    } finally {
      setSaving(false)
    }
  }

  async function handleEquipToggle(item: Item) {
    try {
      await patchItem(item.id, { equipped: !item.equipped })
      setItems((prev) =>
        prev.map((i) => (i.id === item.id ? { ...i, equipped: !item.equipped } : i))
      )
    } catch (err) {
      console.error(err)
    }
  }

  async function handleDelete(id: number) {
    try {
      await deleteItem(id)
      setItems((prev) => prev.filter((i) => i.id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  if (characterId === null) return null

  const sorted = [
    ...items.filter((i) => i.equipped),
    ...items.filter((i) => !i.equipped),
  ]

  return (
    <div className="inventory-panel">
      <div className="inventory-panel-label">Inventory</div>

      {sorted.length === 0 && (
        <p className="empty">No items yet.</p>
      )}

      {sorted.map((item) => (
        <div key={item.id} className={`inventory-item${item.equipped ? ' equipped' : ''}`}>
          <div className="inventory-item-row">
            {item.equipped && <span className="inventory-equipped-icon" title="Equipped">⚔</span>}
            <span className="inventory-item-name">{item.name}</span>
            {item.quantity > 1 && (
              <span className="inventory-item-qty">×{item.quantity}</span>
            )}
            <div className="inventory-actions">
              <button
                className={`inventory-equip-btn${item.equipped ? ' active' : ''}`}
                onClick={() => handleEquipToggle(item)}
                title={item.equipped ? 'Unequip' : 'Equip'}
              >
                {item.equipped ? '▣' : '□'}
              </button>
              <button
                className="inventory-delete-btn"
                onClick={() => handleDelete(item.id)}
                title="Remove item"
              >
                ×
              </button>
            </div>
          </div>
          {item.description && (
            <div className="inventory-item-desc">{item.description}</div>
          )}
        </div>
      ))}

      <div className="inventory-add-form">
        <input
          className="inventory-add-input"
          placeholder="Item name…"
          value={addName}
          onChange={(e) => setAddName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleAdd()
          }}
        />
        <button
          className="inventory-add-btn"
          onClick={handleAdd}
          disabled={saving || !addName.trim()}
        >
          {saving ? '…' : 'Add'}
        </button>
      </div>
    </div>
  )
}

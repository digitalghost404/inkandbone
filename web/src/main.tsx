import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <div style={{ fontFamily: 'monospace', padding: '2rem' }}>
      <h1>TTRPG Companion</h1>
      <p>Frontend coming in Plan 3.</p>
    </div>
  </StrictMode>
)

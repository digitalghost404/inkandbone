import { useState, useEffect } from 'react';
import { setAmbientVolume, setAmbientMuted } from './audio/ambient';

const STORAGE_KEY_MUTED = 'inkandbone_audio_muted';
const STORAGE_KEY_VOLUME = 'inkandbone_audio_volume';

// Exported state accessor for sounds.ts and other consumers
export function getAudioMuted(): boolean {
  return localStorage.getItem(STORAGE_KEY_MUTED) === 'true';
}

export function getAudioVolume(): number {
  const v = localStorage.getItem(STORAGE_KEY_VOLUME);
  return v ? parseFloat(v) : 0.7;
}

export default function AudioControls() {
  const [muted, setMuted] = useState<boolean>(() => getAudioMuted());
  const [volume, setVolume] = useState<number>(() => getAudioVolume());

  useEffect(() => {
    setAmbientMuted(muted);
    localStorage.setItem(STORAGE_KEY_MUTED, String(muted));
  }, [muted]);

  useEffect(() => {
    setAmbientVolume(volume);
    localStorage.setItem(STORAGE_KEY_VOLUME, String(volume));
  }, [volume]);

  return (
    <div className="audio-controls">
      <button
        onClick={() => setMuted(m => !m)}
        title={muted ? 'Unmute audio' : 'Mute audio'}
        className="audio-mute-btn"
      >
        {muted ? '🔇' : '🔊'}
      </button>
      <input
        type="range"
        min={0}
        max={1}
        step={0.05}
        value={volume}
        onChange={e => setVolume(parseFloat(e.target.value))}
        disabled={muted}
        title="Volume"
        className="audio-volume-slider"
      />
    </div>
  );
}

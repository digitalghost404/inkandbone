// ambient.ts — Ambient audio loop manager with fade in/out

const FADE_STEPS = 20;
const FADE_INTERVAL_MS = 50;
const MAX_VOLUME = 0.6;

interface AmbientTrack {
  audio: HTMLAudioElement;
  tag: string;
}

let currentTrack: AmbientTrack | null = null;
let masterVolume = 1.0;
let muted = false;
let paused = false;

export function setAmbientVolume(volume: number): void {
  masterVolume = Math.max(0, Math.min(1, volume));
  if (currentTrack && !muted && !paused) {
    currentTrack.audio.volume = masterVolume * MAX_VOLUME;
  }
}

export function setAmbientMuted(isMuted: boolean): void {
  muted = isMuted;
  if (currentTrack) {
    if (muted) {
      currentTrack.audio.volume = 0;
    } else if (!paused) {
      currentTrack.audio.volume = masterVolume * MAX_VOLUME;
    }
  }
}

export function pauseAmbient(): void {
  paused = true;
  if (currentTrack) {
    currentTrack.audio.pause();
  }
}

export function resumeAmbient(): void {
  paused = false;
  if (currentTrack && !muted) {
    currentTrack.audio.play().catch(() => {});
  }
}

function fadeOut(audio: HTMLAudioElement): Promise<void> {
  return new Promise(resolve => {
    const startVol = audio.volume;
    const step = startVol / FADE_STEPS;
    let count = 0;
    const interval = setInterval(() => {
      count++;
      audio.volume = Math.max(0, startVol - step * count);
      if (count >= FADE_STEPS) {
        clearInterval(interval);
        audio.pause();
        audio.currentTime = 0;
        resolve();
      }
    }, FADE_INTERVAL_MS);
  });
}

function fadeIn(audio: HTMLAudioElement, targetVolume: number): void {
  audio.volume = 0;
  audio.play().catch(() => {});
  const step = targetVolume / FADE_STEPS;
  let count = 0;
  const interval = setInterval(() => {
    count++;
    audio.volume = Math.min(targetVolume, step * count);
    if (count >= FADE_STEPS) {
      clearInterval(interval);
    }
  }, FADE_INTERVAL_MS);
}

export async function setAmbientTrack(tag: string | null): Promise<void> {
  // If same tag, do nothing
  if (currentTrack && tag === currentTrack.tag) return;

  // Fade out current track
  if (currentTrack) {
    await fadeOut(currentTrack.audio);
    currentTrack = null;
  }

  if (!tag) return;

  // Load new track; respect paused and muted state
  const audio = new Audio(`/api/files/audio/${tag}.mp3`);
  audio.loop = true;
  currentTrack = { audio, tag };

  if (!paused && !muted) {
    fadeIn(audio, masterVolume * MAX_VOLUME);
  }
}

export function stopAmbient(): void {
  if (currentTrack) {
    currentTrack.audio.pause();
    currentTrack.audio.currentTime = 0;
    currentTrack = null;
  }
}

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fetchRuleset, patchCharacter, uploadPortrait, ingestRulebook } from './api';

beforeEach(() => {
  vi.restoreAllMocks();
});

describe('fetchRuleset', () => {
  it('returns ruleset on success', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ id: 1, name: 'dnd5e', schema_json: '{}' }),
    }));
    const rs = await fetchRuleset(1);
    expect(rs.name).toBe('dnd5e');
  });

  it('throws on error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }));
    await expect(fetchRuleset(1)).rejects.toThrow('fetchRuleset failed: 404');
  });
});

describe('patchCharacter', () => {
  it('calls PATCH and returns updated character', async () => {
    const char = { id: 2, name: 'Aria', data_json: '{"hp":10}', campaign_id: 1, portrait_path: '' };
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(char),
    }));
    const result = await patchCharacter(2, { hp: 10 });
    expect(result.id).toBe(2);
    const calls = (fetch as ReturnType<typeof vi.fn>).mock.calls;
    expect(calls[0][0]).toBe('/api/characters/2');
    expect(calls[0][1].method).toBe('PATCH');
  });

  it('throws on error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 400 }));
    await expect(patchCharacter(2, {})).rejects.toThrow('patchCharacter failed: 400');
  });
});

describe('uploadPortrait', () => {
  it('calls POST with FormData and returns updated character', async () => {
    const char = { id: 3, name: 'Bard', data_json: '{}', campaign_id: 1, portrait_path: 'portraits/a.jpg' };
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(char),
    }));
    const file = new File(['data'], 'portrait.jpg', { type: 'image/jpeg' });
    const result = await uploadPortrait(3, file);
    expect(result.portrait_path).toBe('portraits/a.jpg');
  });

  it('throws on error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 415 }));
    const file = new File([''], 'x.bmp');
    await expect(uploadPortrait(3, file)).rejects.toThrow('uploadPortrait failed: 415');
  });
});

describe('ingestRulebook', () => {
  it('posts plain text and returns chunk count', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ chunks_created: 5 }),
    }));
    const result = await ingestRulebook(1, '# Heading\nSome rules.');
    expect(result.chunks_created).toBe(5);
    const calls = (fetch as ReturnType<typeof vi.fn>).mock.calls;
    expect(calls[0][0]).toBe('/api/rulesets/1/rulebook');
    expect(calls[0][1].headers['Content-Type']).toBe('text/plain');
  });

  it('throws on error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 422 }));
    await expect(ingestRulebook(1, 'bad')).rejects.toThrow('ingestRulebook failed: 422');
  });
});

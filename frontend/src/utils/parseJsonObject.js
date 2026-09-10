/**
 * Safely parse a JSON object string.
 * Also recovers common invalid forms like `{Content-Type: application/json}`.
 */
export function parseJsonObject(raw, fallback = {}) {
  if (raw == null || raw === '') return fallback;
  if (typeof raw === 'object' && !Array.isArray(raw)) return raw;

  const text = String(raw).trim();
  if (!text || text === '{}') return fallback;

  try {
    const parsed = JSON.parse(text);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed;
    }
    return fallback;
  } catch {
    if (!text.startsWith('{') || !text.endsWith('}')) {
      return fallback;
    }

    const recovered = {};
    const inner = text.slice(1, -1).trim();
    if (!inner) return fallback;

    for (const part of inner.split(',')) {
      const colon = part.indexOf(':');
      const eq = part.indexOf('=');
      let idx = -1;
      if (colon >= 0 && eq >= 0) idx = Math.min(colon, eq);
      else if (colon >= 0) idx = colon;
      else idx = eq;
      if (idx <= 0) continue;

      const key = part.slice(0, idx).trim().replace(/^['"]|['"]$/g, '');
      const val = part.slice(idx + 1).trim().replace(/^['"]|['"]$/g, '');
      if (key) recovered[key] = val;
    }

    return Object.keys(recovered).length ? recovered : fallback;
  }
}

/** Normalize headers/params fields to valid JSON before persisting. */
export function normalizeJsonObjectField(raw, fallback = '{}') {
  const obj = parseJsonObject(raw, null);
  if (obj == null) return fallback;
  return JSON.stringify(obj);
}

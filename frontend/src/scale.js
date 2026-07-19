// Sequential single-hue (blue) color scale for the familiarity heatmap.
// Familiarity average is a continuous magnitude in [0, 3]; a single hue that
// darkens with magnitude is the correct, colorblind-safe encoding.
//
// Anchors are steps from the validated reference blue ramp (light -> dark).

const STOPS = [
  { t: 0.0, rgb: [205, 226, 251] }, // #cde2fb  (avg 0)
  { t: 0.33, rgb: [134, 182, 239] }, // #86b6ef (avg ~1)
  { t: 0.66, rgb: [57, 135, 229] }, // #3987e5  (avg ~2)
  { t: 1.0, rgb: [24, 79, 149] }, // #184f95    (avg 3)
]

function lerp(a, b, t) {
  return Math.round(a + (b - a) * t)
}

// colorForAverage returns a CSS rgb() string for an average in [0,3], or a
// neutral "no data" token name when avg is null/undefined (n === 0).
export function colorForAverage(avg) {
  if (avg === null || avg === undefined) return 'var(--nodata)'
  const t = Math.max(0, Math.min(1, avg / 3))
  for (let i = 0; i < STOPS.length - 1; i++) {
    const a = STOPS[i]
    const b = STOPS[i + 1]
    if (t >= a.t && t <= b.t) {
      const local = (t - a.t) / (b.t - a.t)
      const r = lerp(a.rgb[0], b.rgb[0], local)
      const g = lerp(a.rgb[1], b.rgb[1], local)
      const bl = lerp(a.rgb[2], b.rgb[2], local)
      return `rgb(${r}, ${g}, ${bl})`
    }
  }
  return `rgb(${STOPS[STOPS.length - 1].rgb.join(', ')})`
}

// textOn returns a readable ink color (dark or light) for text placed on top of
// a heatmap cell of the given average, keeping contrast AA-legible.
export function textOn(avg) {
  if (avg === null || avg === undefined) return 'var(--text-secondary)'
  // Cells get visibly dark above ~avg 1.7; flip to light text there.
  return avg >= 1.7 ? '#ffffff' : '#0b1b33'
}

// The four legend swatches (discrete anchors) for the scale legend.
export const LEGEND_SWATCHES = [0, 1, 2, 3].map((v) => ({
  value: v,
  color: colorForAverage(v),
  text: textOn(v),
}))

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'

const props = defineProps({
  points: { type: Array, required: true }, // [{ x, z }]
  curX:   { type: Number, default: null },
  curZ:   { type: Number, default: null }
})

const canvas = ref(null)
let dirty = false
let rafId = null

function markDirty() { dirty = true }

// Redraw whenever the car position changes (curX/curZ update every packet but
// the parent only calls markDirty() when a new route point is actually pushed,
// so the position dot would otherwise freeze between route-point additions).
watch(() => props.curX, markDirty)

function loop() {
  if (dirty) {
    dirty = false
    draw()
  }
  rafId = requestAnimationFrame(loop)
}

function draw() {
  const c = canvas.value
  if (!c) return
  const ctx = c.getContext('2d')
  const W = c.width
  const H = c.height

  // Background
  ctx.fillStyle = '#060c18'
  ctx.fillRect(0, 0, W, H)

  const pts = props.points
  if (pts.length < 2) {
    ctx.fillStyle = '#2a3a50'
    ctx.font = '12px system-ui'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('No position data yet', W / 2, H / 2)
    return
  }

  // Bounding box
  let minX = pts[0].x, maxX = pts[0].x
  let minZ = pts[0].z, maxZ = pts[0].z
  for (const p of pts) {
    if (p.x < minX) minX = p.x
    if (p.x > maxX) maxX = p.x
    if (p.z < minZ) minZ = p.z
    if (p.z > maxZ) maxZ = p.z
  }
  if (props.curX != null) {
    if (props.curX < minX) minX = props.curX
    if (props.curX > maxX) maxX = props.curX
    if (props.curZ < minZ) minZ = props.curZ
    if (props.curZ > maxZ) maxZ = props.curZ
  }

  const pad = 20
  const rangeX = maxX - minX || 1
  const rangeZ = maxZ - minZ || 1

  // Uniform scale preserving aspect ratio
  const scaleX = (W - pad * 2) / rangeX
  const scaleZ = (H - pad * 2) / rangeZ
  const scale = Math.min(scaleX, scaleZ)

  // Center the drawing within the canvas
  const ox = pad + ((W - pad * 2) - rangeX * scale) / 2
  const oz = pad + ((H - pad * 2) - rangeZ * scale) / 2

  // World → canvas: flip Z so world +Z appears at top of canvas
  function tc(x, z) {
    return [
      ox + (x - minX) * scale,
      H - oz - (z - minZ) * scale
    ]
  }

  // Full route in dim color
  ctx.beginPath()
  ctx.strokeStyle = '#1e4a7a'
  ctx.lineWidth = 1.5
  ctx.lineJoin = 'round'
  ctx.lineCap = 'round'
  let [sx, sy] = tc(pts[0].x, pts[0].z)
  ctx.moveTo(sx, sy)
  for (let i = 1; i < pts.length; i++) {
    const [px, py] = tc(pts[i].x, pts[i].z)
    ctx.lineTo(px, py)
  }
  ctx.stroke()

  // Recent ~200 points in bright color
  const recentN = Math.min(200, pts.length)
  if (recentN > 1) {
    const start = pts.length - recentN
    ctx.beginPath()
    ctx.strokeStyle = '#4a90d9'
    ctx.lineWidth = 2
    const [rx, ry] = tc(pts[start].x, pts[start].z)
    ctx.moveTo(rx, ry)
    for (let i = start + 1; i < pts.length; i++) {
      const [px, py] = tc(pts[i].x, pts[i].z)
      ctx.lineTo(px, py)
    }
    ctx.stroke()
  }

  // Start marker (dim dot)
  const [startX, startY] = tc(pts[0].x, pts[0].z)
  ctx.beginPath()
  ctx.arc(startX, startY, 3, 0, Math.PI * 2)
  ctx.fillStyle = '#445577'
  ctx.fill()

  // Current position with glow
  if (props.curX != null) {
    const [cx, cy] = tc(props.curX, props.curZ)
    const glow = ctx.createRadialGradient(cx, cy, 0, cx, cy, 12)
    glow.addColorStop(0, 'rgba(68, 221, 136, 0.45)')
    glow.addColorStop(1, 'rgba(68, 221, 136, 0)')
    ctx.beginPath()
    ctx.arc(cx, cy, 12, 0, Math.PI * 2)
    ctx.fillStyle = glow
    ctx.fill()
    ctx.beginPath()
    ctx.arc(cx, cy, 4, 0, Math.PI * 2)
    ctx.fillStyle = '#44dd88'
    ctx.fill()
  }
}

onMounted(() => { dirty = true; loop() })
onUnmounted(() => { if (rafId) cancelAnimationFrame(rafId) })

defineExpose({ markDirty })
</script>

<template>
  <div class="route-map">
    <div class="route-map-label">Route</div>
    <canvas ref="canvas" width="320" height="320"></canvas>
  </div>
</template>

<style scoped>
.route-map {
  background: #131922;
  border-radius: 10px;
  padding: 14px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 348px;
}

.route-map-label {
  font-size: 11px;
  color: #6677aa;
  text-transform: uppercase;
  letter-spacing: 0.8px;
}

canvas {
  display: block;
  border-radius: 6px;
  background: #060c18;
  width: 100%;
  height: auto;
}
</style>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { StartRecord, StartReplay, StopSession, ReplayPause, ReplayResume, ReplaySeek, ListSessions, OpenSessionsDir, RenameSession, GetListenAddr, SaveListenAddr, ListTunes } from '../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import SessionCombobox from './components/SessionCombobox.vue'
import RouteMap from './components/RouteMap.vue'
import TuningPanel from './components/TuningPanel.vue'
import TuneSelect from './components/TuneSelect.vue'
import { groups as tuneGroups } from './tuneSchema'

const mode = ref('live') // 'live' | 'recording' | 'replaying'
const filename = ref('session.bin')
const realtimeReplay = ref(true)
const telemetry = ref(null)
const errorMsg = ref('')
const replayFrame = ref(0)
const replayTotal = ref(0)
const replayPaused = ref(false)
const scrubbing = ref(false)
const sessions = ref([])
const renaming = ref(false)
const renameValue = ref('')
const routePoints = ref([])
const routeMap = ref(null)
const MAX_ROUTE_POINTS = 10000

const showSettings = ref(false)
const listenAddr = ref('')
const settingsError = ref('')

const showTuning = ref(false)

// Car tune selector shown in the telemetry view
const tunes = ref([])
const selectedTuneId = ref(0)
const selectedTune = computed(() =>
  tunes.value.find((t) => t.id === selectedTuneId.value) || null
)

async function loadTunes() {
  tunes.value = await ListTunes('')
  if (selectedTuneId.value && !tunes.value.some((t) => t.id === selectedTuneId.value)) {
    selectedTuneId.value = 0
  }
}

// Groups (with at least one set value on the selected tune) for the readout
const selectedTuneGroups = computed(() => {
  const t = selectedTune.value
  if (!t) return []
  return tuneGroups
    .map((g) => ({
      title: g.title,
      fields: g.fields.filter((f) => t[f.key] != null),
    }))
    .filter((g) => g.fields.length > 0)
})

async function openSettings() {
  settingsError.value = ''
  listenAddr.value = await GetListenAddr()
  showSettings.value = true
}

async function saveSettings() {
  const addr = listenAddr.value.trim()
  if (!addr) { settingsError.value = 'Address cannot be empty'; return }
  const err = await SaveListenAddr(addr)
  if (err) { settingsError.value = err; return }
  showSettings.value = false
}

function addRoutePoint(x, z, isRaceOn) {
  if (!isRaceOn) return // game paused — position data is unreliable
  const pts = routePoints.value
  if (pts.length > 0) {
    const last = pts[pts.length - 1]
    const dx = x - last.x
    const dz = z - last.z
    const distSq = dx * dx + dz * dz
    if (distSq < 4) return       // < 2 m — no movement, skip
    if (distSq > 10000) return   // > 100 m jump — teleport/pause artifact, skip
  }
  pts.push({ x, z })
  if (pts.length > MAX_ROUTE_POINTS) pts.splice(0, pts.length - MAX_ROUTE_POINTS)
  routeMap.value?.markDirty()
}

async function refreshSessions() {
  sessions.value = await ListSessions()
  if (sessions.value.length && !sessions.value.includes(filename.value)) {
    filename.value = sessions.value[0]
  }
}

function startRename() {
  renameValue.value = filename.value.replace(/\.bin$/i, '')
  renaming.value = true
}

async function confirmRename() {
  const newName = renameValue.value.trim()
  const currentBase = filename.value.replace(/\.bin$/i, '')
  if (!newName || newName === currentBase) { renaming.value = false; return }
  const finalName = newName.endsWith('.bin') ? newName : newName + '.bin'
  const err = await RenameSession(filename.value, finalName)
  renaming.value = false
  if (err) { errorMsg.value = err; return }
  filename.value = finalName
  await refreshSessions()
}

async function startRecord() {
  routePoints.value = []
  errorMsg.value = ''
  const now = new Date()
  const pad = (n, len = 2) => String(n).padStart(len, '0')
  const stamp = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}` +
                `${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`
  filename.value = `session_${stamp}.bin`
  const err = await StartRecord(filename.value)
  if (err) { errorMsg.value = err; return }
  mode.value = 'recording'
  // refresh list after recording finishes (triggered by session:mode event)
}

async function startReplay() {
  routePoints.value = []
  errorMsg.value = ''
  const err = await StartReplay(filename.value, realtimeReplay.value)
  if (err) { errorMsg.value = err; return }
  mode.value = 'replaying'
}

function stopSession() {
  StopSession()
}

function togglePlayPause() {
  if (replayPaused.value) {
    replayPaused.value = false
    ReplayResume()
  } else {
    replayPaused.value = true
    ReplayPause()
  }
}

function onScrubberPointerDown() {
  scrubbing.value = true
  if (!replayPaused.value) {
    ReplayPause()
    // don't update replayPaused — we'll always resume on release
  }
}

function onScrubberInput(e) {
  replayFrame.value = parseInt(e.target.value, 10)
}

function onScrubberPointerUp(e) {
  const frame = parseInt(e.target.value, 10)
  replayFrame.value = frame
  ReplaySeek(frame)
  replayPaused.value = false
  ReplayResume()
  scrubbing.value = false
}

function onScrubberChange(e) {
  const frame = parseInt(e.target.value, 10)
  replayFrame.value = frame
  ReplaySeek(frame)
}

function fmt1(v) { return v != null ? v.toFixed(1) : '—' }
function fmtInt(v) { return v != null ? Math.round(v) : '—' }
function fmtGear(v) {
  if (v == null) return '—'
  if (v === 0) return 'N'
  if (v === 11) return 'R'
  return String(v)
}

onMounted(async () => {
  await refreshSessions()
  await loadTunes()
  EventsOn('telemetry', (data) => {
    telemetry.value = data
    addRoutePoint(data.posX, data.posZ, data.isRaceOn)
  })
  EventsOn('replay:progress', (p) => {
    if (!scrubbing.value) {
      replayFrame.value = p.frame
      replayTotal.value = p.total
    }
  })
  EventsOn('session:mode', (m) => {
    mode.value = m
    if (m === 'live') {
      telemetry.value = null
      replayFrame.value = 0
      replayTotal.value = 0
      replayPaused.value = false
      refreshSessions()
    }
  })
})

onUnmounted(() => {
  EventsOff('telemetry')
  EventsOff('replay:progress')
  EventsOff('session:mode')
})
</script>

<template>
  <div class="app">
    <div class="app-header">
      <h1>Forza Horizon 6 Telemetry</h1>
      <div class="header-actions">
        <button class="btn-icon tooltip" @click="showTuning = true">
          🔧
          <span class="tooltip-text">Car Tuning</span>
        </button>
        <button class="btn-icon tooltip" @click="openSettings">
          ⚙️
          <span class="tooltip-text">Settings</span>
        </button>
      </div>
    </div>

    <div class="controls">
      <div class="row">
        <label>File</label>
        <SessionCombobox v-model="filename" :sessions="sessions" :disabled="mode !== 'live' || renaming" />
        <button class="btn-icon tooltip" @click="startRename" :disabled="mode !== 'live' || !filename || renaming">
          ✏️
          <span class="tooltip-text">Rename</span>
        </button>
        <!-- Refresh button: only needed when a .bin file is manually dropped into the sessions folder from outside the app
        <button class="btn-icon tooltip" @click="refreshSessions" :disabled="mode !== 'live'">
          ↻
          <span class="tooltip-text">Refresh list</span>
        </button>
        -->
        <button class="btn-icon tooltip" @click="OpenSessionsDir">
          📂
          <span class="tooltip-text">Show in folder</span>
        </button>
      </div>
      <div v-if="renaming" class="row rename-row">
        <label>→</label>
        <input v-model="renameValue" class="rename-input" @keyup.enter="confirmRename" @keyup.escape="renaming = false" autofocus />
        <button class="btn btn-save" @click="confirmRename">✓ Save</button>
        <button class="btn btn-cancel" @click="renaming = false">✗ Cancel</button>
      </div>
      <div class="row">
        <button @click="startRecord" :disabled="mode !== 'live'" class="btn record">
          ● Record
        </button>
        <button @click="startReplay" :disabled="mode !== 'live'" class="btn replay">
          ▶ Replay
        </button>
        <label class="check tooltip">
          <input type="checkbox" v-model="realtimeReplay" :disabled="mode !== 'live'" />
          Realtime
          <span class="tooltip-text">Without Realtime, replay runs at full speed and may spike CPU usage.</span>
        </label>
        <button @click="stopSession" :disabled="mode === 'live'" class="btn stop">
          ■ Stop
        </button>
      </div>
      <div class="status" :class="mode">{{ mode.toUpperCase() }}</div>
      <div v-if="errorMsg" class="error">{{ errorMsg }}</div>

      <div class="row tune-select-row">
        <label>Tune</label>
        <TuneSelect v-model="selectedTuneId" :tunes="tunes" />
        <button class="btn-icon tooltip" @click="showTuning = true">
          🔧
          <span class="tooltip-text">Manage tunes</span>
        </button>
      </div>

      <!-- Playback bar — visible only during replay -->
      <div v-if="mode === 'replaying'" class="playback-bar">
        <button class="btn-icon" @click="togglePlayPause" :title="replayPaused ? 'Resume' : 'Pause'">
          {{ replayPaused ? '▶' : '⏸' }}
        </button>
        <input
          class="scrubber"
          type="range"
          min="0"
          :max="replayTotal > 0 ? replayTotal - 1 : 0"
          :value="replayFrame"
          @pointerdown="onScrubberPointerDown"
          @input="onScrubberInput"
          @pointerup="onScrubberPointerUp"
        />
        <span class="playback-counter">{{ replayFrame + 1 }} / {{ replayTotal }}</span>
      </div>
    </div>

    <div v-if="telemetry" class="telemetry-layout">
      <RouteMap ref="routeMap" :points="routePoints" :cur-x="telemetry.posX" :cur-z="telemetry.posZ" />
      <div class="telemetry">
      <div class="race-badge" :class="telemetry.isRaceOn ? 'on' : 'off'">
        {{ telemetry.isRaceOn ? '🟢 Race ON' : '🔴 Race OFF' }}
      </div>

      <div v-if="selectedTune" class="section tune-readout">
        <div class="tune-readout-head">
          <span class="tune-readout-title">🔧 {{ selectedTune.name || '(unnamed)' }}</span>
          <span v-if="selectedTune.notes" class="tune-readout-notes">{{ selectedTune.notes }}</span>
        </div>
        <div class="tune-readout-groups">
          <div v-for="g in selectedTuneGroups" :key="g.title" class="tune-readout-group">
            <h4>{{ g.title }}</h4>
            <div class="tune-readout-fields">
              <div v-for="f in g.fields" :key="f.key" class="tune-readout-field">
                <span>{{ f.label }}</span>
                <strong>{{ selectedTune[f.key] }}</strong>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="section">
        <div class="stat big">
          <label>Speed</label>
          <span>{{ fmt1(telemetry.speed) }} <em>km/h</em></span>
        </div>
        <div class="stat big">
          <label>RPM</label>
          <span>{{ fmtInt(telemetry.rpm) }} <em>/ {{ fmtInt(telemetry.maxRpm) }}</em></span>
        </div>
        <div class="stat big">
          <label>Gear</label>
          <span>{{ fmtGear(telemetry.gear) }}</span>
        </div>
      </div>

      <div class="section">
        <div class="stat">
          <label>Throttle</label>
          <div class="bar-wrap">
            <div class="bar throttle" :style="{ width: telemetry.throttle + '%' }"></div>
          </div>
          <span>{{ fmt1(telemetry.throttle) }}%</span>
        </div>
        <div class="stat">
          <label>Brake</label>
          <div class="bar-wrap">
            <div class="bar brake" :style="{ width: telemetry.brake + '%' }"></div>
          </div>
          <span>{{ fmt1(telemetry.brake) }}%</span>
        </div>
        <div class="stat">
          <label>Boost</label>
          <span>{{ fmt1(telemetry.boost) }}</span>
        </div>
        <div class="stat">
          <label>Fuel</label>
          <span>{{ fmt1(telemetry.fuel * 100) }}%</span>
        </div>
      </div>

      <div class="section">
        <div class="stat">
          <label>Tire °C</label>
          <span class="tires">
            <span>FL {{ fmt1(telemetry.tireTempFL) }}</span>
            <span>FR {{ fmt1(telemetry.tireTempFR) }}</span>
            <span>RL {{ fmt1(telemetry.tireTempRL) }}</span>
            <span>RR {{ fmt1(telemetry.tireTempRR) }}</span>
          </span>
        </div>
      </div>

      <div class="section">
        <div class="stat">
          <label>Lap</label>
          <span>{{ telemetry.lapNumber }}</span>
        </div>
        <div class="stat">
          <label>Position</label>
          <span>{{ telemetry.racePosition }}</span>
        </div>
        <div class="stat">
          <label>Current Lap</label>
          <span>{{ fmt1(telemetry.currentLapTime) }}s</span>
        </div>
        <div class="stat">
          <label>Best Lap</label>
          <span>{{ fmt1(telemetry.bestLap) }}s</span>
        </div>
        <div class="stat">
          <label>Last Lap</label>
          <span>{{ fmt1(telemetry.lastLap) }}s</span>
        </div>
      </div>
      </div>
    </div>

    <div v-else class="placeholder">
      Waiting for telemetry from Forza...
    </div>

    <div v-if="showSettings" class="modal-overlay" @click.self="showSettings = false">
      <div class="modal">
        <h2>Settings</h2>
        <label class="modal-label">UDP Listen Address</label>
        <input
          v-model="listenAddr"
          class="modal-input"
          placeholder="0.0.0.0:8000"
          @keyup.enter="saveSettings"
          @keyup.escape="showSettings = false"
        />
        <p class="modal-hint">Host:port the app listens on for Forza data. Default is 0.0.0.0:8000.</p>
        <div v-if="settingsError" class="error">{{ settingsError }}</div>
        <div class="modal-actions">
          <button class="btn btn-save" @click="saveSettings">✓ Save</button>
          <button class="btn btn-cancel" @click="showSettings = false">✗ Cancel</button>
        </div>
      </div>
    </div>

    <TuningPanel :show="showTuning" @close="showTuning = false" @changed="loadTunes" />
  </div>
</template>


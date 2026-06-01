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
  <div class="max-w-[1400px] mx-auto px-8 py-7">

    <!-- Header -->
    <div class="flex items-center justify-between gap-3 mb-5">
      <h1 class="text-xl font-bold text-white tracking-wide">Forza Horizon 6 Telemetry</h1>
      <div class="flex items-center gap-2">
        <button class="relative group bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center flex-shrink-0 transition-colors hover:bg-[#2a3a50]" @click="showTuning = true">
          🔧
          <span class="hidden group-hover:block absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 bg-[#1a2030] text-[#c0cce0] border border-[#2a3a50] rounded-md px-2.5 py-1.5 text-xs whitespace-nowrap pointer-events-none z-10">Car Tuning</span>
        </button>
        <button class="relative group bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center flex-shrink-0 transition-colors hover:bg-[#2a3a50]" @click="openSettings">
          ⚙️
          <span class="hidden group-hover:block absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 bg-[#1a2030] text-[#c0cce0] border border-[#2a3a50] rounded-md px-2.5 py-1.5 text-xs whitespace-nowrap pointer-events-none z-10">Settings</span>
        </button>
      </div>
    </div>

    <!-- Controls -->
    <div class="mb-6">
      <!-- File row -->
      <div class="flex items-center gap-2.5 mb-2.5">
        <label class="min-w-[34px] text-[#aabbcc]">File</label>
        <SessionCombobox v-model="filename" :sessions="sessions" :disabled="mode !== 'live' || renaming" />
        <button class="relative group bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center flex-shrink-0 transition-colors hover:bg-[#2a3a50] disabled:opacity-50 disabled:cursor-not-allowed" @click="startRename" :disabled="mode !== 'live' || !filename || renaming">
          ✏️
          <span class="hidden group-hover:block absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 bg-[#1a2030] text-[#c0cce0] border border-[#2a3a50] rounded-md px-2.5 py-1.5 text-xs whitespace-nowrap pointer-events-none z-10">Rename</span>
        </button>
        <button class="relative group bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center flex-shrink-0 transition-colors hover:bg-[#2a3a50]" @click="OpenSessionsDir">
          📂
          <span class="hidden group-hover:block absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 bg-[#1a2030] text-[#c0cce0] border border-[#2a3a50] rounded-md px-2.5 py-1.5 text-xs whitespace-nowrap pointer-events-none z-10">Show in folder</span>
        </button>
      </div>

      <!-- Rename row -->
      <div v-if="renaming" class="flex items-center gap-2.5 mb-2.5 -mt-1">
        <label class="min-w-[34px] text-[#aabbcc]">→</label>
        <input v-model="renameValue" class="bg-[#1a2030] border border-[#4a90d9] text-[#e0e6f0] px-2.5 py-1.5 rounded-md flex-1 outline-none text-[13px] focus:border-[#60aaff]" @keyup.enter="confirmRename" @keyup.escape="renaming = false" autofocus />
        <button class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm bg-[#2a7ae2] text-white hover:bg-[#3a8af2] transition-colors" @click="confirmRename">✓ Save</button>
        <button class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm bg-[#333d4d] text-[#aabbcc] hover:bg-[#3d4d60] transition-colors" @click="renaming = false">✗ Cancel</button>
      </div>

      <!-- Record / Replay / Stop row -->
      <div class="flex items-center gap-2.5 mb-2.5">
        <button @click="startRecord" :disabled="mode !== 'live'" class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm transition-opacity disabled:opacity-[0.35] disabled:cursor-not-allowed bg-[#d94040] text-white">● Record</button>
        <button @click="startReplay" :disabled="mode !== 'live'" class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm transition-opacity disabled:opacity-[0.35] disabled:cursor-not-allowed bg-[#2a7ae2] text-white">▶ Replay</button>
        <label class="relative group flex items-center gap-1.5 text-[#aabbcc] cursor-pointer">
          <input type="checkbox" v-model="realtimeReplay" :disabled="mode !== 'live'" />
          Realtime
          <span class="hidden group-hover:block absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 bg-[#1a2030] text-[#c0cce0] border border-[#2a3a50] rounded-md px-2.5 py-1.5 text-xs pointer-events-none z-10 w-[260px] text-center leading-snug">Without Realtime, replay runs at full speed and may spike CPU usage.</span>
        </label>
        <button @click="stopSession" :disabled="mode === 'live'" class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm transition-opacity disabled:opacity-[0.35] disabled:cursor-not-allowed bg-[#555] text-white">■ Stop</button>
      </div>

      <!-- Mode badge -->
      <div class="inline-block px-3 py-0.5 rounded-full text-[11px] font-bold tracking-widest mt-1"
           :class="{
             'bg-[#0d2e1a] text-[#44dd88]': mode === 'live',
             'bg-[#3d1515] text-[#ff6060]': mode === 'recording',
             'bg-[#0d2040] text-[#60aaff]': mode === 'replaying',
           }">{{ mode.toUpperCase() }}</div>
      <div v-if="errorMsg" class="text-[#ff6060] text-[13px] mt-2">{{ errorMsg }}</div>

      <!-- Tune selector row -->
      <div class="flex items-center gap-2.5 mt-2.5 mb-2.5">
        <label class="min-w-[34px] text-[#aabbcc]">Tune</label>
        <TuneSelect v-model="selectedTuneId" :tunes="tunes" />
        <button class="relative group bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center flex-shrink-0 transition-colors hover:bg-[#2a3a50]" @click="showTuning = true">
          🔧
          <span class="hidden group-hover:block absolute bottom-[calc(100%+6px)] left-1/2 -translate-x-1/2 bg-[#1a2030] text-[#c0cce0] border border-[#2a3a50] rounded-md px-2.5 py-1.5 text-xs whitespace-nowrap pointer-events-none z-10">Manage tunes</span>
        </button>
      </div>

      <!-- Playback bar -->
      <div v-if="mode === 'replaying'" class="flex items-center gap-2.5 mt-2.5 bg-[#131922] rounded-lg px-3 py-2">
        <button class="bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center flex-shrink-0 transition-colors hover:bg-[#2a3a50]" @click="togglePlayPause" :title="replayPaused ? 'Resume' : 'Pause'">
          {{ replayPaused ? '▶' : '⏸' }}
        </button>
        <input
          class="scrubber flex-1"
          type="range"
          min="0"
          :max="replayTotal > 0 ? replayTotal - 1 : 0"
          :value="replayFrame"
          @pointerdown="onScrubberPointerDown"
          @input="onScrubberInput"
          @pointerup="onScrubberPointerUp"
        />
        <span class="text-[11px] text-[#6677aa] whitespace-nowrap min-w-[72px] text-right">{{ replayFrame + 1 }} / {{ replayTotal }}</span>
      </div>
    </div>

    <!-- Telemetry layout -->
    <div v-if="telemetry" class="flex gap-5 items-start">
      <RouteMap ref="routeMap" :points="routePoints" :cur-x="telemetry.posX" :cur-z="telemetry.posZ" />

      <div class="flex flex-col gap-4 flex-1 min-w-0">

        <!-- Race badge -->
        <div class="self-start px-3.5 py-1 rounded-full text-xs font-bold tracking-[0.5px]"
             :class="telemetry.isRaceOn ? 'bg-[#0d2e1a] text-[#44dd77]' : 'bg-[#2e1010] text-[#dd4444]'">
          {{ telemetry.isRaceOn ? '🟢 Race ON' : '🔴 Race OFF' }}
        </div>

        <!-- Tune readout -->
        <div v-if="selectedTune" class="flex flex-col gap-3.5 bg-[#131922] rounded-xl px-5 py-[18px]">
          <div class="flex items-baseline gap-3 flex-wrap">
            <span class="text-[15px] font-bold text-[#8fb3e0]">🔧 {{ selectedTune.name || '(unnamed)' }}</span>
            <span v-if="selectedTune.notes" class="text-xs text-[#6677aa]">{{ selectedTune.notes }}</span>
          </div>
          <div class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3.5 w-full">
            <div v-for="g in selectedTuneGroups" :key="g.title">
              <h4 class="text-[11px] font-bold text-[#6677aa] uppercase tracking-[0.6px] mb-1.5">{{ g.title }}</h4>
              <div class="flex flex-col gap-[3px]">
                <div v-for="f in g.fields" :key="f.key" class="flex items-baseline justify-between gap-2.5 text-[13px] py-0.5 border-b border-[#1c2533]">
                  <span class="text-[#8899aa]">{{ f.label }}</span>
                  <strong class="text-[#e0e6f0] tabular-nums">{{ selectedTune[f.key] }}</strong>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Speed / RPM / Gear -->
        <div class="flex flex-wrap gap-4 bg-[#131922] rounded-xl px-5 py-[18px]">
          <div class="flex flex-col gap-1 min-w-[150px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Speed</label>
            <span class="text-[38px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.speed) }} <em class="text-sm font-normal text-[#8caac8] not-italic">km/h</em></span>
          </div>
          <div class="flex flex-col gap-1 min-w-[150px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">RPM</label>
            <span class="text-[38px] font-bold text-[#f0f6ff]">{{ fmtInt(telemetry.rpm) }} <em class="text-sm font-normal text-[#8caac8] not-italic">/ {{ fmtInt(telemetry.maxRpm) }}</em></span>
          </div>
          <div class="flex flex-col gap-1 min-w-[150px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Gear</label>
            <span class="text-[38px] font-bold text-[#f0f6ff]">{{ fmtGear(telemetry.gear) }}</span>
          </div>
        </div>

        <!-- Throttle / Brake / Boost / Fuel -->
        <div class="flex flex-wrap gap-4 bg-[#131922] rounded-xl px-5 py-[18px]">
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Throttle</label>
            <div class="h-1.5 bg-[#1e2a3a] rounded overflow-hidden w-full">
              <div class="h-full rounded bg-[#44cc66] transition-[width] duration-100" :style="{ width: telemetry.throttle + '%' }"></div>
            </div>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.throttle) }}%</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Brake</label>
            <div class="h-1.5 bg-[#1e2a3a] rounded overflow-hidden w-full">
              <div class="h-full rounded bg-[#dd4444] transition-[width] duration-100" :style="{ width: telemetry.brake + '%' }"></div>
            </div>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.brake) }}%</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Boost</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.boost) }}</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Fuel</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.fuel * 100) }}%</span>
          </div>
        </div>

        <!-- Tire temps -->
        <div class="flex flex-wrap gap-4 bg-[#131922] rounded-xl px-5 py-[18px]">
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Tire °C</label>
            <div class="grid grid-cols-2 gap-x-4 gap-y-1 text-base font-semibold text-[#f0f6ff]">
              <span>FL {{ fmt1(telemetry.tireTempFL) }}</span>
              <span>FR {{ fmt1(telemetry.tireTempFR) }}</span>
              <span>RL {{ fmt1(telemetry.tireTempRL) }}</span>
              <span>RR {{ fmt1(telemetry.tireTempRR) }}</span>
            </div>
          </div>
        </div>

        <!-- Lap times -->
        <div class="flex flex-wrap gap-4 bg-[#131922] rounded-xl px-5 py-[18px]">
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Lap</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ telemetry.lapNumber }}</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Position</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ telemetry.racePosition }}</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Current Lap</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.currentLapTime) }}s</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Best Lap</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.bestLap) }}s</span>
          </div>
          <div class="flex flex-col gap-1 min-w-[110px] flex-1">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.8px]">Last Lap</label>
            <span class="text-[28px] font-bold text-[#f0f6ff]">{{ fmt1(telemetry.lastLap) }}s</span>
          </div>
        </div>

      </div>
    </div>

    <div v-else class="text-center text-[#445566] mt-16 text-[15px]">
      Waiting for telemetry from Forza...
    </div>

    <!-- Settings modal -->
    <div v-if="showSettings" class="fixed inset-0 bg-[rgba(4,8,16,0.7)] flex items-center justify-center z-[100]" @click.self="showSettings = false">
      <div class="bg-[#131922] border border-[#2a3a50] rounded-xl p-6 w-[420px] max-w-[calc(100vw-48px)] flex flex-col gap-2.5">
        <h2 class="text-[17px] font-bold text-white mb-1">Settings</h2>
        <label class="text-[11px] text-[#6677aa] uppercase tracking-[0.8px]">UDP Listen Address</label>
        <input
          v-model="listenAddr"
          class="bg-[#1a2030] border border-[#2a3a50] text-[#e0e6f0] px-3 py-2 rounded-md outline-none text-sm focus:border-[#4a90d9]"
          placeholder="0.0.0.0:8000"
          @keyup.enter="saveSettings"
          @keyup.escape="showSettings = false"
        />
        <p class="text-xs text-[#6677aa] leading-snug">Host:port the app listens on for Forza data. Default is 0.0.0.0:8000.</p>
        <div v-if="settingsError" class="text-[#ff6060] text-[13px] mt-2">{{ settingsError }}</div>
        <div class="flex gap-2.5 mt-2">
          <button class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm bg-[#2a7ae2] text-white hover:bg-[#3a8af2] transition-colors" @click="saveSettings">✓ Save</button>
          <button class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm bg-[#333d4d] text-[#aabbcc] hover:bg-[#3d4d60] transition-colors" @click="showSettings = false">✗ Cancel</button>
        </div>
      </div>
    </div>

    <TuningPanel :show="showTuning" @close="showTuning = false" @changed="loadTunes" />
  </div>
</template>


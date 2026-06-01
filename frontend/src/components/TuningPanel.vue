<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ListTunes, SaveTune, DeleteTune, ImportTunes } from '../../wailsjs/go/main/App'
import { groups, allFields } from '../tuneSchema'

const props = defineProps({
  show: { type: Boolean, default: false },
})
const emit = defineEmits(['close', 'changed'])

function emptyForm() {
  const f = { id: 0, name: '', notes: '' }
  for (const field of allFields) f[field.key] = ''
  return f
}

const form = reactive(emptyForm())
const tunes = ref([])
const search = ref('')
const error = ref('')
const fieldErrors = reactive({})

async function refresh() {
  tunes.value = await ListTunes(search.value.trim())
}

function loadTune(t) {
  form.id = t.id
  form.name = t.name
  form.notes = t.notes || ''
  for (const field of allFields) {
    const v = t[field.key]
    form[field.key] = v == null ? '' : v
  }
  clearFieldErrors()
  error.value = ''
}

function newTune() {
  Object.assign(form, emptyForm())
  clearFieldErrors()
  error.value = ''
}

function clearFieldErrors() {
  for (const k of Object.keys(fieldErrors)) delete fieldErrors[k]
}

function validate() {
  clearFieldErrors()
  let ok = true
  if (!form.name.trim()) { error.value = 'Name is required'; ok = false }
  for (const field of allFields) {
    const raw = form[field.key]
    if (raw === '' || raw == null) continue // optional
    const n = Number(raw)
    if (Number.isNaN(n)) { fieldErrors[field.key] = 'NaN'; ok = false; continue }
    if (n < field.min || n > field.max) {
      fieldErrors[field.key] = `${field.min}–${field.max}`
      ok = false
    }
  }
  return ok
}

function buildPayload() {
  const payload = { id: form.id, name: form.name.trim(), notes: form.notes.trim() }
  for (const field of allFields) {
    const raw = form[field.key]
    payload[field.key] = raw === '' || raw == null ? null : Number(raw)
  }
  return payload
}

async function save() {
  error.value = ''
  if (!validate()) {
    if (!error.value) error.value = 'Some values are out of range'
    return
  }
  const res = await SaveTune(buildPayload())
  if (typeof res === 'string') { error.value = res; return }
  const wasUpdate = form.id !== 0
  form.id = res.id
  await refresh()
  emit('changed')
  showToast(wasUpdate ? 'Update successful' : 'Saved successfully')
}

async function remove(t) {
  const err = await DeleteTune(t.id)
  if (err) { error.value = err; return }
  if (form.id === t.id) newTune()
  await refresh()
  emit('changed')
}

const fileInput = ref(null)

function triggerImport() {
  error.value = ''
  fileInput.value?.click()
}

async function onImportFile(e) {
  const file = e.target.files?.[0]
  e.target.value = '' // allow re-importing the same file
  if (!file) return
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    const arr = Array.isArray(data) ? data : [data]
    const normalized = arr.map((t) => {
      const out = { id: 0, name: String(t.name ?? ''), notes: String(t.notes ?? '') }
      for (const field of allFields) {
        const v = t[field.key]
        out[field.key] = v == null || v === '' ? null : Number(v)
      }
      return out
    })
    const res = await ImportTunes(normalized)
    if (typeof res === 'string') { error.value = res; return }
    await refresh()
    emit('changed')
  } catch (err) {
    error.value = 'Import failed: ' + (err?.message || String(err))
  }
}

function close() {
  emit('close')
}

const filteredCount = computed(() => tunes.value.length)

const activeTab = ref(0)

const toast = ref('')
let toastTimer = null
function showToast(msg) {
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = '' }, 2500)
}

onMounted(refresh)
</script>

<template>
  <div v-if="show" class="fixed inset-0 bg-[rgba(4,8,16,0.7)] flex items-center justify-center z-[100]" @click.self="close">
    <div class="bg-[#131922] border border-[#2a3a50] rounded-xl overflow-hidden w-[1000px] max-w-[calc(100vw-48px)] max-h-[calc(100vh-48px)] flex flex-col">

      <!-- Header -->
      <div class="flex items-center justify-between px-[22px] py-[18px] border-b border-[#2a3a50] flex-shrink-0">
        <h2 class="text-[17px] font-bold text-white m-0">Car Tuning</h2>
        <button class="bg-[#1e2a3a] border-0 text-[#e0e6f0] text-base w-8 h-8 rounded-md cursor-pointer flex items-center justify-center hover:bg-[#2a3a50] transition-colors" @click="close" title="Close">✕</button>
      </div>

      <div class="flex min-h-0 flex-1">
        <!-- Sidebar -->
        <aside class="w-[260px] flex-shrink-0 border-r border-[#2a3a50] p-4 flex flex-col gap-2.5 overflow-y-auto">
          <input
            v-model="search"
            class="appearance-none block w-full h-[34px] px-2.5 bg-[#1a2030] border border-[#2a3a50] rounded-md text-[#e0e6f0] text-sm outline-none placeholder:text-[#445566] focus:border-[#4a90d9]"
            placeholder="Search cars…"
            @input="refresh"
          />
          <div class="flex items-center justify-between text-xs text-[#8caac8] uppercase tracking-[0.6px]">
            <span>{{ filteredCount }} saved</span>
            <div class="flex gap-1.5">
              <button class="px-2 py-1 border-0 rounded-md font-semibold cursor-pointer text-xs bg-[#333d4d] text-[#aabbcc] hover:bg-[#3d4d60] transition-colors" @click="triggerImport">⬆ Import</button>
              <button class="px-2 py-1 border-0 rounded-md font-semibold cursor-pointer text-xs bg-[#2a7ae2] text-white hover:bg-[#3a8af2] transition-colors" @click="newTune">＋ New</button>
            </div>
          </div>
          <input ref="fileInput" type="file" accept="application/json,.json" hidden @change="onImportFile" />
          <ul class="list-none flex flex-col gap-1">
            <li
              v-for="t in tunes"
              :key="t.id"
              class="flex items-center justify-between gap-1.5 px-2.5 py-2 rounded-md cursor-pointer bg-[#1a2030] transition-colors hover:bg-[#20283a]"
              :class="{ 'bg-[#1d3050] outline outline-1 outline-[#4a90d9]': t.id === form.id }"
              @click="loadTune(t)"
            >
              <span class="overflow-hidden text-ellipsis whitespace-nowrap">{{ t.name || '(unnamed)' }}</span>
              <button class="bg-transparent border-0 text-[#e0e6f0] text-[13px] w-[26px] h-[26px] rounded-md cursor-pointer flex items-center justify-center hover:bg-[#3d1515] transition-colors" @click.stop="remove(t)" title="Delete">🗑</button>
            </li>
            <li v-if="!tunes.length" class="flex items-center justify-center px-2.5 py-2 text-[#8caac8] text-sm">No tunes yet</li>
          </ul>
        </aside>

        <!-- Editor -->
        <section class="flex-1 min-w-0 px-[22px] py-[18px] overflow-y-auto flex flex-col gap-4">
          <div class="grid grid-cols-[auto_1fr] items-center gap-x-3 gap-y-2">
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.6px] whitespace-nowrap">Name *</label>
            <input v-model="form.name" class="bg-[#1a2030] border border-[#2a3a50] text-[#e0e6f0] px-3 py-2 rounded-md outline-none text-sm focus:border-[#4a90d9]" placeholder="e.g. 2019 Porsche 911 GT3 RS" />
            <label class="text-xs text-[#8caac8] uppercase tracking-[0.6px] whitespace-nowrap">Notes</label>
            <input v-model="form.notes" class="bg-[#1a2030] border border-[#2a3a50] text-[#e0e6f0] px-3 py-2 rounded-md outline-none text-sm focus:border-[#4a90d9]" placeholder="Optional notes" />
          </div>

          <div class="flex flex-wrap gap-1 pb-3 border-b border-[#2a3a50]">
            <button
              v-for="(g, i) in groups"
              :key="g.title"
              class="px-3 py-1.5 text-xs font-semibold bg-[#1a2030] border border-[#2a3a50] rounded-md cursor-pointer uppercase tracking-[0.5px] whitespace-nowrap transition-all hover:bg-[#20283a]"
              :class="activeTab === i ? 'bg-[#1d3050] !border-[#4a90d9] text-[#e0e6f0]' : 'text-[#8caac8] hover:text-[#aabbcc]'"
              @click="activeTab = i"
            >{{ g.title }}</button>
          </div>

          <div class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-x-6 gap-y-[18px] pr-2">
              <label v-for="field in groups[activeTab].fields" :key="field.key" class="flex flex-col gap-1.5 relative">
                <div class="flex items-center justify-between gap-2">
                  <span class="text-sm text-[#aabbcc]">{{ field.label }}</span>
                  <span class="text-base font-bold min-w-[48px] text-right whitespace-nowrap"
                        :class="form[field.key] === '' || form[field.key] == null ? 'text-[#445566]' : 'text-[#f0f6ff]'">
                    {{ form[field.key] === '' || form[field.key] == null ? '—' : form[field.key] }}
                  </span>
                </div>
                <input
                  type="range"
                  class="tune-range"
                  :class="{ invalid: fieldErrors[field.key] }"
                  :min="field.min"
                  :max="field.max"
                  :step="field.step"
                  :value="form[field.key] === '' || form[field.key] == null ? field.min : form[field.key]"
                  @input="form[field.key] = $event.target.value"
                />
                <div class="flex justify-between text-[13px] text-[#8caac8] mt-1.5">
                  <span>{{ field.min }}</span>
                  <span>{{ field.max }}</span>
                </div>
                <span v-if="fieldErrors[field.key]" class="text-[10px] text-[#ff6060]">{{ fieldErrors[field.key] }}</span>
              </label>
            </div>

          <div v-if="error" class="text-[#ff6060] text-[13px] mt-2">{{ error }}</div>

          <div class="sticky bottom-0 bg-[#131922] pt-2 flex gap-2.5">
            <button class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm bg-[#2a7ae2] text-white hover:bg-[#3a8af2] transition-colors" @click="save">✓ {{ form.id ? 'Update' : 'Save' }}</button>
            <button class="px-4 py-1.5 border-0 rounded-md font-semibold cursor-pointer text-sm bg-[#333d4d] text-[#aabbcc] hover:bg-[#3d4d60] transition-colors" @click="newTune">Clear</button>
          </div>
        </section>
      </div>
    </div>

    <Transition name="tune-toast">
      <div v-if="toast" class="fixed top-6 right-7 bg-[#1b4d2e] border border-[#2e7d4f] text-[#5dd98a] text-sm font-semibold px-[22px] py-2.5 rounded-lg pointer-events-none z-[9999] whitespace-nowrap shadow-[0_4px_18px_rgba(0,0,0,0.5)]">{{ toast }}</div>
    </Transition>
  </div>
</template>



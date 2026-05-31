<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ListTunes, SaveTune, DeleteTune, ImportTunes } from '../../wailsjs/go/main/App'
import { groups, allFields } from '../tuneSchema'
import './TuningPanel.css'

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
  form.id = res.id
  await refresh()
  emit('changed')
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

onMounted(refresh)
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="close">
    <div class="modal tune-modal">
      <div class="tune-header">
        <h2>Car Tuning</h2>
        <button class="btn-icon" @click="close" title="Close">✕</button>
      </div>

      <div class="tune-body">
        <!-- Sidebar: saved tunes -->
        <aside class="tune-list">
          <input
            v-model="search"
            class="tune-search"
            placeholder="Search cars…"
            @input="refresh"
          />
          <div class="tune-list-head">
            <span>{{ filteredCount }} saved</span>
            <div class="tune-list-actions">
              <button class="btn btn-cancel sm" @click="triggerImport">⬆ Import</button>
              <button class="btn btn-save sm" @click="newTune">＋ New</button>
            </div>
          </div>
          <input ref="fileInput" type="file" accept="application/json,.json" hidden @change="onImportFile" />
          <ul>
            <li
              v-for="t in tunes"
              :key="t.id"
              :class="{ active: t.id === form.id }"
              @click="loadTune(t)"
            >
              <span class="tune-name">{{ t.name || '(unnamed)' }}</span>
              <button class="btn-icon del" @click.stop="remove(t)" title="Delete">🗑</button>
            </li>
            <li v-if="!tunes.length" class="empty">No tunes yet</li>
          </ul>
        </aside>

        <!-- Editor -->
        <section class="tune-editor">
          <div class="tune-meta">
            <label>Name *</label>
            <input v-model="form.name" class="modal-input" placeholder="e.g. 2019 Porsche 911 GT3 RS" />
            <label>Notes</label>
            <input v-model="form.notes" class="modal-input" placeholder="Optional notes" />
          </div>

          <div class="tune-groups">
            <div v-for="g in groups" :key="g.title" class="tune-group">
              <h3>{{ g.title }}</h3>
              <div class="tune-fields">
                <label v-for="field in g.fields" :key="field.key" class="tune-field">
                  <span class="tune-field-label">{{ field.label }}</span>
                  <input
                    v-model="form[field.key]"
                    type="number"
                    :step="field.step"
                    :min="field.min"
                    :max="field.max"
                    :placeholder="`${field.min}–${field.max}`"
                    :class="{ invalid: fieldErrors[field.key] }"
                  />
                  <span v-if="fieldErrors[field.key]" class="tune-field-err">{{ fieldErrors[field.key] }}</span>
                </label>
              </div>
            </div>
          </div>

          <div v-if="error" class="error">{{ error }}</div>

          <div class="modal-actions tune-actions">
            <button class="btn btn-save" @click="save">✓ {{ form.id ? 'Update' : 'Save' }}</button>
            <button class="btn btn-cancel" @click="newTune">Clear</button>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>



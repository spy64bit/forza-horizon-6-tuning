<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  sessions:   { type: Array,  default: () => [] },
  disabled:   { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const strip = (s) => s.replace(/\.bin$/i, '')

const search    = ref(strip(props.modelValue))
const open      = ref(false)
const highlight = ref(-1)

const filtered = computed(() => {
  if (!search.value) return props.sessions
  const f = search.value.toLowerCase()
  return props.sessions.filter(s => strip(s).toLowerCase().includes(f))
})

watch(() => props.modelValue, (v) => {
  if (!open.value) search.value = strip(v)
})

function onFocus(e) {
  search.value = strip(props.modelValue)
  open.value = true
  highlight.value = -1
  e.target.select()
}
function onInput(e) {
  search.value = e.target.value
  highlight.value = -1
}
function onBlur() {
  open.value = false
  search.value = strip(props.modelValue)
}
function select(s) {
  emit('update:modelValue', s)
  search.value = strip(s)
  open.value = false
}
function onDown() {
  if (!open.value) { open.value = true; return }
  highlight.value = Math.min(highlight.value + 1, filtered.value.length - 1)
}
function onUp() {
  highlight.value = Math.max(highlight.value - 1, 0)
}
function onEnter() {
  if (highlight.value >= 0 && filtered.value[highlight.value]) {
    select(filtered.value[highlight.value])
  } else if (filtered.value.length === 1) {
    select(filtered.value[0])
  }
}
</script>

<template>
  <div class="combobox">
    <input
      class="combo-input"
      :value="search"
      :disabled="disabled"
      placeholder="No sessions yet"
      autocomplete="off"
      @focus="onFocus"
      @input="onInput"
      @blur="onBlur"
      @keydown.escape.prevent="onBlur"
      @keydown.enter.prevent="onEnter"
      @keydown.down.prevent="onDown"
      @keydown.up.prevent="onUp"
    />
    <ul v-if="open && filtered.length > 0" class="combo-dropdown">
      <li
        v-for="(s, i) in filtered"
        :key="s"
        :class="{ active: highlight === i }"
        @mousedown.prevent="select(s)"
      >{{ strip(s) }}</li>
    </ul>
    <div v-if="open && filtered.length === 0" class="combo-empty">No matches</div>
  </div>
</template>

<style scoped>
.combobox { position: relative; flex: 1; }

.combo-input {
  background: #1a2030;
  border: 1px solid #2a3a50;
  color: #e0e6f0;
  padding: 6px 10px;
  border-radius: 6px;
  outline: none;
  width: 100%;
  font-size: 13px;
  cursor: text;
  text-align: left;
}
.combo-input::placeholder { color: #445566; }
.combo-input:focus { border-color: #4a90d9; }
.combo-input:disabled { opacity: 0.5; cursor: not-allowed; }

.combo-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: #1a2030;
  border: 1px solid #2a3a50;
  border-radius: 6px;
  list-style: none;
  margin: 0;
  padding: 4px 0;
  max-height: 200px;
  overflow-y: auto;
  z-index: 20;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
}
.combo-dropdown li {
  padding: 6px 12px;
  cursor: pointer;
  font-size: 13px;
  color: #c0cce0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left;
}
.combo-dropdown li:hover,
.combo-dropdown li.active { background: #253045; color: #fff; }

.combo-empty {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: #1a2030;
  border: 1px solid #2a3a50;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 13px;
  color: #445566;
  z-index: 20;
}
</style>

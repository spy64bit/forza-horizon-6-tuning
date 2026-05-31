<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  modelValue: { type: Number, default: 0 }, // selected tune id (0 = none)
  tunes:      { type: Array,  default: () => [] },
})
const emit = defineEmits(['update:modelValue'])

const nameOf = (id) => {
  const t = props.tunes.find((x) => x.id === id)
  return t ? (t.name || '(unnamed)') : ''
}

const search    = ref(nameOf(props.modelValue))
const open      = ref(false)
const highlight = ref(-1)

const filtered = computed(() => {
  if (!search.value) return props.tunes
  const f = search.value.toLowerCase()
  return props.tunes.filter((t) => (t.name || '').toLowerCase().includes(f))
})

watch(() => props.modelValue, (v) => {
  if (!open.value) search.value = nameOf(v)
})

function onFocus(e) {
  search.value = ''
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
  search.value = nameOf(props.modelValue)
}
function select(id) {
  emit('update:modelValue', id)
  search.value = nameOf(id)
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
    select(filtered.value[highlight.value].id)
  } else if (filtered.value.length === 1) {
    select(filtered.value[0].id)
  }
}
</script>

<template>
  <div class="combobox">
    <input
      class="combo-input"
      :value="search"
      placeholder="— No car tune —"
      autocomplete="off"
      @focus="onFocus"
      @input="onInput"
      @blur="onBlur"
      @keydown.escape.prevent="onBlur"
      @keydown.enter.prevent="onEnter"
      @keydown.down.prevent="onDown"
      @keydown.up.prevent="onUp"
    />
    <ul v-if="open" class="combo-dropdown">
      <li
        class="none"
        :class="{ active: highlight === -1 && modelValue === 0 }"
        @mousedown.prevent="select(0)"
      >— No car tune —</li>
      <li
        v-for="(t, i) in filtered"
        :key="t.id"
        :class="{ active: highlight === i }"
        @mousedown.prevent="select(t.id)"
        @mousemove="highlight = i"
      >{{ t.name || '(unnamed)' }}</li>
      <li v-if="filtered.length === 0" class="empty">No matches</li>
    </ul>
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
  max-height: 240px;
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

.combo-dropdown li.none {
  color: #6677aa;
  border-bottom: 1px solid #2a3a50;
  margin-bottom: 4px;
  padding-bottom: 8px;
}

.combo-dropdown li.empty {
  color: #445566;
  cursor: default;
}
.combo-dropdown li.empty:hover { background: transparent; color: #445566; }
</style>

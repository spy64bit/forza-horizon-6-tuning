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
  <div class="relative flex-1">
    <input
      class="bg-[#1a2030] border border-[#2a3a50] text-[#e0e6f0] px-2.5 py-1.5 rounded-md outline-none w-full text-[13px] cursor-text placeholder:text-[#445566] focus:border-[#4a90d9]"
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
    <ul v-if="open" class="absolute top-[calc(100%+4px)] left-0 right-0 bg-[#1a2030] border border-[#2a3a50] rounded-md list-none m-0 py-1 max-h-[240px] overflow-y-auto z-20 shadow-[0_4px_16px_rgba(0,0,0,0.4)]">
      <li
        class="px-3 py-1.5 cursor-pointer text-[13px] text-[#6677aa] whitespace-nowrap overflow-hidden text-ellipsis border-b border-[#2a3a50] mb-1 pb-2 hover:bg-[#253045] hover:text-white"
        :class="{ 'bg-[#253045] text-white': highlight === -1 && modelValue === 0 }"
        @mousedown.prevent="select(0)"
      >— No car tune —</li>
      <li
        v-for="(t, i) in filtered"
        :key="t.id"
        class="px-3 py-1.5 cursor-pointer text-[13px] text-[#c0cce0] whitespace-nowrap overflow-hidden text-ellipsis hover:bg-[#253045] hover:text-white"
        :class="{ 'bg-[#253045] text-white': highlight === i }"
        @mousedown.prevent="select(t.id)"
        @mousemove="highlight = i"
      >{{ t.name || '(unnamed)' }}</li>
      <li v-if="filtered.length === 0" class="px-3 py-1.5 text-[13px] text-[#445566] cursor-default">No matches</li>
    </ul>
  </div>
</template>

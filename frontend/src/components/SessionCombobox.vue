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
  <div class="relative flex-1">
    <input
      class="bg-[#1a2030] border border-[#2a3a50] text-[#e0e6f0] px-2.5 py-1.5 rounded-md outline-none w-full text-[13px] cursor-text placeholder:text-[#445566] focus:border-[#4a90d9] disabled:opacity-50 disabled:cursor-not-allowed"
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
    <ul v-if="open && filtered.length > 0" class="absolute top-[calc(100%+4px)] left-0 right-0 bg-[#1a2030] border border-[#2a3a50] rounded-md list-none m-0 py-1 max-h-[200px] overflow-y-auto z-20 shadow-[0_4px_16px_rgba(0,0,0,0.4)]">
      <li
        v-for="(s, i) in filtered"
        :key="s"
        class="px-3 py-1.5 cursor-pointer text-[13px] text-[#c0cce0] whitespace-nowrap overflow-hidden text-ellipsis hover:bg-[#253045] hover:text-white"
        :class="{ 'bg-[#253045] text-white': highlight === i }"
        @mousedown.prevent="select(s)"
      >{{ strip(s) }}</li>
    </ul>
    <div v-if="open && filtered.length === 0" class="absolute top-[calc(100%+4px)] left-0 right-0 bg-[#1a2030] border border-[#2a3a50] rounded-md px-3 py-2 text-[13px] text-[#445566] z-20">No matches</div>
  </div>
</template>

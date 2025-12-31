<!-- Custom styled select dropdown -->
<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ChevronDown } from 'lucide-vue-next'

const props = withDefaults(defineProps<{
  modelValue: string | number
  options: Array<{ value: string | number; label: string }>
  placeholder?: string
  disabled?: boolean
  size?: 'sm' | 'md' | 'lg'
  background?: string
}>(), {
  disabled: false,
  size: 'md'
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const isOpen = ref(false)
const selectRef = ref<HTMLDivElement>()

const selectedOption = computed(() => {
  return props.options.find(opt => opt.value === props.modelValue) || null
})

const selectOption = (value: string | number) => {
  emit('update:modelValue', value)
  isOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  if (selectRef.value && !selectRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div ref="selectRef" class="custom-select"
    :class="{ 'is-open': isOpen, 'is-disabled': disabled, [`size-${size}`]: true }">
    <button type="button" class="select-trigger" :disabled="disabled" :style="background ? { background } : {}"
      @click="!disabled && (isOpen = !isOpen)">
      <span class="select-value">
        {{ selectedOption?.label || placeholder || 'Select...' }}
      </span>
      <ChevronDown :size="16" class="select-icon" :class="{ 'is-open': isOpen }" />
    </button>
    <Transition name="dropdown">
      <div v-if="isOpen" class="select-dropdown">
        <button v-for="option in options" :key="String(option.value)" type="button" class="select-option"
          :class="{ 'is-selected': option.value === modelValue }" @click="selectOption(option.value)">
          {{ option.label }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.custom-select {
  position: relative;
  width: 100%;
}

.select-trigger {
  width: 100%;
  padding: 8px 12px;
  background: var(--bg-overlay);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--text-primary);
  font-family: inherit;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  transition: all var(--transition-fast);
  height: 40px;
  box-sizing: border-box;
}

.custom-select.size-sm .select-trigger {
  padding: 6px 10px;
  font-size: 12px;
  height: 36px;
}

.custom-select.size-lg .select-trigger {
  padding: 10px 14px;
  font-size: 14px;
}

.select-trigger:hover:not(:disabled) {
  border-color: var(--border-default);
  background: var(--bg-hover);
}

.select-trigger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.custom-select.is-open .select-trigger {
  border-color: var(--accent-primary);
  background: var(--bg-surface);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.select-value {
  flex: 1;
  text-align: left;
}

.select-icon {
  color: var(--text-secondary);
  transition: transform 0.2s;
  flex-shrink: 0;
}

.select-icon.is-open {
  transform: rotate(180deg);
}

.select-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  z-index: 1000;
  max-height: 200px;
  overflow-y: auto;
  margin: 0;
  padding: 4px;
}

/* Dropdown animation */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
}

.select-option {
  width: 100%;
  padding: 8px 10px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  font-family: inherit;
  cursor: pointer;
  text-align: left;
  transition: background-color var(--transition-fast);
}

.select-option:hover {
  background: var(--bg-hover);
}

.select-option.is-selected {
  background: var(--bg-overlay);
  color: var(--text-primary);
  font-weight: 500;
}
</style>
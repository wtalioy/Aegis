<!-- YAML Editor Component - Phase 4 -->
<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { configureMonaco, createYamlEditor } from '../../lib/monaco'
import type * as monaco from 'monaco-editor'

const props = defineProps<{
  value: string
  readOnly?: boolean
}>()

const emit = defineEmits<{
  'update:value': [value: string]
  change: [value: string]
}>()

const editorContainer = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null

onMounted(async () => {
  if (!editorContainer.value) return

  configureMonaco()
  editor = createYamlEditor(editorContainer.value, props.value)

  editor.onDidChangeModelContent(() => {
    const value = editor?.getValue() || ''
    emit('update:value', value)
    emit('change', value)
  })
})

watch(() => props.value, (newValue) => {
  if (editor && editor.getValue() !== newValue) {
    editor.setValue(newValue)
  }
})

onUnmounted(() => {
  editor?.dispose()
})
</script>

<template>
  <div class="yaml-editor">
    <div ref="editorContainer" class="editor-container"></div>
  </div>
</template>

<style scoped>
.yaml-editor {
  width: 100%;
  height: 100%;
  min-height: 400px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.editor-container {
  width: 100%;
  height: 100%;
  min-height: 400px;
}
</style>


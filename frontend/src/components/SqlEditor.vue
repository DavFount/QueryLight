<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { EditorView, basicSetup } from 'codemirror'
import { keymap } from '@codemirror/view'
import { sql, MySQL } from '@codemirror/lang-sql'
import { oneDark } from '@codemirror/theme-one-dark'

const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
  (event: 'run'): void
}>()
const host = ref<HTMLElement>()
let editor: EditorView | undefined
onMounted(() => {
  editor = new EditorView({
    parent: host.value,
    doc: props.modelValue,
    extensions: [
      keymap.of([
        {
          key: 'Mod-Enter',
          run: () => {
            emit('run')
            return true
          },
        },
        { key: 'Mod-s', run: () => true },
      ]),
      basicSetup,
      sql({ dialect: MySQL }),
      oneDark,
      EditorView.contentAttributes.of({ 'aria-label': 'SQL statement', spellcheck: 'false' }),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) emit('update:modelValue', update.state.doc.toString())
      }),
      EditorView.theme({
        '&': { height: '100%', background: 'transparent' },
        '.cm-scroller': { overflow: 'auto', fontFamily: 'Consolas, monospace', fontSize: '14px' },
        '.cm-gutters': { background: 'transparent', border: 'none' },
        '&.cm-focused': { outline: 'none' },
      }),
    ],
  })
})
watch(
  () => props.modelValue,
  (value) => {
    if (editor && editor.state.doc.toString() !== value)
      editor.dispatch({ changes: { from: 0, to: editor.state.doc.length, insert: value } })
  },
)
onBeforeUnmount(() => editor?.destroy())
</script>

<template><div ref="host" class="sql-editor" /></template>

<style scoped>
.sql-editor {
  height: 100%;
  min-height: 100px;
  padding: 12px 8px;
  box-sizing: border-box;
}
</style>

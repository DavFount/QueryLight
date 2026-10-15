<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { application } from '../../wailsjs/go/models'
import { pendingRows, type Drafts, type Cell } from '../features/table-drafts'
defineProps<{ drafts: Drafts; result: application.QueryResult }>()
const emit = defineEmits<{ close: [] }>()
const dialog = ref<HTMLDialogElement>()
const show = (cell: Cell) => cell.kind === 'null' ? 'NULL' : cell.value === '' ? '(empty text)' : cell.value
onMounted(() => dialog.value?.showModal())
</script>
<template>
  <dialog ref="dialog" class="confirm-dialog draft-review" aria-labelledby="draft-review-title" @cancel.prevent="emit('close')">
    <h2 id="draft-review-title">Review local drafts</h2>
    <p>{{ pendingRows(drafts) }} pending rows · {{ Object.keys(drafts).length }} changed cells. No database writes have been made. Apply Changes is not available yet.</p>
    <article v-for="(draft, key) in drafts" :key="key">
      <strong>{{ result.columns[draft.col].name }} · preview row {{ draft.row + 1 }}</strong>
      <details><summary>Original row key</summary><pre>{{ draft.identity }}</pre></details>
      <div class="draft-values"><div>Original<pre>{{ show(draft.original) }}</pre></div><div>Draft<pre>{{ show(draft.value) }}</pre></div></div>
    </article>
    <button autofocus @click="emit('close')">Back to table</button>
  </dialog>
</template>
<style scoped>
.draft-review { width: min(760px, 90vw); max-height: 80vh; overflow: auto; }
article { border-top: 1px solid var(--line); padding: 12px 0; }
.draft-values { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
pre { white-space: pre-wrap; overflow-wrap: anywhere; max-height: 140px; overflow: auto; user-select: text; }
dialog::backdrop { background: #080e16ba; }
</style>

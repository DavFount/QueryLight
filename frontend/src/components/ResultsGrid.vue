<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useVirtualizer } from '@tanstack/vue-virtual'
import type { application } from '../../wailsjs/go/models'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'
import { draftKey, draftReason, stageDraft, type Drafts } from '../features/table-drafts'

const props = defineProps<{ result: application.QueryResult; drafts?: Drafts }>()
const emit = defineEmits<{ draft: [row: number, col: number, text: string, isNull: boolean] }>()
const editing = ref(false)
const draftText = ref('')
const draftNull = ref(false)
const draftError = ref('')
const reason = computed(() => selected.value ? draftReason(props.result, selected.value.row, selected.value.col) : '')
const nullable = computed(() => props.result.table?.columns.find((c) => c.name === column.value?.name)?.nullable ?? false)
function visibleCell(row: number, col: number) {
  return props.drafts?.[draftKey(row, col)]?.value ?? props.result.rows[row][col]
}
function edit() {
  if (!cell.value || reason.value) return
  draftText.value = cell.value.kind === 'null' ? '' : cell.value.value
  draftNull.value = cell.value.kind === 'null'
  editing.value = true
  draftError.value = ''
}
function stage() {
  if (!selected.value) return
  try { stageDraft(props.result, props.drafts ?? {}, selected.value.row, selected.value.col, draftText.value, draftNull.value) }
  catch (error) { draftError.value = error instanceof Error ? error.message : 'Invalid draft.'; return }
  emit('draft', selected.value.row, selected.value.col, draftText.value, draftNull.value)
  editing.value = false
}
const scroll = ref<HTMLElement>()
const selected = ref<{ row: number; col: number }>()
const copyMessage = ref('')
const rowVirtualizer = useVirtualizer(
  computed(() => ({
    count: props.result.rows.length,
    getScrollElement: () => scroll.value ?? null,
    estimateSize: () => 34,
    overscan: 8,
    scrollMargin: 44,
  })),
)
const colVirtualizer = useVirtualizer(
  computed(() => ({
    horizontal: true,
    count: props.result.columns.length,
    getScrollElement: () => scroll.value ?? null,
    estimateSize: () => 210,
    overscan: 2,
  })),
)
const virtualRows = computed(() => rowVirtualizer.value.getVirtualItems())
const virtualColumns = computed(() => colVirtualizer.value.getVirtualItems())
const cell = computed(() =>
  selected.value ? visibleCell(selected.value.row, selected.value.col) : undefined,
)
const column = computed(() =>
  selected.value ? props.result.columns[selected.value.col] : undefined,
)
const gridWidth = computed(() => Math.max(colVirtualizer.value.getTotalSize(), 210))
function display(value: application.Cell) {
  if (value.unavailable) return `Unavailable · ${value.byteLength.toLocaleString()} bytes`
  if (value.kind === 'null') return 'NULL'
  if (value.kind === 'binary' || value.kind === 'unsupported')
    return value.value === '' ? '(empty bytes)' : `base64: ${value.value.slice(0, 60)}`
  return value.value === '' ? '(empty)' : value.value.slice(0, 160).replace(/\n/g, ' ↵ ')
}
async function copy() {
  if (!cell.value || cell.value.unavailable) return
  try {
    const text = cell.value.kind === 'null' ? 'NULL' : cell.value.value
    if ((window as unknown as { runtime?: unknown }).runtime) {
      if (!(await ClipboardSetText(text))) throw new Error('Clipboard unavailable')
    } else await navigator.clipboard.writeText(text)
    copyMessage.value = 'Copied'
  } catch {
    copyMessage.value = 'Select and copy the value below.'
  }
}
function select(row: number, col: number) {
  if (selected.value?.row !== row || selected.value?.col !== col) editing.value = false
  selected.value = { row, col }
  copyMessage.value = ''
}
function navigate(event: KeyboardEvent, row: number, col: number) {
  const changes: Record<string, [number, number]> = {
    ArrowDown: [1, 0],
    ArrowUp: [-1, 0],
    ArrowLeft: [0, -1],
    ArrowRight: [0, 1],
  }
  if (!changes[event.key]) return
  event.preventDefault()
  const [dr, dc] = changes[event.key]
  const nextRow = Math.max(0, Math.min(props.result.rows.length - 1, row + dr))
  const nextCol = Math.max(0, Math.min(props.result.columns.length - 1, col + dc))
  select(nextRow, nextCol)
  rowVirtualizer.value.scrollToIndex(nextRow, { align: 'auto' })
  colVirtualizer.value.scrollToIndex(nextCol, { align: 'auto' })
  requestAnimationFrame(() =>
    scroll.value?.querySelector<HTMLElement>(`[data-cell="${nextRow}-${nextCol}"]`)?.focus(),
  )
}
watch(
  () => props.result,
  () => {
    selected.value = undefined
    if (scroll.value) {
      scroll.value.scrollTop = 0
      scroll.value.scrollLeft = 0
    }
  },
)
</script>

<template>
  <div class="result-grid-wrap">
    <div v-if="!result.columns.length" class="empty-state compact">
      <span class="empty-symbol">✓</span>
      <h3>Statement complete</h3>
      <p>
        {{ result.affectedRows !== '' ? `${result.affectedRows} rows affected` : result.notice }}
      </p>
    </div>
    <template v-else>
      <div
        ref="scroll"
        class="grid-scroll"
        role="grid"
        :aria-label="result.table?.editCandidate ? 'Table results with local drafts' : 'Read-only results'"
        :aria-rowcount="result.rows.length + 1"
        :aria-colcount="result.columns.length"
      >
        <div class="grid-header" role="row" :style="{ width: `${gridWidth}px` }">
          <div
            v-for="col in virtualColumns"
            :key="col.index"
            class="grid-heading"
            role="columnheader"
            :aria-colindex="col.index + 1"
            :style="{ width: `${col.size}px`, left: `${col.start}px` }"
            :title="result.columns[col.index].type"
          >
            <strong>{{ result.columns[col.index].name }}</strong
            ><small>{{ result.columns[col.index].type }}</small>
          </div>
        </div>
        <div
          class="grid-body"
          :style="{ height: `${rowVirtualizer.getTotalSize()}px`, width: `${gridWidth}px` }"
        >
          <div
            v-for="row in virtualRows"
            :key="row.index"
            class="grid-row"
            role="row"
            :aria-rowindex="row.index + 2"
            :style="{
              height: `${row.size}px`,
              transform: `translateY(${row.start - 44}px)`,
              width: `${gridWidth}px`,
            }"
          >
            <div
              v-for="col in virtualColumns"
              :key="col.index"
              role="gridcell"
              :aria-colindex="col.index + 1"
              :tabindex="
                selected
                  ? selected.row === row.index && selected.col === col.index
                    ? 0
                    : -1
                  : row.index === 0 && col.index === 0
                    ? 0
                    : -1
              "
              :data-cell="`${row.index}-${col.index}`"
              class="grid-cell"
              :class="{
                selected: selected?.row === row.index && selected?.col === col.index,
                muted: visibleCell(row.index, col.index).kind === 'null',
                unavailable: result.rows[row.index][col.index].unavailable,
                changed: Boolean(drafts?.[draftKey(row.index, col.index)]),
              }"
              :style="{ width: `${col.size}px`, left: `${col.start}px` }"
              @click="select(row.index, col.index)"
              @focus="select(row.index, col.index)"
              @keydown="navigate($event, row.index, col.index)"
            >
              <span v-if="drafts?.[draftKey(row.index, col.index)]" aria-label="Changed cell">● </span>{{ display(visibleCell(row.index, col.index)) }}
            </div>
          </div>
        </div>
        <p v-if="!result.rows.length" class="no-rows">No rows returned.</p>
      </div>
      <div v-if="cell && column" class="value-inspector">
        <div class="inspector-title">
          <strong>{{ column.name }}</strong
          ><span
            >{{ cell.kind }} · {{ cell.byteLength.toLocaleString() }} bytes<span
              v-if="cell.kind === 'binary' || cell.kind === 'unsupported'"
            >
              · base64</span
            ></span
          ><button :disabled="cell.unavailable" @click="copy">Copy value</button
          ><button v-if="result.table?.editCandidate" :disabled="Boolean(reason)" :title="reason || 'Change a local draft; no database write'" @click="edit">Edit local draft</button
          ><button aria-label="Close value inspector" @click="selected = undefined">×</button>
        </div>
        <form v-if="editing" class="draft-editor" @submit.prevent="stage">
          <p>Local draft only. Database application is not available yet.</p>
          <label><input v-model="draftNull" type="checkbox" :disabled="!nullable" /> NULL</label>
          <label>Draft value<textarea v-model="draftText" aria-label="Draft value" :disabled="draftNull" maxlength="262144" rows="3" /></label>
          <button type="submit">Queue draft</button>
          <button type="button" @click="editing = false">Cancel edit</button>
          <p v-if="draftError" role="alert">{{ draftError }}</p>
        </form>
        <p v-else-if="cell.unavailable">
          Value exceeds the 256 KiB cell limit and was not transferred. Query a smaller projection
          to inspect it.
        </p>
        <pre v-else tabindex="0">{{
          cell.kind === 'null' ? 'NULL' : cell.value === '' ? '(empty value)' : cell.value
        }}</pre>
        <small aria-live="polite">{{ copyMessage }}</small>
        <small v-if="result.table?.editCandidate && reason">{{ reason }}</small>
      </div>
    </template>
  </div>
</template>

<style scoped>
.result-grid-wrap {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
.grid-scroll {
  overflow: auto;
  flex: 1;
  min-height: 100px;
  position: relative;
}
.grid-header {
  position: sticky;
  top: 0;
  height: 44px;
  z-index: 2;
  background: var(--panel);
  border-bottom: 1px solid var(--line);
}
.grid-heading {
  position: absolute;
  height: 44px;
  border-right: 1px solid var(--line);
  padding: 5px 12px;
  box-sizing: border-box;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.grid-heading strong {
  display: block;
  font-size: 12px;
}
.grid-heading small {
  font-size: 10px;
  color: var(--muted);
}
.grid-body {
  position: relative;
}
.grid-row {
  position: absolute;
  top: 0;
  left: 0;
}
.grid-row:nth-child(even) {
  background: #ffffff02;
}
.grid-cell {
  position: absolute;
  height: 34px;
  padding: 8px 12px;
  box-sizing: border-box;
  border-right: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
  font:
    12px Consolas,
    monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: default;
}
.grid-cell.selected {
  background: #62e4c315;
  box-shadow: inset 0 0 0 1px var(--mint);
}
.grid-cell.changed { background: #62e4c322; border-bottom: 2px solid var(--mint); }
.draft-editor textarea { display: block; width: 100%; font-family: Consolas, monospace; }
.draft-editor p { margin: 4px 0; color: var(--muted); }
.draft-editor button { margin: 6px 8px 0 0; }
.grid-cell.muted {
  font-style: italic;
  color: var(--muted);
}
.grid-cell.unavailable {
  color: var(--amber);
}
.no-rows {
  padding: 24px;
  color: var(--muted);
}
.value-inspector {
  border-top: 1px solid var(--line);
  background: var(--panel);
  padding: 10px 16px;
  max-height: 210px;
  overflow: auto;
}
.inspector-title {
  display: flex;
  align-items: center;
  gap: 12px;
}
.inspector-title span {
  color: var(--muted);
  font-size: 11px;
  flex: 1;
}
.value-inspector pre {
  font:
    12px Consolas,
    monospace;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  user-select: text;
}
.value-inspector small {
  color: var(--mint);
}
</style>

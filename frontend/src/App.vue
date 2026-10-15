<script setup lang="ts">
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import ResultsGrid from './components/ResultsGrid.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import DraftReview from './components/DraftReview.vue'
import { pendingRows, stageDraft, type Drafts } from './features/table-drafts'
import {
  db,
  checked,
  message,
  DatabaseError,
  type Config,
  type Connection,
  type Failure,
  type Result,
  type Table,
} from './services/database'

interface Tab {
  drafts: Drafts
  id: string
  kind: 'sql' | 'table'
  title: string
  schema: string
  table?: string
  editorID?: string
  sql: string
  mode: 'rows' | 'statement'
  result?: Result
  error?: Failure
  busy: boolean
  lost: boolean
  operation: string
  generation: number
}
const SqlEditor = defineAsyncComponent(() => import('./components/SqlEditor.vue'))
interface Workspace {
  connection: Connection
  schemas: string[]
  schema: string
  tables: Table[]
  tabs: Tab[]
  activeTab: string
  search: string
  metadataBusy: boolean
  metadataGeneration: number
  error: string
}
const workspaces = reactive<Workspace[]>([])
const activeConnection = ref('')
const workspace = computed(() => workspaces.find((w) => w.connection.id === activeConnection.value))
const activeTab = computed(() =>
  workspace.value?.tabs.find((t) => t.id === workspace.value?.activeTab),
)
const manager = ref(true)
const connectionBusy = ref(false)
const connectionError = ref('')
const connectionSuccess = ref('')
const openingTab = ref(false)
const disconnecting = ref(false)
const confirm = ref<{ title: string; body: string; action: () => void }>()
const reviewing = ref<Tab>()
function queueDraft(w: Workspace, tab: Tab, row: number, col: number, text: string, isNull: boolean) {
  if (!tab.result || tab.busy) return
  try {
    tab.drafts = stageDraft(tab.result, tab.drafts, row, col, text, isNull)
    w.error = ''
  } catch (error) { w.error = message(error) }
}
function discardDrafts(tab: Tab) {
  confirm.value = { title: 'Discard local drafts?', body: 'This removes the active table tab’s local changes. The database has not been changed.', action: () => { tab.drafts = {} } }
}
function refreshTable(w: Workspace, tab: Tab) {
  if (!pendingRows(tab.drafts)) { void run(w, tab); return }
  confirm.value = { title: 'Discard drafts and refresh?', body: 'Refreshing replaces this preview and discards its local drafts. Keep working to retain them.', action: () => { tab.drafts = {}; void run(w, tab) } }
}
const form = reactive<Config>({
  name: '',
  host: '127.0.0.1',
  port: 3306,
  user: '',
  password: '',
  database: '',
  tls: true,
})
const filteredTables = computed(
  () =>
    workspace.value?.tables.filter((t) =>
      t.name.toLowerCase().includes(workspace.value!.search.toLowerCase()),
    ) ?? [],
)
function selectConnection(id: string) {
  activeConnection.value = id
  manager.value = false
}
let alive = true
let sqlNumber = 0
const bridgeAvailable = Boolean((window as unknown as { go?: unknown }).go)

function preset(engine: 'mariadb' | 'mysql') {
  Object.assign(form, {
    name: engine === 'mariadb' ? 'Local MariaDB' : 'Local MySQL',
    host: '127.0.0.1',
    port: engine === 'mariadb' ? 13306 : 13307,
    user: 'querylight',
    password: 'querylight-local-only',
    database: 'querylight_test',
    tls: false,
  })
  connectionError.value = ''
  connectionSuccess.value = ''
}
async function connect(testOnly: boolean) {
  if (connectionBusy.value) return
  connectionBusy.value = true
  connectionError.value = ''
  connectionSuccess.value = ''
  try {
    const reply = checked(
      await (testOnly ? db.TestConnection({ ...form }) : db.Connect({ ...form })),
    )
    if (!reply.connection) throw new Error('The server did not return a connection.')
    if (testOnly) {
      connectionSuccess.value = `Connection verified · ${reply.connection.version}`
      return
    }
    const w: Workspace = reactive({
      connection: reply.connection,
      schemas: [],
      schema: reply.connection.database,
      tables: [],
      tabs: [],
      activeTab: '',
      search: '',
      metadataBusy: false,
      metadataGeneration: 0,
      error: '',
    })
    workspaces.push(w)
    activeConnection.value = w.connection.id
    manager.value = false
    form.password = ''
    await loadSchemas(w)
    await newSQL(w)
  } catch (error) {
    connectionError.value = message(error)
  } finally {
    connectionBusy.value = false
  }
}
async function loadSchemas(w: Workspace) {
  const generation = ++w.metadataGeneration
  w.metadataBusy = true
  w.error = ''
  try {
    const reply = checked(await db.ListSchemas(w.connection.id))
    if (generation !== w.metadataGeneration || !workspaces.includes(w)) return
    w.schemas = reply.schemas ?? []
    if (!w.schemas.includes(w.schema))
      w.schema =
        w.schemas.find(
          (s) => !['information_schema', 'performance_schema', 'mysql', 'sys'].includes(s),
        ) ??
        w.schemas[0] ??
        ''
  } catch (error) {
    if (generation === w.metadataGeneration) w.error = message(error)
  } finally {
    if (generation === w.metadataGeneration) w.metadataBusy = false
  }
  if (w.schema && generation === w.metadataGeneration) await loadTables(w)
}
async function loadTables(w: Workspace) {
  const generation = ++w.metadataGeneration
  const schema = w.schema
  w.metadataBusy = true
  w.error = ''
  w.tables = []
  try {
    const reply = checked(await db.ListTables(w.connection.id, schema))
    if (generation === w.metadataGeneration) w.tables = reply.tables ?? []
  } catch (error) {
    if (generation === w.metadataGeneration) w.error = message(error)
  } finally {
    if (generation === w.metadataGeneration) w.metadataBusy = false
  }
}
function makeTab(kind: 'sql' | 'table', schema: string): Tab {
  return reactive({
    id: crypto.randomUUID(),
    drafts: {},
    kind,
    title: '',
    schema,
    sql: 'SELECT *\nFROM value_samples\nLIMIT 100;',
    mode: 'rows',
    busy: false,
    lost: false,
    operation: '',
    generation: 0,
  })
}
async function newSQL(w = workspace.value) {
  if (!w || openingTab.value) return
  if (w.tabs.length >= 8) {
    w.error = 'Close a tab before opening another (maximum 8 per workspace).'
    return
  }
  openingTab.value = true
  const schema = w.schema
  try {
    const reply = checked(await db.OpenEditor(w.connection.id, schema))
    if (!workspaces.includes(w)) {
      await db.CloseEditor(reply.id)
      return
    }
    const tab = makeTab('sql', schema)
    tab.editorID = reply.id
    tab.title = `SQL ${++sqlNumber}`
    if (schema !== 'querylight_test') tab.sql = 'SELECT VERSION();'
    w.tabs.push(tab)
    w.activeTab = tab.id
  } catch (error) {
    w.error = message(error)
  } finally {
    openingTab.value = false
  }
}
async function openTable(table: Table) {
  const w = workspace.value
  if (!w) return
  const existing = w.tabs.find(
    (t) => t.kind === 'table' && t.schema === w.schema && t.table === table.name,
  )
  if (existing) {
    w.activeTab = existing.id
    return
  }
  if (w.tabs.length >= 8) {
    w.error = 'Close a tab before opening another (maximum 8 per workspace).'
    return
  }
  const tab = makeTab('table', w.schema)
  tab.title = table.name
  tab.table = table.name
  w.tabs.push(tab)
  w.activeTab = tab.id
  await run(w, tab)
}
async function run(w: Workspace, tab: Tab) {
  if (pendingRows(tab.drafts)) return
  if (tab.busy || tab.lost || (tab.kind === 'sql' && !tab.sql.trim())) return
  const generation = ++tab.generation
  tab.busy = true
  tab.error = undefined
  tab.result = undefined
  try {
    if (tab.operation) {
      await db.ReleaseOperation(tab.operation)
      tab.operation = ''
    }
    const reply = checked(
      await (tab.kind === 'table'
        ? db.StartInspection(w.connection.id, tab.schema, tab.table!)
        : db.StartExecution({ editorID: tab.editorID!, sql: tab.sql, mode: tab.mode })),
    )
    if (!alive || generation !== tab.generation || !w.tabs.includes(tab)) {
      await db.ReleaseOperation(reply.id)
      return
    }
    tab.operation = reply.id
    while (alive && generation === tab.generation && w.tabs.includes(tab)) {
      const operation = await db.GetOperation(reply.id)
      if (generation !== tab.generation || !w.tabs.includes(tab)) return
      if (operation.status !== 'running') {
        tab.result = operation.result ?? undefined
        tab.error = operation.error ?? undefined
        tab.lost = Boolean(operation.error?.sessionLost)
        break
      }
      await new Promise((resolve) => setTimeout(resolve, 180))
    }
  } catch (error) {
    if (generation === tab.generation) {
      tab.error =
        error instanceof DatabaseError
          ? error.details
          : { code: 'bridge', message: message(error), sessionLost: false, outcomeUnknown: true }
      tab.lost = tab.error.sessionLost
      if (tab.operation) void db.CancelOperation(tab.operation).catch(() => {})
    }
  } finally {
    if (generation === tab.generation) tab.busy = false
  }
}
async function cancel(tab: Tab) {
  try {
    checked(await db.CancelOperation(tab.operation))
  } catch (error) {
    const w = workspace.value
    if (w) w.error = message(error)
  }
}
function requestClose(w: Workspace, tab: Tab) {
  const action = () => {
    void closeTab(w, tab)
  }
  if (tab.busy || tab.kind === 'sql' || pendingRows(tab.drafts))
    confirm.value = {
      title: `Close ${tab.title}?`,
      body: pendingRows(tab.drafts) ? 'This discards this table tab’s local drafts. No draft changes have been written to the database.' : tab.busy
        ? 'This cancels running work and closes the session. A write may already have executed.'
        : 'SQL text is kept in memory only and will be discarded when this tab closes.',
      action,
    }
  else action()
}
async function closeTab(w: Workspace, tab: Tab) {
  ++tab.generation
  w.tabs = w.tabs.filter((t) => t.id !== tab.id)
  if (w.activeTab === tab.id) w.activeTab = w.tabs.at(-1)?.id ?? ''
  try {
    if (tab.operation) await db.ReleaseOperation(tab.operation)
    if (tab.editorID) await db.CloseEditor(tab.editorID)
  } catch (error) {
    w.error = message(error)
  }
}
function requestDisconnect(w: Workspace) {
  confirm.value = {
    title: `Disconnect ${w.connection.name}?`,
    body: 'This closes all tabs, discards their SQL text and local table drafts, and cancels running work. Nothing is saved to disk.',
    action: () => {
      void disconnect(w)
    },
  }
}
async function disconnect(w: Workspace) {
  disconnecting.value = true
  try {
    checked(await db.Disconnect(w.connection.id))
    for (const t of w.tabs) ++t.generation
    workspaces.splice(workspaces.indexOf(w), 1)
    if (activeConnection.value === w.connection.id)
      activeConnection.value = workspaces[0]?.connection.id ?? ''
    if (!workspaces.length) manager.value = true
  } catch (error) {
    w.error = message(error)
  } finally {
    disconnecting.value = false
  }
}
function acceptConfirm() {
  const action = confirm.value?.action
  confirm.value = undefined
  action?.()
}
function keydown(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') event.preventDefault()
  if (event.key === 'Escape') confirm.value = undefined
}
onMounted(() => window.addEventListener('keydown', keydown))
onBeforeUnmount(() => {
  alive = false
  window.removeEventListener('keydown', keydown)
})
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <a class="brand" href="#" @click.prevent="manager = true"
        ><img src="/favicon.svg" alt="" /><strong>Querylight</strong
        ><span class="brand-caption">SQL, with clarity.</span></a
      >
      <div class="header-actions">
        <template v-if="workspaces.length"
          ><select
            v-model="activeConnection"
            aria-label="Active connection"
            @change="manager = false"
          >
            <option v-for="w in workspaces" :key="w.connection.id" :value="w.connection.id">
              {{ w.connection.name }}
            </option>
          </select></template
        ><button @click="manager = !manager">
          {{ manager && workspace ? 'Back to workspace' : 'Connections' }}
        </button>
      </div>
    </header>

    <main v-if="manager || !workspace" class="connection-manager">
      <aside class="connection-intro">
        <span class="eyebrow">YOUR DATABASE WORKSPACE</span>
        <h1>Less noise.<br /><span>More clarity.</span></h1>
        <p>Connect, explore, and work with your data in one focused place.</p>
        <div v-if="workspaces.length" class="connection-list">
          <h3>Open connections</h3>
          <button
            v-for="w in workspaces"
            :key="w.connection.id"
            @click="selectConnection(w.connection.id)"
          >
            <span class="status-dot" /><span
              ><strong>{{ w.connection.name }}</strong
              ><small>{{ w.connection.address }}</small></span
            ><span>↗</span>
          </button>
        </div>
        <div class="intro-note">
          <span>◇</span>
          <p>Passwords stay in memory.<br />Saved connections are coming later.</p>
        </div>
      </aside>
      <section class="connection-card">
        <div class="card-heading">
          <span class="eyebrow">CONNECTION MANAGER</span>
          <h2>Make a connection</h2>
          <p>MariaDB or MySQL. Your next query starts here.</p>
        </div>
        <div class="preset-row">
          <span>Local test databases</span
          ><button :disabled="connectionBusy" @click="preset('mariadb')">MariaDB :13306</button
          ><button :disabled="connectionBusy" @click="preset('mysql')">MySQL :13307</button>
        </div>
        <form @submit.prevent="connect(false)">
          <fieldset :disabled="connectionBusy">
            <label
              >Connection name
              <input
                v-model="form.name"
                maxlength="100"
                placeholder="e.g. Development"
                autocomplete="off"
            /></label>
            <div class="form-row host-row">
              <label
                >Host
                <input v-model="form.host" required maxlength="255" autocomplete="off" /></label
              ><label
                >Port <input v-model.number="form.port" required type="number" min="1" max="65535"
              /></label>
            </div>
            <div class="form-row">
              <label
                >Username
                <input v-model="form.user" required maxlength="256" autocomplete="off" /></label
              ><label
                >Password
                <input v-model="form.password" type="password" maxlength="4096" autocomplete="off"
              /></label>
            </div>
            <label
              >Default database <span class="optional">optional</span
              ><input
                v-model="form.database"
                maxlength="256"
                placeholder="Choose after connecting"
                autocomplete="off"
            /></label>
            <label class="checkbox-label"
              ><input v-model="form.tls" type="checkbox" /><span
                >Use verified TLS
                <small>Requires a server certificate trusted by Windows.</small></span
              ></label
            >
          </fieldset>
          <p v-if="!bridgeAvailable" class="banner error" role="alert">
            Open Querylight with wails dev or the desktop executable to connect.
          </p>
          <p v-if="connectionError" class="banner error" role="alert">{{ connectionError }}</p>
          <p v-if="connectionSuccess" class="banner success" role="status">
            {{ connectionSuccess }}
          </p>
          <div class="form-actions">
            <button
              type="button"
              :disabled="connectionBusy || !bridgeAvailable"
              @click="connect(true)"
            >
              {{ connectionBusy ? 'Connecting…' : 'Test Connection' }}</button
            ><button class="primary" type="submit" :disabled="connectionBusy || !bridgeAvailable">
              {{ connectionBusy ? 'Please wait…' : 'Connect →' }}
            </button>
          </div>
        </form>
      </section>
    </main>

    <main v-else class="workspace">
      <aside class="explorer">
        <div class="explorer-heading">
          <span class="eyebrow">EXPLORER</span
          ><button
            class="icon-button"
            aria-label="Refresh schemas and tables"
            :disabled="workspace.metadataBusy"
            @click="loadSchemas(workspace)"
          >
            ↻
          </button>
        </div>
        <label class="schema-select"
          >Database<select
            v-model="workspace.schema"
            :disabled="workspace.metadataBusy"
            @change="loadTables(workspace)"
          >
            <option v-for="schema in workspace.schemas" :key="schema">{{ schema }}</option>
          </select></label
        >
        <input
          v-model="workspace.search"
          class="object-search"
          aria-label="Search tables"
          placeholder="Search tables…"
        />
        <div class="object-heading">
          <span>TABLES & VIEWS</span><span>{{ filteredTables.length }}</span>
        </div>
        <div class="table-list">
          <p v-if="workspace.metadataBusy" class="sidebar-note" role="status">Loading objects…</p>
          <p v-else-if="!filteredTables.length" class="sidebar-note">
            {{ workspace.search ? 'No matching objects.' : 'No tables available.' }}
          </p>
          <button
            v-for="table in filteredTables"
            :key="table.name"
            :class="{
              active: activeTab?.table === table.name && activeTab.schema === workspace.schema,
            }"
            @click="openTable(table)"
          >
            <span class="table-glyph">{{ table.type === 'VIEW' ? '◈' : '▦' }}</span
            ><span>{{ table.name }}</span>
          </button>
        </div>
        <div class="explorer-footer">
          <strong>{{ workspace.connection.name }}</strong
          ><small>{{ workspace.connection.address }}</small
          ><button :disabled="disconnecting" @click="requestDisconnect(workspace)">
            Disconnect
          </button>
        </div>
      </aside>
      <section class="work-area">
        <div v-if="workspace.error" class="banner error workspace-error" role="alert">
          {{ workspace.error
          }}<button aria-label="Dismiss error" @click="workspace.error = ''">×</button>
        </div>
        <div class="tab-bar" role="tablist" aria-label="Workspace tabs">
          <div
            v-for="tab in workspace.tabs"
            :key="tab.id"
            class="tab"
            :class="{ active: tab.id === workspace.activeTab }"
          >
            <button
              role="tab"
              :aria-selected="tab.id === workspace.activeTab"
              :title="`${tab.schema || 'No default database'} · ${tab.title}`"
              @click="workspace.activeTab = tab.id"
            >
              <span>{{ tab.kind === 'sql' ? '⌘' : '▦' }}</span
              >{{ tab.title }}<span v-if="pendingRows(tab.drafts)"> · {{ pendingRows(tab.drafts) }} pending</span><span v-if="tab.busy" class="running-dot" /></button
            ><button
              class="tab-close"
              :aria-label="`Close ${tab.title}`"
              @click="requestClose(workspace, tab)"
            >
              ×
            </button>
          </div>
          <button
            class="new-tab"
            :disabled="openingTab"
            aria-label="New SQL editor"
            @click="newSQL()"
          >
            ＋ SQL
          </button>
        </div>
        <template v-if="activeTab"
          ><div class="context-bar">
            <span
              >{{ workspace.connection.name }}<span class="breadcrumb"> / </span
              >{{ activeTab.schema || 'No default database'
              }}<span v-if="activeTab.table"
                ><span class="breadcrumb"> / </span>{{ activeTab.table }}</span
              ></span
            ><span class="context-hint">{{
              activeTab.kind === 'table' ? 'READ-ONLY PREVIEW' : 'DEDICATED SQL SESSION'
            }}</span>
          </div>
          <template v-if="activeTab.kind === 'sql'"
            ><div class="editor-toolbar">
              <span>SQL editor</span>
              <div>
                <label class="mode-label"
                  >Run as
                  <select
                    v-model="activeTab.mode"
                    :disabled="activeTab.busy"
                    title="Rows returns a result grid; Statement reports affected rows. Both can modify data."
                  >
                    <option value="rows">Rows</option>
                    <option value="statement">Statement</option>
                  </select></label
                ><button
                  v-if="activeTab.busy"
                  :disabled="!activeTab.operation"
                  @click="cancel(activeTab)"
                >
                  Cancel</button
                ><button
                  v-else
                  class="primary"
                  :disabled="activeTab.lost || !activeTab.sql.trim()"
                  @click="run(workspace, activeTab)"
                >
                  ▶ Run <kbd>Ctrl ↵</kbd>
                </button>
              </div>
            </div>
            <div class="editor-pane">
              <SqlEditor
                :key="activeTab.id"
                v-model="activeTab.sql"
                @run="run(workspace!, activeTab!)"
              />
            </div>
          </template>
          <div v-else class="table-toolbar">
            <strong>{{ activeTab.title }}</strong
            ><span>{{ activeTab.result?.table?.editCandidate ? 'Text drafts only · Apply Changes not available yet.' : activeTab.result?.table?.readOnlyReason || 'Read-only preview · checking table metadata.' }}</span
            ><button
              v-if="activeTab.busy"
              :disabled="!activeTab.operation"
              @click="cancel(activeTab)"
            >
              Cancel</button
            ><button v-else @click="refreshTable(workspace, activeTab)">↻ Refresh</button>
          </div>
          <div class="results-header">
            <strong>{{ activeTab.kind === 'table' ? 'Data' : 'Results' }}</strong
            ><span v-if="activeTab.busy" role="status">Running… · 30-second timeout</span
            ><span v-else-if="activeTab.result"
              >{{
                activeTab.result.affectedRows !== ''
                  ? `${activeTab.result.affectedRows} affected`
                  : `${activeTab.result.rows.length} rows fetched`
              }}
              · {{ activeTab.result.elapsedMS }} ms</span
            ><span v-else>Ready when you are</span
            ><span v-if="activeTab.result?.truncated" class="limit-badge"
              >Result limit reached</span
            >
          </div>
          <p v-if="activeTab.error" class="banner error" role="alert">
            {{ activeTab.error.message
            }}<button v-if="activeTab.lost" @click="newSQL()">New SQL session</button>
          </p>
          <p
            v-if="activeTab.result?.notice"
            class="result-notice"
          >
            {{ activeTab.result.notice }}
          </p>
          <ResultsGrid
            v-if="activeTab.result"
            :key="`${activeTab.id}-${activeTab.generation}`"
            :result="activeTab.result"
            :drafts="activeTab.drafts"
            @draft="(row, col, text, isNull) => queueDraft(workspace!, activeTab!, row, col, text, isNull)"
          />
          <div v-else class="empty-state">
            <span class="empty-symbol">{{
              activeTab.busy ? '◌' : activeTab.error ? '!' : '⌘'
            }}</span>
            <h3>
              {{
                activeTab.busy
                  ? 'Working on your query'
                  : activeTab.error
                    ? 'The operation did not complete'
                    : 'A little SQL. A clear answer.'
              }}
            </h3>
            <p>
              {{
                activeTab.busy
                  ? 'You can keep editing or switch tabs while the server works.'
                  : activeTab.error
                    ? 'Your SQL text is still here. Review the message above.'
                    : 'Run a statement to see results here. Rows and Statement modes can both change data.'
              }}
            </p>
          </div>
          <div v-if="activeTab.kind === 'table' && pendingRows(activeTab.drafts)" class="draft-bar">
            <span role="status">{{ pendingRows(activeTab.drafts) }} pending rows · {{ Object.keys(activeTab.drafts).length }} changed cells · local only</span>
            <button @click="reviewing = activeTab">Review Changes</button>
            <button @click="discardDrafts(activeTab)">Discard Changes</button>
            <button disabled title="Transactional apply is not implemented yet">Apply Changes (unavailable)</button>
          </div>
        </template>
        <div v-else class="empty-state">
          <span class="empty-symbol">▦</span>
          <h2>Your database, in focus.</h2>
          <p>Open a table from the explorer or start a SQL query.</p>
          <button class="primary" :disabled="openingTab" @click="newSQL()">New SQL editor</button>
        </div>
      </section>
    </main>
    <footer class="status-bar">
      <span
        ><span class="status-dot" :class="{ offline: !workspace }" />{{
          workspace ? `Connected · ${workspace.connection.version}` : 'Not connected'
        }}</span
      ><span v-if="workspace"
        >{{ workspace.connection.tls ? 'Verified TLS' : 'TLS disabled'
        }}<span class="status-separator">·</span>500 rows / 5 MiB per result</span
      ><span v-else>MariaDB + MySQL</span>
    </footer>
    <ConfirmDialog
      v-if="confirm"
      :title="confirm.title"
      :body="confirm.body"
      @cancel="confirm = undefined"
      @accept="acceptConfirm"
    />
    <DraftReview v-if="reviewing?.result" :drafts="reviewing.drafts" :result="reviewing.result" @close="reviewing = undefined" />
  </div>
</template>

<style scoped>
.draft-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; padding: 8px 16px; border-top: 1px solid var(--line); background: var(--panel); }
.draft-bar span { flex: 1; font-size: 12px; }
</style>

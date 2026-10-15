import { expect, test, type Page } from '@playwright/test'

// Deliberate browser-only Wails mock. Go integration tests exercise real engines;
// these checks cover UI state, virtualization, focus, and stale response guards.
async function bridge(page: Page) {
  await page.addInitScript(() => {
    let next = 0
    const operations = new Map<string, any>()
    const editors = new Map<string, string>()
    const cell = (value: string, kind = 'text') => ({
      kind,
      value,
      unavailable: false,
      byteLength: value.length,
    })
    const column = (name: string, type = 'VARCHAR') => ({
      name,
      type,
      nullable: false,
      nullableKnown: true,
    })
    const rows = (count: number) => ({
      columns: [column('id', 'INT'), column('label')],
      rows: Array.from({ length: count }, (_, i) => [
        cell(String(i + 1), 'integer'),
        cell(`Synthetic row ${i + 1}`),
      ]),
      truncated: count === 500,
      notice: '',
      affectedRows: '',
      elapsedMS: 12,
    })
    const connection = (cfg: any) => ({
      id: `connection-${++next}`,
      name: cfg.name,
      address: `${cfg.host}:${cfg.port}`,
      version: cfg.port === 13306 ? '11.4.12-MariaDB' : '8.4.11',
      database: cfg.database,
      tls: cfg.tls,
    })
    const start = (result: any, slow = false) => {
      const id = `operation-${++next}`
      operations.set(id, {
        id,
        status: slow ? 'running' : 'succeeded',
        result: slow ? null : result,
        error: null,
      })
      return { id, error: null }
    }
    const App = {
      TestConnection: async (cfg: any) => ({ connection: connection(cfg), error: null }),
      Connect: async (cfg: any) =>
        cfg.password === 'bad'
          ? {
              connection: null,
              error: {
                code: 'authentication',
                message: 'Authentication failed.',
                sessionLost: false,
                outcomeUnknown: false,
              },
            }
          : { connection: connection(cfg), error: null },
      Disconnect: async () => ({ error: null }),
      ListSchemas: async () => ({
        schemas: ['information_schema', 'querylight_test'],
        tables: [],
        error: null,
      }),
      ListTables: async () => ({
        schemas: [],
        tables: ['grid_rows', 'value_samples', 'empty_table', 'wide_table', 'large_cells'].map(
          (name) => ({ name, type: 'BASE TABLE' }),
        ),
        error: null,
      }),
      OpenEditor: async (_id: string, schema: string) => {
        if ((window as any).__delayEditor) {
          ;(window as any).__openingSchema = schema
          await new Promise<void>((resolve) => { (window as any).__finishEditor = resolve })
        }
        const id = `editor-${++next}`
        editors.set(id, schema)
        return { id, error: null }
      },
      CloseEditor: async (id: string) => {
        editors.delete(id)
        return { error: null }
      },
      StartInspection: async (_id: string, _schema: string, table: string) => {
        if (table === 'wide_table')
          return start({
            ...rows(1),
            columns: Array.from({ length: 100 }, (_, i) => column(`column_${i}`)),
            rows: [Array.from({ length: 100 }, (_, i) => cell(`value_${i}`))],
          })
        if (table === 'large_cells')
          return start({
            ...rows(1),
            rows: [[cell('1', 'integer'), { ...cell(''), unavailable: true, byteLength: 300000 }]],
          })
        if (table === 'value_samples')
          return start({
            ...rows(1),
            columns: [
              column('unsigned_bigint', 'UNSIGNED BIGINT'),
              column('exact_decimal', 'DECIMAL'),
            ],
            rows: [
              [
                cell('18446744073709551615', 'integer'),
                cell('12345678901234567890.1234500000', 'decimal'),
              ],
            ],
          })
        return start({
          ...rows(table === 'empty_table' ? 0 : 500),
          table: { engine: 'InnoDB', primaryKey: ['id'], columns: [], editCandidate: true,
            readOnlyReason: 'Queued editing is not implemented yet; this preview is read-only.' },
        })
      },
      StartExecution: async (req: any) => {
        if (req.sql.includes('DELAY_START'))
          await new Promise((resolve) => setTimeout(resolve, 500))
        return start(
          req.mode === 'statement'
            ? { ...rows(0), columns: [], affectedRows: '2', notice: 'Statement executed. No automatic retry.' }
            : req.sql.includes('NO_COLUMNS')
              ? { ...rows(0), columns: [], notice: 'No result columns. SQL has already run; it was not retried.' }
              : rows(2),
          req.sql.includes('SLEEP'),
        )
      },
      GetOperation: async (id: string) => operations.get(id),
      CancelOperation: async (id: string) => {
        operations.set(id, {
          id,
          status: 'cancelled',
          result: null,
          error: {
            code: 'cancelled',
            message: 'Operation cancelled. Editor session lost. A write may have executed.',
            sessionLost: true,
            outcomeUnknown: true,
          },
        })
        return { error: null }
      },
      ReleaseOperation: async (id: string) => {
        operations.delete(id)
        ;(window as any).__released = ((window as any).__released ?? 0) + 1
        return { error: null }
      },
    }
    ;(window as any).go = { main: { App } }
    ;(window as any).runtime = {
      ClipboardSetText: async (value: string) => {
        ;(window as any).__copied = value
        return true
      },
    }
  })
}

async function connect(page: Page) {
  await page.getByRole('button', { name: 'MariaDB :13306' }).click()
  await page.getByRole('button', { name: 'Connect →', exact: true }).click()
  await expect(page.getByRole('tab', { name: /SQL 1/ })).toBeVisible()
}

test.beforeEach(async ({ page }) => {
  await bridge(page)
  await page.goto('/')
})

test('connection manager layout', async ({ page }) => {
  await expect(page.getByRole('heading', { name: 'Make a connection' })).toBeVisible()
  await page.screenshot({ path: 'test-results/connection-manager.png', fullPage: true })
})

test('opening SQL retains the requested database when explorer selection changes', async ({ page }) => {
  await connect(page)
  await page.evaluate(() => { (window as any).__delayEditor = true })
  await page.getByRole('button', { name: 'New SQL editor', exact: true }).click()
  await expect.poll(() => page.evaluate(() => (window as any).__openingSchema)).toBe('querylight_test')
  await page.getByRole('combobox', { name: 'Database', exact: true }).selectOption('information_schema')
  await page.evaluate(() => (window as any).__finishEditor())
  await expect(page.getByRole('tab', { name: /SQL 2/ })).toHaveAttribute('title', 'querylight_test · SQL 2')
  await expect(page.getByRole('textbox', { name: 'SQL statement' })).toContainText('value_samples')
  await expect(page.getByRole('combobox', { name: 'Database', exact: true })).toHaveValue('information_schema')
})

test('zero-column results retain execution feedback without implying a retry', async ({ page }) => {
  await connect(page)
  await page.getByRole('textbox', { name: 'SQL statement' }).fill('DO 1 /* NO_COLUMNS */')
  await page.getByRole('button', { name: /▶ Run/ }).click()
  await expect(page.locator('.result-notice')).toContainText('SQL has already run; it was not retried.')
  await page.getByLabel('Run as').selectOption('statement')
  await page.getByRole('button', { name: /▶ Run/ }).click()
  await expect(page.locator('.result-notice')).toHaveText('Statement executed. No automatic retry.')
})

test('connection feedback, SQL execution and cancellation preserve text', async ({ page }) => {
  await page.getByRole('button', { name: 'MariaDB :13306' }).click()
  await page.getByRole('textbox', { name: 'Password', exact: true }).fill('bad')
  await page.getByRole('button', { name: 'Connect →', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Authentication failed')
  await connect(page)
  await page.getByRole('button', { name: /▶ Run/ }).click()
  await expect(page.getByRole('grid')).toBeVisible()
  await expect(page.getByText('2 rows fetched')).toBeVisible()
  const editor = page.getByRole('textbox', { name: 'SQL statement' })
  await editor.fill('SELECT SLEEP(15);')
  await editor.press('Control+Enter')
  await expect(page.getByRole('button', { name: 'Cancel', exact: true })).toBeEnabled()
  await page.getByRole('button', { name: 'Cancel', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('A write may have executed')
  await expect(editor).toHaveText('SELECT SLEEP(15);')
  await expect(page.getByRole('button', { name: /▶ Run/ })).toBeDisabled()
  await page.getByRole('button', { name: 'New SQL session', exact: true }).click()
  await expect(page.getByRole('tab', { name: /SQL 2/ })).toBeVisible()
})

test('virtualizes rows and columns and inspects exact values', async ({ page }) => {
  await connect(page)
  await page.getByRole('button', { name: '▦ grid_rows', exact: true }).click()
  await expect(page.getByRole('grid')).toBeVisible()
  await expect(page.getByText('Result limit reached')).toBeVisible()
  await expect(page.locator('.table-toolbar')).toContainText('this preview is read-only')
  expect(await page.getByRole('gridcell').count()).toBeLessThan(100)
  await page.locator('.grid-scroll').evaluate((el) => {
    el.scrollTop = el.scrollHeight
  })
  await expect(page.getByRole('gridcell', { name: 'Synthetic row 500', exact: true })).toBeVisible()
  await page.getByRole('button', { name: '▦ wide_table', exact: true }).click()
  expect(await page.getByRole('columnheader').count()).toBeLessThan(15)
  await page.locator('.grid-scroll').evaluate((el) => {
    el.scrollLeft = el.scrollWidth
  })
  await expect(page.getByRole('gridcell', { name: 'value_99', exact: true })).toBeVisible()
  await page.getByRole('button', { name: '▦ value_samples', exact: true }).click()
  await page.getByRole('gridcell', { name: '18446744073709551615', exact: true }).click()
  await expect(page.locator('.value-inspector pre')).toHaveText('18446744073709551615')
  await page.getByRole('gridcell', { name: '12345678901234567890.1234500000', exact: true }).click()
  await expect(page.locator('.value-inspector pre')).toHaveText('12345678901234567890.1234500000')
  await page.getByRole('button', { name: 'Copy value', exact: true }).click()
  await expect(page.locator('.value-inspector')).toContainText('Copied')
  expect(await page.evaluate(() => (window as any).__copied)).toBe(
    '12345678901234567890.1234500000',
  )
  await page.screenshot({ path: 'test-results/workspace.png', fullPage: true })
})

test('empty and oversized results and guarded tab closing', async ({ page }) => {
  await connect(page)
  await page.getByRole('button', { name: '▦ empty_table', exact: true }).click()
  await expect(page.getByText('No rows returned.')).toBeVisible()
  await page.getByRole('button', { name: '▦ large_cells', exact: true }).click()
  await page.getByRole('gridcell', { name: /Unavailable/ }).click()
  await expect(page.getByRole('button', { name: 'Copy value' })).toBeDisabled()
  await page.getByRole('button', { name: 'Close SQL 1', exact: true }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.getByRole('button', { name: 'Keep working', exact: true }).click()
  await expect(page.getByRole('tab', { name: /SQL 1/ })).toBeVisible()
  await page.getByRole('button', { name: 'Close SQL 1', exact: true }).click()
  await page.getByRole('button', { name: 'Continue', exact: true }).click()
  await expect(page.getByRole('tab', { name: /SQL 1/ })).toHaveCount(0)
})

test('connection workspaces retain independent tabs and statement mode reports writes', async ({
  page,
}) => {
  await connect(page)
  await page.getByLabel('Run as').selectOption('statement')
  await page.getByRole('button', { name: /▶ Run/ }).click()
  await expect(page.getByText('2 rows affected', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Connections', exact: true }).click()
  await page.getByRole('button', { name: 'MySQL :13307' }).click()
  await page.getByRole('button', { name: 'Connect →', exact: true }).click()
  await expect(page.getByRole('tab', { name: /SQL 2/ })).toBeVisible()
  await page.getByLabel('Active connection').selectOption({ label: 'Local MariaDB' })
  await expect(page.getByRole('tab', { name: /SQL 1/ })).toBeVisible()
  await expect(page.getByText('2 rows affected', { exact: true })).toBeVisible()
})

test('closing a tab before operation registration cannot overwrite another result', async ({
  page,
}) => {
  await connect(page)
  await page.getByRole('button', { name: '▦ empty_table', exact: true }).click()
  await expect(page.getByText('No rows returned.')).toBeVisible()
  await page.getByRole('tab', { name: /SQL 1/ }).click()
  await page.getByRole('textbox', { name: 'SQL statement' }).fill('SELECT 1 /* DELAY_START */')
  await page.getByRole('button', { name: /▶ Run/ }).click()
  await page.getByRole('button', { name: 'Close SQL 1', exact: true }).click()
  await page.getByRole('button', { name: 'Continue', exact: true }).click()
  await expect(page.getByText('No rows returned.')).toBeVisible()
  await expect(page.getByRole('tab', { name: /SQL 1/ })).toHaveCount(0)
  await expect.poll(() => page.evaluate(() => (window as any).__released ?? 0)).toBeGreaterThan(0)
  await expect(page.getByText('No rows returned.')).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
})

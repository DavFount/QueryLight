import type { application } from '../../wailsjs/go/models'

export type Cell = application.Cell
export interface Draft {
  identity: string
  row: number
  col: number
  original: Cell
  value: Cell
}
export type Drafts = Record<string, Draft>
const textTypes = new Set(['char', 'varchar', 'tinytext', 'text', 'mediumtext', 'longtext'])
export const draftKey = (row: number, col: number) => `${row}:${col}`
export function rowIdentity(result: application.QueryResult, row: number): string | undefined {
  const keys = result.table?.primaryKey
  if (!keys?.length) return
  const cells = keys.map((key) => result.rows[row]?.[result.columns.findIndex((c) => c.name === key)])
  if (cells.some((c) => !c || c.unavailable || c.kind === 'null')) return
  return JSON.stringify(cells.map((c) => [c.kind, c.value]))
}
export function draftReason(result: application.QueryResult, row: number, col: number): string {
  if (!result.table?.editCandidate) return result.table?.readOnlyReason || 'SQL results are read-only.'
  const metadata = result.table.columns.find((c) => c.name === result.columns[col]?.name)
  const cell = result.rows[row]?.[col]
  if (!metadata || metadata.generated) return 'Generated or unknown columns are read-only.'
  if (result.table.primaryKey.includes(metadata.name)) return 'Primary-key editing is not supported yet.'
  if (!textTypes.has(metadata.dataType.toLowerCase())) return 'Local drafts currently support text columns only.'
  if (!cell || cell.unavailable || !['text', 'null'].includes(cell.kind)) return 'This value cannot be drafted losslessly.'
  if (!rowIdentity(result, row)) return 'The complete primary key is unavailable.'
  return ''
}
export function stageDraft(result: application.QueryResult, drafts: Drafts, row: number, col: number, text: string, isNull: boolean): Drafts {
  const reason = draftReason(result, row, col)
  if (reason) throw new Error(reason)
  const metadata = result.table!.columns.find((c) => c.name === result.columns[col].name)!
  if (isNull && !metadata.nullable) throw new Error('This column does not allow NULL.')
  const bytes = new TextEncoder().encode(isNull ? '' : text)
  if (!isNull && new TextDecoder().decode(bytes) !== text) throw new Error('The value contains invalid Unicode.')
  if (bytes.length > 256 * 1024) throw new Error('Draft values are limited to 256 KiB per cell.')
  const value: Cell = { kind: isNull ? 'null' : 'text', value: isNull ? '' : text, unavailable: false, byteLength: bytes.length }
  const original = result.rows[row][col]
  const next = { ...drafts }
  const key = draftKey(row, col)
  if (value.kind === original.kind && value.value === original.value) delete next[key]
  else next[key] = { identity: rowIdentity(result, row)!, row, col, original, value }
  if (Object.keys(next).length > 256 || new TextEncoder().encode(JSON.stringify(next)).length > 1024 * 1024)
    throw new Error('Local drafts are limited to 256 cells and 1 MiB per table tab. Discard some changes first.')
  return next
}
export const pendingRows = (drafts: Drafts) => new Set(Object.values(drafts).map((d) => d.identity)).size

import { expect, test } from '@playwright/test'
import { stageDraft, pendingRows, draftReason } from '../src/features/table-drafts'
import type { Result } from '../src/services/database'

function fixture(): Result {
  const cell = (kind: string, value: string) => ({ kind, value, byteLength: value.length, unavailable: false })
  return {
    columns: [{name: 'id'}, {name: 'label'}, {name: 'other'}],
    rows: Array.from({length: 257}, (_, i) => [cell('integer', String(i)), cell('text', 'original'), cell('null', '')]),
    table: { editCandidate: true, primaryKey: ['id'], columns: [
      {name: 'id', dataType: 'int', nullable: false, generated: false},
      {name: 'label', dataType: 'varchar', nullable: false, generated: false},
      {name: 'other', dataType: 'text', nullable: true, generated: false},
    ] },
  } as Result
}
test('drafts coalesce, preserve original data and distinguish NULL from empty', () => {
  const result = fixture()
  let drafts = stageDraft(result, {}, 0, 1, 'changed', false)
  drafts = stageDraft(result, drafts, 0, 1, 'changed twice', false)
  drafts = stageDraft(result, drafts, 0, 2, '', false)
  expect(Object.keys(drafts)).toHaveLength(2)
  expect(pendingRows(drafts)).toBe(1)
  expect(drafts['0:2'].value.kind).toBe('text')
  expect(result.rows[0][1].value).toBe('original')
  expect(drafts['0:1'].original.value).toBe('original')
  expect(drafts['0:1'].identity).toBe('[["integer","0"]]')
  drafts = stageDraft(result, drafts, 0, 1, 'original', false)
  drafts = stageDraft(result, drafts, 0, 2, '', true)
  expect(drafts).toEqual({})
})
test('drafts reject unsupported or unavailable values and invalid NULL', () => {
  const result = fixture()
  expect(() => stageDraft(result, {}, 0, 0, 'new key', false)).toThrow('Primary-key')
  expect(() => stageDraft(result, {}, 0, 1, '', true)).toThrow('does not allow NULL')
  result.table!.columns[1].generated = true
  expect(draftReason(result, 0, 1)).toContain('Generated')
  result.table!.columns[1].generated = false
  result.rows[0][1].unavailable = true
  expect(draftReason(result, 0, 1)).toContain('losslessly')
  result.rows[0][1].unavailable = false
  result.rows[0][0].unavailable = true
  expect(draftReason(result, 0, 1)).toContain('primary key')
})
test('draft bounds reject excess bytes and cells without mutating existing drafts', () => {
  const result = fixture()
  expect(() => stageDraft(result, {}, 0, 1, '雪'.repeat(100000), false)).toThrow('256 KiB')
  expect(() => stageDraft(result, {}, 0, 1, '\ud800', false)).toThrow('Unicode')
  let drafts = {}
  for (let i = 0; i < 256; i++) drafts = stageDraft(result, drafts, i, 1, 'small', false)
  expect(() => stageDraft(result, drafts, 256, 1, 'small', false)).toThrow('256 cells')
  expect(Object.keys(drafts)).toHaveLength(256)
  let large = stageDraft(result, {}, 0, 1, 'x'.repeat(262144), false)
  large = stageDraft(result, large, 1, 1, 'x'.repeat(262144), false)
  large = stageDraft(result, large, 2, 1, 'x'.repeat(262144), false)
  expect(() => stageDraft(result, large, 3, 1, 'x'.repeat(262144), false)).toThrow('1 MiB')
})

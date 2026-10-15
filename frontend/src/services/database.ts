import * as backend from '../../wailsjs/go/main/App'
import type { application } from '../../wailsjs/go/models'

export type Config = application.ConnectionConfig
export type Connection = application.ConnectionInfo
export type Failure = application.Failure
export type Result = application.QueryResult
export type Table = application.TableInfo

export class DatabaseError extends Error {
  constructor(public details: Failure) {
    super(details.message)
  }
}
export function checked<T extends { error?: Failure | null }>(reply: T): T {
  if (reply.error) throw new DatabaseError(reply.error)
  return reply
}
export function message(error: unknown): string {
  return error instanceof Error
    ? error.message
    : 'The desktop backend is unavailable. Run Querylight through Wails.'
}
export const db = backend

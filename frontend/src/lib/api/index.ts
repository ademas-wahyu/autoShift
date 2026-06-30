import type { ScheduleConfig, ScheduleDetail } from '@/types'

let _base: string | null = null

async function getBaseUrl(): Promise<string> {
  if (_base) return _base

  let base: string
  if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
    const { invoke } = await import('@tauri-apps/api/core')
    const port = await invoke<number>('get_api_port')
    base = `http://localhost:${port}/api/v1`
  } else {
    base = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
  }

  _base = base
  return base
}

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const base = await getBaseUrl()
  const res = await fetch(`${base}${path}`, {
    headers: { 'Content-Type': 'application/json', ...opts?.headers },
    ...opts,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }
  return res.json()
}

export const api = {
  createSchedule: (config: ScheduleConfig) =>
    request<{ data: { schedule_id: number; status: string } }>('/schedules', {
      method: 'POST',
      body: JSON.stringify(config),
    }),

  getSchedule: (id: number) =>
    request<{ data: ScheduleDetail }>(`/schedules/${id}`),

  validateSchedule: (id: number) =>
    request<{ data: { is_valid: boolean; violations: any[] } }>(`/schedules/${id}/validate`),

  updateShifts: (id: number, changes: any) =>
    request<{ data: any }>(`/schedules/${id}/shifts`, {
      method: 'PUT',
      body: JSON.stringify(changes),
    }),

  publishSchedule: (id: number, notes?: string) =>
    request<{ data: any }>(`/schedules/${id}/publish`, {
      method: 'PUT',
      body: notes ? JSON.stringify({ notes }) : undefined,
    }),

  regenerate: (id: number, body: any) =>
    request<{ data: any }>(`/schedules/${id}/regenerate`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
}

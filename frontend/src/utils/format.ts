// 共享格式化工具
export function formatDate(s?: string | null): string {
  if (!s) return '-'
  return s.slice(0, 10)
}

export function formatDateTime(s?: string | null): string {
  if (!s) return '-'
  return s
}

export function formatWeight(w?: number | null): string {
  if (w === undefined || w === null) return '-'
  return `${w.toFixed(2)} kg`
}

export function formatArea(a?: number | null): string {
  if (a === undefined || a === null) return '-'
  return `${a.toFixed(1)} m²`
}

export function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

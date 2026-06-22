import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Initials for an actor like "agent:claude-1" -> "CL", "human:shah" -> "SH". */
export function initials(actor?: string): string {
  if (!actor) return "?"
  const name = actor.includes(":") ? actor.slice(actor.indexOf(":") + 1) : actor
  const letters = name.replace(/[^a-zA-Z]/g, "")
  return (letters.slice(0, 2) || name.slice(0, 2)).toUpperCase()
}

/** Whether an actor is an AI agent, a human, or unknown — by the "agent:"/"human:" prefix. */
export function actorKind(actor?: string): "agent" | "human" | null {
  if (!actor) return null
  if (actor.startsWith("agent:")) return "agent"
  if (actor.startsWith("human:")) return "human"
  return null
}

/** Human label for a status string: "in_progress" -> "In progress". */
export function statusLabel(status: string): string {
  const s = status.replace(/_/g, " ")
  return s.charAt(0).toUpperCase() + s.slice(1)
}

/** Compact relative time: "just now", "5m ago", "3h ago", "2d ago", "4w ago". */
export function timeAgo(iso?: string): string {
  if (!iso) return ""
  const t = new Date(iso).getTime()
  if (Number.isNaN(t)) return ""
  const s = Math.floor((Date.now() - t) / 1000)
  if (s < 45) return "just now"
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  if (d < 7) return `${d}d ago`
  const w = Math.floor(d / 7)
  if (w < 5) return `${w}w ago`
  const mo = Math.floor(d / 30)
  if (mo < 12) return `${mo}mo ago`
  return `${Math.floor(d / 365)}y ago`
}

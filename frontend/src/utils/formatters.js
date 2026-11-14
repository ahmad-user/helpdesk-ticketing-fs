/**
 * Format ISO date string into readable format
 * e.g., "Nov 14, 2026, 08:30 AM"
 */
export function formatDate(isoString) {
  if (!isoString) return '-'
  try {
    const date = new Date(isoString)
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date)
  } catch {
    return isoString
  }
}

/**
 * Format relative time
 */
export function timeAgo(isoString) {
  if (!isoString) return '-'
  try {
    const date = new Date(isoString)
    const now = new Date()
    const diffSec = Math.floor((now - date) / 1000)

    if (diffSec < 60) return 'just now'
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`
    return `${Math.floor(diffSec / 86400)}d ago`
  } catch {
    return isoString
  }
}

/**
 * Human readable status labels
 */
export const STATUS_LABELS = {
  open: 'Open',
  in_progress: 'In Progress',
  resolved: 'Resolved',
  closed: 'Closed'
}

/**
 * Human readable priority labels
 */
export const PRIORITY_LABELS = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent'
}

/**
 * Human readable role labels
 */
export const ROLE_LABELS = {
  requester: 'Requester',
  agent: 'Agent',
  admin: 'Administrator'
}

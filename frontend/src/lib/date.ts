const TZ = 'Asia/Bangkok'

/** Returns "YYYY-MM-DD" in Asia/Bangkok time. */
export function todayDate(): string {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: TZ,
    dateStyle: 'short',
  }).format(new Date())
}

/** Returns "YYYY-MM" in Asia/Bangkok time. */
export function todayMonth(): string {
  return todayDate().slice(0, 7)
}

/**
 * Add `days` (may be negative) to a "YYYY-MM-DD" date, returning "YYYY-MM-DD".
 * Anchors at noon to avoid DST/timezone drift around midnight.
 */
export function addDays(date: string, days: number): string {
  const d = new Date(date + 'T12:00:00')
  d.setDate(d.getDate() + days)
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: TZ,
    dateStyle: 'short',
  }).format(d)
}

/** Day-of-week (0=Sun .. 6=Sat) for a "YYYY-MM-DD" date. */
export function dayOfWeek(date: string): number {
  return new Date(date + 'T12:00:00').getDay()
}

/**
 * Formats an ISO timestamp (or date string) into a human-readable date.
 * Default: { day:'numeric', month:'short', year:'numeric' } e.g. "1 Jun 2025".
 * Pass custom opts to override (e.g. add weekday, omit year, etc.).
 */
export function formatDate(
  iso: string,
  opts: Intl.DateTimeFormatOptions = {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  },
): string {
  return new Date(iso).toLocaleDateString(undefined, opts)
}

/**
 * Formats an ISO timestamp into a human-readable time.
 * Default: { hour:'2-digit', minute:'2-digit' } e.g. "02:30 PM".
 */
export function formatTime(
  iso: string,
  opts: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' },
): string {
  return new Date(iso).toLocaleTimeString(undefined, opts)
}

/** Present age in whole years from a "YYYY-MM-DD" birthdate (Asia/Bangkok). */
export function calculateAge(birthdate: string): number {
  const [by, bm, bd] = birthdate.split('-').map(Number)
  const [ty, tm, td] = todayDate().split('-').map(Number)
  let age = ty - by
  if (tm < bm || (tm === bm && td < bd)) age--
  return age
}

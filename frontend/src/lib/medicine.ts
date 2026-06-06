import type {
  AdjustmentType,
  FrequencyType,
  IntakeStatus,
  Medicine,
  MedicineSourceType,
  MedicineTiming,
} from './api/types'

// ── Source ────────────────────────────────────────────────────────────────────

export const SOURCE_TYPES: MedicineSourceType[] = ['medication', 'supplement']

export const SOURCE_LABELS: Record<MedicineSourceType, string> = {
  medication: 'Medication',
  supplement: 'Supplement',
}

export const SOURCE_BADGE: Record<MedicineSourceType, string> = {
  medication: 'bg-indigo-900/50 text-indigo-300',
  supplement: 'bg-emerald-900/50 text-emerald-300',
}

// ── Frequency ─────────────────────────────────────────────────────────────────

export const FREQUENCY_TYPES: FrequencyType[] = [
  'daily',
  'weekly',
  'as_needed',
  'custom',
]

export const FREQUENCY_LABELS: Record<FrequencyType, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  as_needed: 'As needed',
  custom: 'Custom',
}

// ── Timing ────────────────────────────────────────────────────────────────────

export const TIMING_LABELS: Record<MedicineTiming, string> = {
  anytime: 'Anytime',
  before_meal: 'Before meal',
  after_meal: 'After meal',
  before_breakfast: 'Before breakfast',
  after_breakfast: 'After breakfast',
  before_lunch: 'Before lunch',
  after_lunch: 'After lunch',
  before_dinner: 'Before dinner',
  after_dinner: 'After dinner',
  before_bed: 'Before bed',
}

/** Ordered timing options for select inputs, including a blank "Not specified" entry. */
export const TIMINGS: Array<{ value: MedicineTiming | ''; label: string }> = [
  { value: '', label: 'Not specified' },
  { value: 'anytime', label: 'Anytime' },
  { value: 'before_meal', label: 'Before meal' },
  { value: 'after_meal', label: 'After meal' },
  { value: 'before_breakfast', label: 'Before breakfast' },
  { value: 'after_breakfast', label: 'After breakfast' },
  { value: 'before_lunch', label: 'Before lunch' },
  { value: 'after_lunch', label: 'After lunch' },
  { value: 'before_dinner', label: 'Before dinner' },
  { value: 'after_dinner', label: 'After dinner' },
  { value: 'before_bed', label: 'Before bed' },
]

// ── Intake status ─────────────────────────────────────────────────────────────

export const STATUS_LABELS: Record<IntakeStatus, string> = {
  taken: 'Taken',
  skipped: 'Skipped',
  missed: 'Missed',
}

export const STATUS_COLORS: Record<IntakeStatus, string> = {
  taken: 'text-emerald-400 bg-emerald-900/30',
  skipped: 'text-yellow-400 bg-yellow-900/30',
  missed: 'text-red-400 bg-red-900/30',
}

// ── Stock adjustment ──────────────────────────────────────────────────────────

export const ADJ_LABELS: Record<AdjustmentType, string> = {
  add: '+ Add',
  remove: '- Remove',
  correction: '= Set',
}

export const ADJ_COLORS: Record<AdjustmentType, string> = {
  add: 'text-emerald-400 bg-emerald-900/30',
  remove: 'text-red-400 bg-red-900/30',
  correction: 'text-blue-400 bg-blue-900/30',
}

/**
 * Estimate the number of doses per day based on frequency settings.
 * Returns null when we can't determine a daily dose (as_needed, custom).
 */
export function dailyDoseEstimate(m: Medicine): number | null {
  const amount = m.dosage_amount
  if (amount <= 0) return null

  switch (m.frequency_type) {
    case 'daily':
      return amount * (m.frequency_value ?? 1)
    case 'weekly':
      return (amount * (m.frequency_value ?? 1)) / 7
    default:
      return null
  }
}

/**
 * Estimate how many days of stock remain.
 * Returns null when the daily dose cannot be calculated.
 */
export function estimatedDaysRemaining(m: Medicine): number | null {
  const dose = dailyDoseEstimate(m)
  if (dose === null || dose <= 0) return null
  return Math.floor(m.stock_quantity / dose)
}

/**
 * Returns true when the medicine is at or below its low-stock threshold.
 */
export function isLowStock(m: Medicine): boolean {
  return m.stock_quantity <= m.low_stock_threshold
}

/**
 * Returns the number of doses required per day, or null when daily gating
 * does not apply (weekly, as_needed, custom).
 */
export function requiredDailyDoses(m: Medicine): number | null {
  if (m.frequency_type !== 'daily') return null
  return m.frequency_value ?? 1
}

/**
 * Returns true when the ISO timestamp falls on the local calendar "today".
 * Uses local-timezone Date getters so midnight boundaries respect the user's TZ.
 */
export function isLocalToday(iso: string): boolean {
  const d = new Date(iso)
  const now = new Date()
  return (
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  )
}

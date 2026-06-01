import type { WorkoutType } from './api/types'

// Labels for all 7 workout types — use this instead of a closed literal record so
// TypeScript catches any missing keys when WorkoutType is widened.
export const WORKOUT_TYPE_LABELS: Record<WorkoutType, string> = {
  weight_training: 'Weight Training',
  body_weight: 'Body Weight',
  running: 'Running',
  walking: 'Walking',
  cardio: 'Cardio',
  mobility: 'Mobility',
  custom: 'Custom Workout',
}

// Types available when creating/editing a preset template (excludes 'custom').
export const PRESET_TYPES: WorkoutType[] = [
  'weight_training',
  'body_weight',
  'running',
  'walking',
  'cardio',
  'mobility',
]

// Quick-start buttons shown on the Today tab (excludes 'walking').
export const QUICK_START_TYPES: { type: WorkoutType; label: string }[] = [
  { type: 'weight_training', label: 'Weight Training' },
  { type: 'body_weight', label: 'Body Weight' },
  { type: 'running', label: 'Running' },
  { type: 'cardio', label: 'Cardio' },
  { type: 'mobility', label: 'Mobility' },
  { type: 'custom', label: 'Custom Workout' },
]

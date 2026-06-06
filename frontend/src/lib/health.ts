import type { Gender, HealthGoal, MealType } from './api/types'

export const GENDERS: Gender[] = ['male', 'female', 'other']

export const GENDER_LABELS: Record<Gender, string> = {
  male: 'Male',
  female: 'Female',
  other: 'Other',
}

export const HEALTH_GOALS: HealthGoal[] = [
  'lose_weight',
  'maintain',
  'gain_muscle',
]

export const GOAL_LABELS: Record<HealthGoal, string> = {
  lose_weight: 'Lose Weight',
  maintain: 'Maintain',
  gain_muscle: 'Gain Muscle',
}

export const MEAL_TYPES: MealType[] = ['breakfast', 'lunch', 'dinner', 'snack']

export const MEAL_LABELS: Record<MealType, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snack',
}

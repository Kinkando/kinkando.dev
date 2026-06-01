export function computeBmi(weightKg: number, heightCm: number): number {
  const heightM = heightCm / 100
  return weightKg / (heightM * heightM)
}

export type BmiCategory =
  | 'Underweight'
  | 'Normal weight'
  | 'Overweight'
  | 'Obese'

export function bmiCategory(bmi: number): BmiCategory {
  if (bmi < 18.5) return 'Underweight'
  if (bmi < 25) return 'Normal weight'
  if (bmi < 30) return 'Overweight'
  return 'Obese'
}

export function bmiColor(category: BmiCategory): string {
  switch (category) {
    case 'Underweight':
      return 'text-blue-400'
    case 'Normal weight':
      return 'text-emerald-400'
    case 'Overweight':
      return 'text-yellow-400'
    case 'Obese':
      return 'text-red-400'
  }
}

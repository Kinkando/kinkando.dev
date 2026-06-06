import { useEffect, useState } from 'react'
import type { HealthProfile, Gender, HealthGoal } from '../../lib/api/types'
import { useUpsertProfile, useCreateWeightLog } from '../../queries/useHealth'
import { todayDate } from '../../lib/date'
import {
  GENDERS,
  GENDER_LABELS,
  HEALTH_GOALS,
  GOAL_LABELS,
} from '../../lib/health'

type Props = {
  profile: HealthProfile | null | undefined
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

export default function SettingsTab({ profile }: Props) {
  const [height, setHeight] = useState('')
  const [birthdate, setBirthdate] = useState('')
  const [gender, setGender] = useState<Gender | ''>('')
  const [goal, setGoal] = useState<HealthGoal | ''>('')
  const [profileError, setProfileError] = useState('')
  const [profileSuccess, setProfileSuccess] = useState(false)

  const [currentWeight, setCurrentWeight] = useState('')
  const [weightError, setWeightError] = useState('')
  const [weightSuccess, setWeightSuccess] = useState(false)

  const upsertProfile = useUpsertProfile()
  const createWeight = useCreateWeightLog()

  // Pre-fill form when profile loads
  useEffect(() => {
    if (profile) {
      setHeight(profile.height != null ? String(profile.height) : '')
      setBirthdate(profile.birthdate ?? '')
      setGender(profile.gender ?? '')
      setGoal(profile.goal ?? '')
    }
  }, [profile])

  async function handleProfileSubmit(e: React.FormEvent) {
    e.preventDefault()
    setProfileError('')
    setProfileSuccess(false)

    const heightNum = height ? parseFloat(height) : null

    if (heightNum != null && heightNum <= 0) {
      setProfileError('Height must be positive.')
      return
    }
    if (birthdate && birthdate > todayDate()) {
      setProfileError("Birthdate can't be in the future.")
      return
    }

    try {
      await upsertProfile.mutateAsync({
        height: heightNum,
        birthdate: birthdate || null,
        gender: gender || null,
        goal: goal || null,
      })
      setProfileSuccess(true)
      setTimeout(() => setProfileSuccess(false), 2500)
    } catch (err) {
      setProfileError(
        err instanceof Error ? err.message : 'Something went wrong.',
      )
    }
  }

  async function handleWeightSubmit(e: React.FormEvent) {
    e.preventDefault()
    setWeightError('')
    setWeightSuccess(false)

    const w = parseFloat(currentWeight)
    if (!currentWeight || isNaN(w) || w <= 0) {
      setWeightError('Enter a valid weight.')
      return
    }

    try {
      await createWeight.mutateAsync({ weight: w })
      setCurrentWeight('')
      setWeightSuccess(true)
      setTimeout(() => setWeightSuccess(false), 2500)
    } catch (err) {
      setWeightError(
        err instanceof Error ? err.message : 'Something went wrong.',
      )
    }
  }

  return (
    <div className="space-y-6">
      {/* Profile form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">Profile</h3>
        <form onSubmit={handleProfileSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className={labelClass}>Height (cm)</label>
              <input
                className={inputClass}
                type="number"
                step="0.5"
                min="1"
                placeholder="e.g. 175"
                value={height}
                onChange={(e) => setHeight(e.target.value)}
              />
            </div>
            <div>
              <label className={labelClass}>Birthdate</label>
              <input
                className={inputClass}
                type="date"
                max={todayDate()}
                value={birthdate}
                onChange={(e) => setBirthdate(e.target.value)}
              />
            </div>
            <div>
              <label className={labelClass}>Gender</label>
              <select
                className={inputClass}
                value={gender}
                onChange={(e) => setGender(e.target.value as Gender | '')}
              >
                <option value="">— Select —</option>
                {GENDERS.map((g) => (
                  <option key={g} value={g}>
                    {GENDER_LABELS[g]}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className={labelClass}>Goal</label>
              <select
                className={inputClass}
                value={goal}
                onChange={(e) => setGoal(e.target.value as HealthGoal | '')}
              >
                <option value="">— Select —</option>
                {HEALTH_GOALS.map((g) => (
                  <option key={g} value={g}>
                    {GOAL_LABELS[g]}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {profileError && (
            <p className="text-sm text-red-400">{profileError}</p>
          )}
          {profileSuccess && (
            <p className="text-sm text-emerald-400">Profile saved.</p>
          )}

          <button
            type="submit"
            disabled={upsertProfile.isPending}
            className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
          >
            {upsertProfile.isPending ? 'Saving…' : 'Save Profile'}
          </button>
        </form>
      </div>
    </div>
  )
}

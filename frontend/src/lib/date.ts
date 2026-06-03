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

/** Present age in whole years from a "YYYY-MM-DD" birthdate (Asia/Bangkok). */
export function calculateAge(birthdate: string): number {
  const [by, bm, bd] = birthdate.split('-').map(Number)
  const [ty, tm, td] = todayDate().split('-').map(Number)
  let age = ty - by
  if (tm < bm || (tm === bm && td < bd)) age--
  return age
}

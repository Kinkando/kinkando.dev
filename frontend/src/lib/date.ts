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

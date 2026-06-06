/**
 * Merges class strings, filtering out falsy values.
 * Prefer this over template-literal className concatenation.
 */
export function cn(
  ...classes: Array<string | false | null | undefined>
): string {
  return classes.filter(Boolean).join(' ')
}

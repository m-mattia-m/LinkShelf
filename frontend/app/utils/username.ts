import * as v from 'valibot'

export const USERNAME_MIN_LENGTH = 3
export const USERNAME_MAX_LENGTH = 30
export const USERNAME_PATTERN = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/

/**
 * The shape of a username, mirroring the backend (see
 * backend/internal/domain/username.go). Reserved words are only checked
 * there - the response explains which one was rejected.
 */
export function usernameSchema() {
  return v.pipe(
    v.string('Username is required'),
    v.nonEmpty('Username is required'),
    v.minLength(USERNAME_MIN_LENGTH, `Must be at least ${USERNAME_MIN_LENGTH} characters`),
    v.maxLength(USERNAME_MAX_LENGTH, `Must be at most ${USERNAME_MAX_LENGTH} characters`),
    v.regex(USERNAME_PATTERN, 'Lowercase letters, numbers and hyphens only, and no hyphen at the start or end')
  )
}

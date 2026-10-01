import * as v from 'valibot'

export const USERNAME_MIN_LENGTH = 3
export const USERNAME_MAX_LENGTH = 30
export const USERNAME_PATTERN = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/

/** Mirrors usernamePattern in backend/internal/domain/username.go. */
export function usernameSchema() {
  const { t } = useNuxtApp().$i18n
  return v.pipe(
    v.string(t('validation.usernameRequired')),
    v.nonEmpty(t('validation.usernameRequired')),
    v.minLength(USERNAME_MIN_LENGTH, t('validation.minLength', { min: USERNAME_MIN_LENGTH })),
    v.maxLength(USERNAME_MAX_LENGTH, t('validation.maxLength', { max: USERNAME_MAX_LENGTH })),
    v.regex(USERNAME_PATTERN, t('validation.usernamePattern'))
  )
}

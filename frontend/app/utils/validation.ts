import * as v from 'valibot'

// Mirrors the minLength on the backend's password fields.
export const PASSWORD_MIN_LENGTH = 8

/** A non-empty string, with the same message for a missing and an empty value. */
export function requiredStringSchema(messageKey: string) {
  const { t } = useNuxtApp().$i18n
  return v.pipe(v.string(t(messageKey)), v.nonEmpty(t(messageKey)))
}

export function emailSchema() {
  const { t } = useNuxtApp().$i18n
  return v.pipe(
    v.string(t('validation.emailRequired')),
    v.nonEmpty(t('validation.emailRequired')),
    v.email(t('validation.email'))
  )
}

/** A password being chosen, as opposed to one being entered to sign in. */
export function newPasswordSchema() {
  const { t } = useNuxtApp().$i18n
  return v.pipe(
    v.string(t('validation.passwordRequired')),
    v.nonEmpty(t('validation.passwordRequired')),
    v.minLength(PASSWORD_MIN_LENGTH, t('validation.minLength', { min: PASSWORD_MIN_LENGTH }))
  )
}

/** A new password plus its confirmation; a mismatch is reported on confirmPassword. */
export function passwordWithConfirmationSchema() {
  const { t } = useNuxtApp().$i18n
  return v.pipe(
    v.object({
      password: newPasswordSchema(),
      confirmPassword: requiredStringSchema('validation.confirmPassword')
    }),
    v.forward(
      v.partialCheck(
        [['password'], ['confirmPassword']],
        input => input.password === input.confirmPassword,
        t('validation.passwordsMismatch')
      ),
      ['confirmPassword']
    )
  )
}

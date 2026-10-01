import { ResponseError, type ErrorModel } from '~~/api'
import type { FormError } from '@nuxt/ui'

export interface ApiErrorResult {
  message: string
  fieldErrors: FormError[]
}

// Resolved before the first await, while the Nuxt app context is still set.
function translator(): (key: string) => string {
  const i18n = tryUseNuxtApp()?.$i18n
  return key => (i18n ? i18n.t(key) : key)
}

export async function parseApiError(err: unknown): Promise<ApiErrorResult> {
  const t = translator()
  if (err instanceof ResponseError) {
    try {
      const body = await err.response.clone().json() as ErrorModel
      const fieldErrors: FormError[] = (body.errors ?? [])
        .filter(detail => detail.location && detail.message)
        .map(detail => ({
          name: detail.location!.replace(/^body\./, ''),
          message: detail.message!
        }))

      // Prefer the specific messages over huma's generic detail.
      const specificMessages = (body.errors ?? [])
        .filter(detail => detail.message)
        .map(detail => detail.location
          ? `${detail.location.replace(/^body\./, '')}: ${detail.message}`
          : detail.message!)

      const message = specificMessages.length > 0
        ? specificMessages.join('; ')
        : (body.detail || t('common.somethingWentWrong'))

      return { message, fieldErrors }
    } catch {
      return { message: err.message, fieldErrors: [] }
    }
  }

  if (err instanceof Error) {
    return { message: err.message, fieldErrors: [] }
  }

  return { message: t('common.somethingWentWrong'), fieldErrors: [] }
}

export async function handleApiError(err: unknown, formRef?: { setErrors: (errs: FormError[]) => void } | null): Promise<ApiErrorResult> {
  const toast = useToast()
  const t = translator()
  const result = await parseApiError(err)

  toast.add({
    title: t('common.error'),
    description: result.message,
    color: 'error'
  })

  if (formRef && result.fieldErrors.length > 0) {
    formRef.setErrors(result.fieldErrors)
  }

  return result
}

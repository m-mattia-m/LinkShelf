import { ResponseError, type ErrorModel } from '~~/api'
import type { FormError } from '@nuxt/ui'

export interface ApiErrorResult {
  message: string
  fieldErrors: FormError[]
}

export async function parseApiError(err: unknown): Promise<ApiErrorResult> {
  if (err instanceof ResponseError) {
    try {
      const body = await err.response.clone().json() as ErrorModel
      const fieldErrors: FormError[] = (body.errors ?? [])
        .filter((detail) => detail.location && detail.message)
        .map((detail) => ({
          name: detail.location!.replace(/^body\./, ''),
          message: detail.message!
        }))

      return { message: body.detail || 'Something went wrong.', fieldErrors }
    } catch {
      return { message: err.message, fieldErrors: [] }
    }
  }

  if (err instanceof Error) {
    return { message: err.message, fieldErrors: [] }
  }

  return { message: 'Something went wrong.', fieldErrors: [] }
}

export async function handleApiError(err: unknown, formRef?: { setErrors: (errs: FormError[]) => void } | null): Promise<ApiErrorResult> {
  const toast = useToast()
  const result = await parseApiError(err)

  toast.add({
    title: 'Error',
    description: result.message,
    color: 'error'
  })

  if (formRef && result.fieldErrors.length > 0) {
    formRef.setErrors(result.fieldErrors)
  }

  return result
}

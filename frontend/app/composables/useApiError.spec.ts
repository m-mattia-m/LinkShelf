import { describe, expect, it } from 'vitest'
import { ResponseError } from '~~/api'
import { handleApiError, parseApiError } from './useApiError'

function responseError(status: number, body: unknown): ResponseError {
  return new ResponseError(new Response(JSON.stringify(body), { status }))
}

describe('parseApiError', () => {
  it('maps huma field errors onto form errors, stripping the "body." prefix', async () => {
    const err = responseError(422, {
      detail: 'validation failed',
      errors: [
        { location: 'body.email', message: 'must be a valid email' },
        { location: 'body.title', message: 'is required' }
      ]
    })

    const result = await parseApiError(err)

    expect(result.fieldErrors).toEqual([
      { name: 'email', message: 'must be a valid email' },
      { name: 'title', message: 'is required' }
    ])
  })

  it('joins per-field messages into the summary, preferring them over the generic detail', async () => {
    const err = responseError(422, {
      detail: 'validation failed',
      errors: [
        { location: 'body.email', message: 'must be a valid email' },
        { message: 'unexpected property' }
      ]
    })

    const result = await parseApiError(err)

    expect(result.message).toBe('email: must be a valid email; unexpected property')
  })

  it('falls back to the top-level detail when no error item carries a message', async () => {
    const err = responseError(500, { detail: 'internal server error', errors: [] })

    const result = await parseApiError(err)

    expect(result.message).toBe('internal server error')
    expect(result.fieldErrors).toEqual([])
  })

  it('falls back to a generic message when detail is also missing', async () => {
    const err = responseError(500, {})

    const result = await parseApiError(err)

    expect(result.message).toBe('Something went wrong.')
  })

  it('falls back to the raw error message when the response body is not JSON', async () => {
    const response = new Response('not json', { status: 500 })
    const err = new ResponseError(response)

    const result = await parseApiError(err)

    expect(result.message).toBe(err.message)
    expect(result.fieldErrors).toEqual([])
  })

  it('unwraps a plain Error', async () => {
    const result = await parseApiError(new Error('boom'))
    expect(result.message).toBe('boom')
  })

  it('falls back to a generic message for a non-Error value', async () => {
    const result = await parseApiError('nope')
    expect(result.message).toBe('Something went wrong.')
  })
})

describe('handleApiError', () => {
  it('pushes a toast and returns the parsed result', async () => {
    const err = responseError(500, { detail: 'internal server error' })

    const result = await handleApiError(err)

    expect(result.message).toBe('internal server error')
  })

  it('forwards field errors to the given form ref', async () => {
    const err = responseError(422, {
      errors: [{ location: 'body.email', message: 'must be a valid email' }]
    })
    const setErrors = vi.fn()

    await handleApiError(err, { setErrors })

    expect(setErrors).toHaveBeenCalledWith([{ name: 'email', message: 'must be a valid email' }])
  })

  it('does not call setErrors when there are no field errors', async () => {
    const err = responseError(500, { detail: 'internal server error' })
    const setErrors = vi.fn()

    await handleApiError(err, { setErrors })

    expect(setErrors).not.toHaveBeenCalled()
  })
})

import * as v from 'valibot'
import { describe, expect, it } from 'vitest'
import { emailSchema, newPasswordSchema, passwordWithConfirmationSchema, PASSWORD_MIN_LENGTH, requiredStringSchema } from './validation'

function messages(schema: v.GenericSchema, value: unknown): string[] {
  const result = v.safeParse(schema, value)
  return result.success ? [] : result.issues.map(issue => issue.message)
}

describe('requiredStringSchema', () => {
  it('accepts a non-empty string', () => {
    expect(messages(requiredStringSchema('validation.firstNameRequired'), 'Ada')).toEqual([])
  })

  it.each(['', undefined])('rejects %j with the given message', (value) => {
    expect(messages(requiredStringSchema('validation.firstNameRequired'), value)).toEqual(['First name is required'])
  })
})

describe('emailSchema', () => {
  it('accepts an email address', () => {
    expect(messages(emailSchema(), 'ada@example.com')).toEqual([])
  })

  it.each([
    ['', 'Email is required'],
    [undefined, 'Email is required'],
    ['not-an-email', 'Please enter a valid email']
  ])('rejects %j', (value, expected) => {
    expect(messages(emailSchema(), value)).toContain(expected)
  })
})

describe('newPasswordSchema', () => {
  it('accepts a password of the minimum length', () => {
    expect(messages(newPasswordSchema(), 'a'.repeat(PASSWORD_MIN_LENGTH))).toEqual([])
  })

  it.each([
    ['', 'Password is required'],
    [undefined, 'Password is required'],
    ['a'.repeat(PASSWORD_MIN_LENGTH - 1), 'Must be at least 8 characters']
  ])('rejects %j', (value, expected) => {
    expect(messages(newPasswordSchema(), value)).toContain(expected)
  })
})

describe('passwordWithConfirmationSchema', () => {
  it('accepts a matching password and confirmation', () => {
    expect(messages(passwordWithConfirmationSchema(), { password: 'secret-123', confirmPassword: 'secret-123' })).toEqual([])
  })

  it('reports a mismatch on confirmPassword', () => {
    const result = v.safeParse(passwordWithConfirmationSchema(), { password: 'secret-123', confirmPassword: 'secret-456' })

    expect(result.success).toBe(false)
    expect(result.issues?.map(issue => [v.getDotPath(issue), issue.message])).toEqual([['confirmPassword', 'Passwords do not match']])
  })

  it('requires the confirmation', () => {
    expect(messages(passwordWithConfirmationSchema(), { password: 'secret-123', confirmPassword: '' })).toContain('Please confirm your password')
  })
})

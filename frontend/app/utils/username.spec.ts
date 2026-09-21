import * as v from 'valibot'
import { describe, expect, it } from 'vitest'
import { USERNAME_PATTERN, usernameSchema } from './username'

function messages(value: unknown): string[] {
  const result = v.safeParse(usernameSchema(), value)
  return result.success ? [] : result.issues.map(issue => issue.message)
}

describe('usernameSchema', () => {
  it.each(['abc', 'a-b', 'user-1', 'a1b', 'john-smith', 'a'.repeat(30)])('accepts %s', (name) => {
    expect(messages(name)).toEqual([])
  })

  it.each([
    ['', 'Username is required'],
    ['ab', 'Must be at least 3 characters'],
    ['a'.repeat(31), 'Must be at most 30 characters'],
    ['Abc', 'Lowercase letters, numbers and hyphens only'],
    ['-abc', 'Lowercase letters, numbers and hyphens only'],
    ['abc-', 'Lowercase letters, numbers and hyphens only'],
    ['a_b', 'Lowercase letters, numbers and hyphens only'],
    ['a b', 'Lowercase letters, numbers and hyphens only'],
    ['a.b', 'Lowercase letters, numbers and hyphens only']
  ])('rejects %j', (name, expected) => {
    expect(messages(name).join(' | ')).toContain(expected)
  })

  it('rejects a value that is not a string', () => {
    expect(messages(undefined)).toContain('Username is required')
  })

  it('exports the same pattern the backend documents', () => {
    expect(USERNAME_PATTERN.source).toBe('^[a-z0-9]([a-z0-9-]*[a-z0-9])?$')
  })
})

import { describe, expect, it } from 'vitest'
import { safeHref } from './safeHref'

describe('safeHref', () => {
  it.each([
    'https://example.com',
    'http://example.com/path?x=1#y',
    'HTTPS://Example.com'
  ])('keeps the http(s) URL %s unchanged', (input) => {
    expect(safeHref(input)).toBe(input)
  })

  // Regression: script-running schemes never become an href.
  it.each([
    'javascript:alert(document.domain)%2F%2F@example.com',
    'JavaScript:alert(1)',
    '  javascript:alert(1)',
    'java\tscript:alert(1)',
    'data:text/html,<script>alert(1)</script>',
    'vbscript:msgbox(1)',
    'mailto:someone@example.com',
    'ftp://example.com',
    'example.com',
    '/relative/path',
    '',
    null,
    undefined
  ])('refuses %s', (input) => {
    expect(safeHref(input)).toBeUndefined()
  })
})

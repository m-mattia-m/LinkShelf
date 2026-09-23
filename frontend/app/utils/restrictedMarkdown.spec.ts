import { describe, expect, it } from 'vitest'
import { renderRestrictedMarkdown } from './restrictedMarkdown'

describe('renderRestrictedMarkdown', () => {
  it('renders bold text', () => {
    expect(renderRestrictedMarkdown('**hello**')).toBe('<strong>hello</strong>')
  })

  it('renders italic text', () => {
    expect(renderRestrictedMarkdown('*hello*')).toBe('<em>hello</em>')
  })

  it('renders a link with a safe href, opening in a new tab', () => {
    expect(renderRestrictedMarkdown('[Site](https://example.com)'))
      .toBe('<a href="https://example.com" target="_blank" rel="noopener noreferrer">Site</a>')
  })

  it('renders a mailto link', () => {
    expect(renderRestrictedMarkdown('[Mail](mailto:hi@example.com)'))
      .toBe('<a href="mailto:hi@example.com" target="_blank" rel="noopener noreferrer">Mail</a>')
  })

  it('combines bold, italic and a link in one string', () => {
    expect(renderRestrictedMarkdown('**Bold** and *italic* and [a link](https://example.com)'))
      .toBe('<strong>Bold</strong> and <em>italic</em> and <a href="https://example.com" target="_blank" rel="noopener noreferrer">a link</a>')
  })

  it('turns a newline into a line break', () => {
    expect(renderRestrictedMarkdown('line one\nline two')).toBe('line one<br>line two')
  })

  it('leaves a non-http(s)/mailto link target as plain escaped text, not a link', () => {
    expect(renderRestrictedMarkdown('[click me](javascript:alert(1))'))
      .toBe('[click me](javascript:alert(1))')
  })

  it('escapes a literal script tag instead of executing it', () => {
    expect(renderRestrictedMarkdown('<script>alert(1)</script>'))
      .toBe('&lt;script&gt;alert(1)&lt;/script&gt;')
  })

  it('escapes an onerror image attribute attempt to plain text', () => {
    expect(renderRestrictedMarkdown('<img src=x onerror=alert(1)>'))
      .toBe('&lt;img src=x onerror=alert(1)&gt;')
  })

  it('does not let a link label or url break out of the anchor tag', () => {
    const result = renderRestrictedMarkdown('[") x=y ("](https://example.com/")x=y(")')
    expect(result).not.toContain('<a href="https://example.com/"x=y')
    expect(result).toContain('&quot;')
  })

  it('renders plain text unchanged aside from HTML-escaping', () => {
    expect(renderRestrictedMarkdown('Just plain text.')).toBe('Just plain text.')
  })

  it('returns an empty string for empty input', () => {
    expect(renderRestrictedMarkdown('')).toBe('')
  })
})

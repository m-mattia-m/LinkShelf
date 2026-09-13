import { renderSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import AppLogo from './AppLogo.vue'

const TwoLogos = defineComponent({
  render() {
    return h('div', [h(AppLogo), h(AppLogo)])
  }
})

describe('AppLogo', () => {
  it('renders an svg with the expected viewBox', async () => {
    const { container } = await renderSuspended(AppLogo)

    const svg = container.querySelector('svg')
    expect(svg).toBeInTheDocument()
    expect(svg).toHaveAttribute('viewBox', '0 0 512 512')
  })

  it('generates a unique clip-path id per instance so multiple logos on one page do not collide', async () => {
    const { container } = await renderSuspended(TwoLogos)

    const clipPaths = container.querySelectorAll('clipPath')
    expect(clipPaths.length).toBe(2)

    const ids = Array.from(clipPaths).map(el => el.id)
    expect(ids[0]).toBeTruthy()
    expect(ids[1]).toBeTruthy()
    expect(ids[0]).not.toBe(ids[1])

    const clippedGroups = container.querySelectorAll('g[clip-path]')
    expect(clippedGroups[0]?.getAttribute('clip-path')).toBe(`url(#${ids[0]})`)
    expect(clippedGroups[1]?.getAttribute('clip-path')).toBe(`url(#${ids[1]})`)
  })
})

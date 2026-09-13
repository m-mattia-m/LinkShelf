import { mountSuspended, renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import ThemeForm from './ThemeForm.vue'

describe('ThemeForm', () => {
  it('initializes its fields from the "initial" prop', async () => {
    await renderSuspended(ThemeForm, {
      props: { initial: { name: 'My Theme', config: '--shelf-bg: #000;' } }
    })

    expect(screen.getByLabelText('Name')).toHaveValue('My Theme')
    expect(screen.getByLabelText('Config')).toHaveValue('--shelf-bg: #000;')
  })

  it('initializes its fields from the "modelValue" prop, taking precedence over "initial"', async () => {
    await renderSuspended(ThemeForm, {
      props: {
        initial: { name: 'Ignored', config: 'ignored' },
        modelValue: { id: 'theme-1', name: 'Existing Theme', config: '--shelf-text: #fff;', scope: 'user' }
      }
    })

    expect(screen.getByLabelText('Name')).toHaveValue('Existing Theme')
    expect(screen.getByLabelText('Config')).toHaveValue('--shelf-text: #fff;')
  })

  it('emits update:modelValue with the current field values as they change', async () => {
    const { emitted } = await renderSuspended(ThemeForm)

    await fireEvent.update(screen.getByLabelText('Name'), 'New Name')

    const events = emitted()['update:modelValue'] as unknown[][] | undefined
    expect(events).toBeTruthy()
    const lastEvent = events![events!.length - 1]!
    expect(lastEvent[0]).toMatchObject({ name: 'New Name' })
  })

  it('lists the available theme CSS custom properties for reference', async () => {
    await renderSuspended(ThemeForm)

    expect(screen.getByText('--shelf-bg')).toBeInTheDocument()
    expect(screen.getByText('--shelf-font-family')).toBeInTheDocument()
  })

  describe('exposed validate()', () => {
    it('fails when required fields are empty', async () => {
      const wrapper = await mountSuspended(ThemeForm, {
        props: { initial: { name: '', config: '' } }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(false)
    })

    it('succeeds when required fields are filled', async () => {
      const wrapper = await mountSuspended(ThemeForm, {
        props: { initial: { name: 'Valid', config: '--x: 1;' } }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(true)
    })
  })
})

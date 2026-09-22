import { mountSuspended, renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ThemeGroupedResponseBodyToJSON } from '~~/api'
import type { SettingPageBody, Shelf } from '~~/api'
import { server } from '../../../test/mocks/server'
import { buildSettingPageBody, buildTheme, buildUser } from '../../../test/mocks/factories'
import { useThemeStore } from '~/stores/theme'
import ShelfForm from './ShelfForm.vue'

const BASE = 'http://localhost:8085'

function buildShelfProp(overrides: Partial<Shelf> = {}): Shelf {
  return {
    id: 'shelf-1',
    title: 'My Shelf',
    description: 'A shelf',
    domain: '',
    path: 'my-shelf',
    icon: 'i-lucide-book',
    theme: {},
    themeId: '',
    themeMissing: false,
    userId: 'user-1',
    ...overrides
  } as Shelf
}

beforeEach(() => {
  const themeStore = useThemeStore()
  themeStore.$reset()
  // ShelfForm fetches themes on mount whenever the store isn't loaded yet.
  // Default to "already loaded" so most tests don't trigger that fetch (and
  // its unawaited promise can't bleed into a later test) - the "theme
  // selection" tests below opt back into the unloaded state explicitly.
  themeStore.loaded = true
})

afterEach(() => {
  // vi.spyOn() is idempotent - without restoring, a later test's spyOn on
  // the same store action would reuse the previous test's spy and its call
  // history instead of starting fresh.
  vi.restoreAllMocks()
})

describe('ShelfForm', () => {
  it('initializes its fields from the "modelValue" prop', async () => {
    await renderSuspended(ShelfForm, {
      props: { modelValue: buildShelfProp({ title: 'Existing Shelf', description: 'Existing description', path: 'existing-path' }) }
    })

    expect(screen.getByLabelText('Title')).toHaveValue('Existing Shelf')
    expect(screen.getByLabelText('Description')).toHaveValue('Existing description')
    // "Path" is also the label of the (unrelated) Path tab trigger, so
    // getByLabelText is ambiguous here - the textbox role disambiguates.
    expect(screen.getByRole('textbox', { name: 'Path' })).toHaveValue('existing-path')
  })

  it('shows the domain field value after switching to the Domain tab', async () => {
    await renderSuspended(ShelfForm, {
      props: { modelValue: buildShelfProp({ domain: 'example.com' }) }
    })

    // Reka UI's TabsTrigger switches tabs on a left mousedown. testing-library's
    // fireEvent.mouseDown doesn't reproduce this reliably here, so dispatch a
    // real MouseEvent directly.
    const domainTab = screen.getByRole('tab', { name: 'Domain' })
    domainTab.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, button: 0 }))
    // A plain nextTick() isn't enough for Reka UI's Tabs to finish switching
    // panels here, so give it a macrotask to settle.
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(screen.getByRole('textbox', { name: 'Domain' })).toHaveValue('example.com')
  })

  it('emits update:modelValue with the current field values as they change', async () => {
    const { emitted } = await renderSuspended(ShelfForm, {
      props: { modelValue: buildShelfProp() }
    })

    await fireEvent.update(screen.getByLabelText('Title'), 'New Title')

    const events = emitted()['update:modelValue'] as unknown[][] | undefined
    expect(events).toBeTruthy()
    const lastEvent = events![events!.length - 1]!
    expect(lastEvent[0]).toMatchObject({ title: 'New Title' })
  })

  describe('hide from search engines', () => {
    it('defaults to off for a shelf that has never set it', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ noIndex: false }) }
      })

      expect(screen.getByLabelText('Hide from search engines')).not.toBeChecked()
    })

    it('reflects an existing shelf that already opted out of indexing', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ noIndex: true }) }
      })

      expect(screen.getByLabelText('Hide from search engines')).toBeChecked()
    })

    it('emits the toggled value', async () => {
      const { emitted } = await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ noIndex: false }) }
      })

      await fireEvent.click(screen.getByLabelText('Hide from search engines'))

      const events = emitted()['update:modelValue'] as unknown[][]
      expect(events[events.length - 1]![0]).toMatchObject({ noIndex: true })
    })
  })

  // The Select's visible current-value text is duplicated by a hidden
  // native <option> Reka UI renders for form semantics, so scope to the
  // visible value slot to avoid ambiguous text matches.
  function selectedThemeLabel(text: string) {
    return screen.getByText(text, { selector: '[data-slot="value"]' })
  }

  describe('theme selection', () => {
    it('fetches themes on mount when the theme store has not loaded yet', async () => {
      const themeStore = useThemeStore()
      themeStore.loaded = false
      const fetchSpy = vi.spyOn(themeStore, 'fetch')
      server.use(http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({ instance: [], mine: [] }))))

      await renderSuspended(ShelfForm)

      await waitFor(() => {
        expect(fetchSpy).toHaveBeenCalledTimes(1)
      })
      // Wait out the fetch's own promise (not just the call) so it can't
      // resolve mid-flight into the next test's freshly-reset store.
      await fetchSpy.mock.results[0]!.value
      expect(selectedThemeLabel('No theme (default look)')).toBeInTheDocument()
    })

    it('does not refetch themes when the theme store is already loaded', async () => {
      const themeStore = useThemeStore()
      themeStore.loaded = true
      themeStore.mine = [buildTheme({ id: 'theme-mine', name: 'My Theme' })]
      const fetchSpy = vi.spyOn(themeStore, 'fetch')

      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ themeId: 'theme-mine' }) }
      })

      expect(fetchSpy).not.toHaveBeenCalled()
      expect(selectedThemeLabel('My Theme')).toBeInTheDocument()
    })

    it('shows a warning and falls back to "No theme" when the shelf theme is missing', async () => {
      const themeStore = useThemeStore()
      themeStore.loaded = true

      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ themeId: 'gone-theme', themeMissing: true }) }
      })

      expect(screen.getByText('Theme unavailable')).toBeInTheDocument()
      expect(selectedThemeLabel('No theme (default look)')).toBeInTheDocument()
    })
  })

  describe('exposed validate()', () => {
    it('fails when the title is empty', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: '', path: 'a-path' }) }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(false)
    })

    it('fails when neither domain nor path is provided', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: '', domain: '' }) }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(false)
    })

    it('fails when the path contains invalid characters', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: 'not a valid path!', domain: '' }) }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(false)
    })

    it('fails when the domain has an invalid format', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: '', domain: 'not-a-domain' }) }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(false)
    })

    it('succeeds when required fields are valid', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: 'valid-path' }) }
      })

      const isValid = await wrapper.vm.validate()

      expect(isValid).toBe(true)
    })
  })

  describe('a path or a domain, never both', () => {
    function lastEmitted(emitted: () => Record<string, unknown[]>) {
      const events = emitted()['update:modelValue'] as unknown[][]
      return events[events.length - 1]![0] as Record<string, unknown>
    }

    async function openDomainTab() {
      screen.getByRole('tab', { name: 'Domain' })
        .dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true, button: 0 }))
      await new Promise(resolve => setTimeout(resolve, 0))
    }

    it('opens on the Domain tab for a shelf that only has a domain', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: '', domain: 'profile.example.com' }) }
      })

      expect(screen.getByRole('textbox', { name: 'Domain' })).toHaveValue('profile.example.com')
      expect(screen.queryByRole('textbox', { name: 'Path' })).not.toBeInTheDocument()
    })

    it('sends only the path while the Path tab is open', async () => {
      const { emitted } = await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'my-shelf', domain: '' }) }
      })

      await fireEvent.update(screen.getByLabelText('Title'), 'Renamed')

      expect(lastEmitted(emitted)).toMatchObject({ title: 'Renamed', path: 'my-shelf', domain: '' })
    })

    it('sends only the domain, normalized, while the Domain tab is open', async () => {
      const { emitted } = await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'my-shelf', domain: '' }) }
      })

      await openDomainTab()
      await fireEvent.update(screen.getByRole('textbox', { name: 'Domain' }), '  Profile.Example.COM.:443/ ')

      expect(lastEmitted(emitted)).toMatchObject({ path: '', domain: 'profile.example.com' })
    })

    it('rewrites the typed domain to its normalized form when the field loses focus', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: '', domain: 'profile.example.com' }) }
      })

      const input = screen.getByRole('textbox', { name: 'Domain' })
      await fireEvent.update(input, 'PROFILE.Example.com:9443/')
      await fireEvent.blur(input)

      expect(input).toHaveValue('profile.example.com:9443')
    })

    it('shows the address the domain will be reached at', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: '', domain: 'Profile.example.com' }) }
      })

      expect(screen.getByText(/https:\/\/profile\.example\.com - point the domain/)).toBeInTheDocument()
    })

    it('warns about a shelf that has both and says which one saving keeps', async () => {
      const { emitted } = await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'my-shelf', domain: 'profile.example.com' }) }
      })

      expect(screen.getByText('Path and domain')).toBeInTheDocument()
      expect(screen.getByText(/Saving keeps the path and clears the other/)).toBeInTheDocument()

      await fireEvent.update(screen.getByLabelText('Title'), 'Renamed')
      expect(lastEmitted(emitted)).toMatchObject({ path: 'my-shelf', domain: '' })
    })

    it('does not warn about a shelf that has only one', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'my-shelf', domain: '' }) }
      })

      expect(screen.queryByText('Path and domain')).not.toBeInTheDocument()
    })

    it('validates the domain, and only on the Domain tab', async () => {
      const valid = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: '', domain: 'Profile.Example.com:9443' }) }
      })
      expect(await valid.vm.validate()).toBe(true)

      for (const domain of ['localhost', '1.2.3.4', '*.example.com', 'https://a.example.com', 'a.example.com:0']) {
        const invalid = await mountSuspended(ShelfForm, {
          props: { modelValue: buildShelfProp({ title: 'Title', path: '', domain }) }
        })
        expect(await invalid.vm.validate(), domain).toBe(false)
      }
    })

    it('requires a domain on the Domain tab and a path on the Path tab', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: '', domain: '' }) }
      })
      // No domain, and the form defaults to the Path tab, which is empty too.
      expect(await wrapper.vm.validate()).toBe(false)
    })

    it('ignores what is typed on the tab that is not in use', async () => {
      // A shelf from before the exclusive rule: the junk domain is dropped on save.
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Title', path: 'valid-path', domain: 'not a domain' }) }
      })

      expect(await wrapper.vm.validate()).toBe(true)
    })
  })

  describe('path rules and URL', () => {
    function setUserBasedPaths(enabled: boolean) {
      useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths: enabled })
    }

    beforeEach(() => {
      setUserBasedPaths(false)
      useAuthStore().$reset()
    })

    it('rejects a word the app itself answers as a top-level path', async () => {
      for (const path of ['docs', 'app', 'auth', 'cloud', 'API']) {
        const wrapper = await mountSuspended(ShelfForm, {
          props: { modelValue: buildShelfProp({ path: 'old-path' }) }
        })
        await fireEvent.update(wrapper.get('input[name="path"]').element as HTMLInputElement, path)

        expect(await wrapper.vm.validate(), path).toBe(false)
      }
    })

    it('shows why the path is rejected', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'old-path' }) }
      })

      await fireEvent.update(screen.getByRole('textbox', { name: 'Path' }), 'docs')
      await fireEvent.click(document.body)

      const form = document.querySelector('form')!
      form.dispatchEvent(new Event('submit', { cancelable: true }))

      await waitFor(() => {
        expect(screen.getByText('This path is used by a page of this app. Please choose another one')).toBeInTheDocument()
      })
    })

    it('accepts the same word behind a username', async () => {
      setUserBasedPaths(true)
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'old-path' }) }
      })
      await fireEvent.update(wrapper.get('input[name="path"]').element as HTMLInputElement, 'docs')

      expect(await wrapper.vm.validate()).toBe(true)
    })

    it('leaves a shelf editable that already has a reserved path', async () => {
      const wrapper = await mountSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ title: 'Renamed title', path: 'docs' }) }
      })

      expect(await wrapper.vm.validate()).toBe(true)
    })

    it('shows the URL without a username while user-based paths are off', async () => {
      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'my-shelf', username: 'alice' }) }
      })

      expect(screen.getByText(`${window.location.origin}/my-shelf`)).toBeInTheDocument()
    })

    it('shows the owner\'s username in the URL while user-based paths are on', async () => {
      setUserBasedPaths(true)

      await renderSuspended(ShelfForm, {
        props: { modelValue: buildShelfProp({ path: 'my-shelf', username: 'alice' }) }
      })

      expect(screen.getByText(`${window.location.origin}/alice/my-shelf`)).toBeInTheDocument()
    })

    it('uses the signed-in user for a shelf that is not saved yet', async () => {
      setUserBasedPaths(true)
      useAuthStore().user = buildUser({ id: 'user-1', username: 'bob' })

      await renderSuspended(ShelfForm)
      await fireEvent.update(screen.getByRole('textbox', { name: 'Path' }), 'fresh')

      expect(screen.getByText(`${window.location.origin}/bob/fresh`)).toBeInTheDocument()
    })
  })
})

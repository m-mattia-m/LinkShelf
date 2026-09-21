import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import type { SettingPageBody, Shelf } from '~~/api'
import { buildSettingPageBody, buildShelf } from '../../../test/mocks/factories'
import ShelfUrlAlert from './ShelfUrlAlert.vue'

function setUserBasedPaths(enabled: boolean) {
  useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths: enabled })
}

async function renderAlert(shelf: Shelf) {
  return renderSuspended(ShelfUrlAlert, { props: { shelf } })
}

const NOW_USER_BASED = 'This shelf\'s URL now includes your username'
const NO_LONGER_USER_BASED = 'This shelf\'s URL no longer includes a username'
const RESERVED = 'This path can\'t be reached'

// Nuxt UI's alert has no ARIA role, so "nothing is shown" has to be checked
// against the alert titles themselves.
function expectNoAlert() {
  for (const title of [NOW_USER_BASED, NO_LONGER_USER_BASED, RESERVED]) {
    expect(screen.queryByText(title)).not.toBeInTheDocument()
  }
}

beforeEach(() => {
  setUserBasedPaths(false)
})

describe('ShelfUrlAlert', () => {
  describe('a shelf created under the same setting as now', () => {
    it('shows nothing while user-based paths are off', async () => {
      await renderAlert(buildShelf({ path: 'my-links', username: 'alice', createdWithUserBasedPaths: false }))

      expectNoAlert()
    })

    it('shows nothing while user-based paths are on', async () => {
      setUserBasedPaths(true)

      await renderAlert(buildShelf({ path: 'my-links', username: 'alice', createdWithUserBasedPaths: true }))

      expectNoAlert()
    })
  })

  describe('a shelf created before user-based paths were switched on', () => {
    it('explains that the URL now has the username in it and that old links stopped working', async () => {
      setUserBasedPaths(true)

      await renderAlert(buildShelf({ path: 'my-links', username: 'alice', createdWithUserBasedPaths: false }))

      expect(screen.getByText(NOW_USER_BASED)).toBeInTheDocument()
      expect(screen.getByText(/available at .*\/alice\/my-links\./)).toBeInTheDocument()
      expect(screen.getByText(/Links without the username, like \/my-links, no longer work/)).toBeInTheDocument()
    })
  })

  describe('a shelf created before user-based paths were switched off', () => {
    it('explains that the username left the URL and that links containing it stopped working', async () => {
      await renderAlert(buildShelf({ path: 'my-links', username: 'alice', createdWithUserBasedPaths: true }))

      expect(screen.getByText(NO_LONGER_USER_BASED)).toBeInTheDocument()
      expect(screen.getByText(/available at .*\/my-links\./)).toBeInTheDocument()
      expect(screen.getByText(/Links with a username, like \/alice\/my-links, no longer work/)).toBeInTheDocument()
    })
  })

  describe('a reserved top-level path', () => {
    it('warns that the shelf cannot be reached', async () => {
      await renderAlert(buildShelf({ path: 'docs', username: 'alice' }))

      expect(screen.getByText(RESERVED)).toBeInTheDocument()
      expect(screen.getByText(/\/docs belongs to a page of this app/)).toBeInTheDocument()
    })

    it('is reported instead of the URL notice, since it is the bigger problem', async () => {
      // Created while user-based paths were on, now off, and the path is reserved.
      await renderAlert(buildShelf({ path: 'docs', username: 'alice', createdWithUserBasedPaths: true }))

      expect(screen.getByText(RESERVED)).toBeInTheDocument()
      expect(screen.queryByText(NO_LONGER_USER_BASED)).not.toBeInTheDocument()
    })

    it('is not a problem behind a username', async () => {
      setUserBasedPaths(true)

      await renderAlert(buildShelf({ path: 'docs', username: 'alice', createdWithUserBasedPaths: true }))

      expect(screen.queryByText(RESERVED)).not.toBeInTheDocument()
    })
  })

  it('shows nothing for a shelf that only has a domain, whatever the settings say', async () => {
    setUserBasedPaths(true)

    await renderAlert(buildShelf({ path: '', domain: 'example.com', username: 'alice', createdWithUserBasedPaths: false }))

    expectNoAlert()
  })
})

describe('ShelfUrlAlert with an older backend', () => {
  it('says nothing when the shelf carries no creation mode, instead of guessing', async () => {
    // A backend that predates the field leaves it out of the response, and
    // the generated client then hands it over as undefined.
    setUserBasedPaths(true)
    const shelf = buildShelf({ path: 'my-links', username: 'alice' })
    delete (shelf as Partial<Shelf>).createdWithUserBasedPaths

    await renderAlert(shelf)

    expectNoAlert()
  })
})

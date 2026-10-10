import type { SettingPageBody } from '~~/api'

// The caller's shelf limit. Admins see every shelf in the store, so only own ones count.
export function useShelfLimit() {
  const authStore = useAuthStore()
  const shelfStore = useShelfStore()
  const settings = useState('settings') as unknown as Ref<SettingPageBody | null>

  const maxShelves = computed(() => authStore.user?.maxShelves ?? null)
  const count = computed(() => shelfStore.shelves.filter(shelf => shelf.userId === authStore.userId).length)
  const reached = computed(() => maxShelves.value !== null && count.value >= maxShelves.value)
  const upgradeUrl = computed(() => safeHref(settings.value?.upgradeUrl))

  return { maxShelves, count, reached, upgradeUrl }
}

/**
 * `callOnce()` (used by several pages for their initial data fetch) dedupes
 * by a key stored on `nuxtApp.payload.once` / `nuxtApp._once`. The "nuxt"
 * vitest environment reuses one Nuxt app per test FILE, so without this,
 * only the first test that mounts a `callOnce`-using page actually runs the
 * fetch - every subsequent test in the same file sees stale (or absent)
 * data. Call this in `beforeEach` for any spec that mounts such a page.
 */
export function resetOnceCache(): void {
  const nuxtApp = useNuxtApp()
  nuxtApp.payload.once.clear()
  nuxtApp._once = {}
}

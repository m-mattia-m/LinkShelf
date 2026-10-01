/** Resets callOnce() keys between tests; call it in beforeEach. */
export function resetOnceCache(): void {
  const nuxtApp = useNuxtApp()
  nuxtApp.payload.once.clear()
  nuxtApp._once = {}
}

<script setup lang="ts">
const { t } = useI18n()

// Set by the page component (via the same key) once it resolves the
// shelf's theme - a plain object of "--shelf-*" CSS custom properties, or
// {} when no theme is selected/it no longer exists. Shared via useState
// (rather than props) because a Nuxt layout can't receive props from its
// page, and CSS custom properties only cascade to descendants, so they have
// to be set here on the layout's own root element - the page's own content
// is a descendant of it and picks them up automatically.
const themeVars = useState<Record<string, string>>('public-shelf-theme-vars', () => ({}))

const rootStyle = computed(() => {
  const vars = themeVars.value ?? {}
  const style: Record<string, string> = {}

  for (const [key, value] of Object.entries(vars)) {
    if (key === '--shelf-bg-image') continue
    style[key] = value
  }
  if (vars['--shelf-font-family']) style.fontFamily = vars['--shelf-font-family']
  if (vars['--shelf-bg-image']) style.backgroundImage = `url(${vars['--shelf-bg-image']})`

  return style
})
</script>

<template>
  <div
    class="min-h-screen flex flex-col bg-[var(--shelf-bg,#f5f5f4)] text-[var(--shelf-text,#1c274c)] bg-cover bg-center"
    :style="rootStyle"
  >
    <main class="flex-1 flex items-start justify-center px-4 py-12">
      <slot />
    </main>

    <footer class="py-6 flex justify-center">
      <a
        href="/"
        class="text-xs opacity-60 hover:opacity-100 transition-opacity"
      >
        {{ t('linkpage.poweredBy') }}
      </a>
    </footer>
  </div>
</template>

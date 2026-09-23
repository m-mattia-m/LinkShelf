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

// Shared with ShelfPublicView.vue the same way theme vars are (see above) -
// defaults to "on, no custom text" so the page shows today's only behavior
// (the default footer) until the shelf has actually loaded.
const footer = useState<{ enabled: boolean, customText: string }>('public-shelf-footer', () => ({ enabled: true, customText: '' }))
const footerHtml = computed(() => footer.value.customText ? renderRestrictedMarkdown(footer.value.customText) : '')

const rootStyle = computed(() => {
  const vars = themeVars.value ?? {}
  const style: Record<string, string> = {}

  for (const [key, value] of Object.entries(vars)) {
    if (key === '--shelf-bg' || key === '--shelf-bg-image') continue
    style[key] = value
  }
  if (vars['--shelf-font-family']) style.fontFamily = vars['--shelf-font-family']

  // --shelf-bg may be a plain color or a CSS gradient (see the backend's
  // validateColorOrGradient) - setting both longhands lets each accept
  // whichever one it actually is, since a gradient is only valid for
  // backgroundImage and a plain color only for backgroundColor; the
  // browser silently drops whichever one doesn't apply. Two longhands
  // (rather than the "background" shorthand) also means this can't reset
  // the theme's own --shelf-bg-image below.
  if (vars['--shelf-bg']) {
    style.backgroundColor = vars['--shelf-bg']
    style.backgroundImage = vars['--shelf-bg']
  }
  // A theme's own background photo always wins over --shelf-bg's own
  // image-typed value (a gradient).
  if (vars['--shelf-bg-image']) style.backgroundImage = `url(${vars['--shelf-bg-image']})`

  return style
})
</script>

<template>
  <div
    class="min-h-screen flex flex-col bg-[#f5f5f4] text-[var(--shelf-text,#1c274c)] bg-cover bg-center"
    :style="rootStyle"
  >
    <main class="flex-1 flex items-start justify-center px-4 py-12">
      <slot />
    </main>

    <footer
      v-if="footer.enabled"
      class="py-6 flex justify-center px-4"
    >
      <!-- eslint-disable vue/no-v-html -- footerHtml only ever contains what renderRestrictedMarkdown produces, see its own doc comment -->
      <div
        v-if="footerHtml"
        class="text-xs text-center opacity-60 hover:opacity-100 transition-opacity [&_a]:underline"
        v-html="footerHtml"
      />
      <!-- eslint-enable vue/no-v-html -->
      <a
        v-else
        href="/"
        class="text-xs opacity-60 hover:opacity-100 transition-opacity"
      >
        {{ t('linkpage.poweredBy') }}
      </a>
    </footer>
  </div>
</template>

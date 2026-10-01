<script setup lang="ts">
const { t } = useI18n()

// The theme's CSS properties, set by the page (layouts can't receive props).
const themeVars = useState<Record<string, string>>('public-shelf-theme-vars', () => ({}))

// Set by ShelfPublicView.vue; defaults to the default footer until it loads.
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

  // --shelf-bg may be a color or a gradient; the browser drops whichever doesn't apply.
  if (vars['--shelf-bg']) {
    style.backgroundColor = vars['--shelf-bg']
    style.backgroundImage = vars['--shelf-bg']
  }
  // A theme's background photo wins over a --shelf-bg gradient.
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

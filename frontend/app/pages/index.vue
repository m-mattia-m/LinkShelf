<script setup lang="ts">
import type { ButtonProps } from '#ui/components/Button.vue'

const { t } = useI18n()

const links = computed<ButtonProps[]>(() => [
  {
    label: t('home.welcome.getStarted'),
    color: 'primary',
    to: '/app',
    trailingIcon: 'i-lucide-arrow-right',
    size: 'xl'
  },
  {
    label: t('home.welcome.sourceCode'),
    color: 'neutral',
    variant: 'subtle',
    leadingIcon: 'i-uil-github',
    to: 'https://github.com/m-mattia-m/Linkshelf',
    size: 'xl'
  }
])

const features = computed(() => [
  { icon: 'i-lucide-folder-plus', title: t('home.features.shelves.title'), description: t('home.features.shelves.description') },
  { icon: 'i-lucide-users', title: t('home.features.users.title'), description: t('home.features.users.description') },
  { icon: 'i-lucide-globe', title: t('home.features.domains.title'), description: t('home.features.domains.description') },
  { icon: 'i-lucide-paintbrush', title: t('home.features.themes.title'), description: t('home.features.themes.description') },
  { icon: 'i-lucide-qr-code', title: t('home.features.qrCodes.title'), description: t('home.features.qrCodes.description') },
  { icon: 'i-lucide-file-text', title: t('home.features.footer.title'), description: t('home.features.footer.description') },
  { icon: 'i-lucide-server', title: t('home.features.selfHosting.title'), description: t('home.features.selfHosting.description') },
  { icon: 'i-lucide-layout-dashboard', title: t('home.features.admin.title'), description: t('home.features.admin.description') },
  { icon: 'i-lucide-shield-check', title: t('home.features.oidc.title'), description: t('home.features.oidc.description') }
])

const ctaLinks = computed<ButtonProps[]>(() => [
  {
    label: t('home.cta.docs'),
    to: '/docs/self-hosting/getting-started',
    trailingIcon: 'i-lucide-book-open',
    color: 'primary'
  },
  {
    label: t('home.cta.github'),
    to: 'https://github.com/m-mattia-m/linkshelf',
    target: '_blank',
    icon: 'i-simple-icons-github',
    color: 'neutral',
    variant: 'outline'
  }
])

definePageMeta({
  layout: 'landingpage'
})

// The domain of the shelf this host serves, or null on the instance's own
// host. Set by middleware/shelf-host.global.ts.
const shelfHost = useState<string | null | undefined>('shelf-host')
</script>

<template>
  <ShelfPublicView
    v-if="shelfHost"
    :domain="shelfHost"
  />
  <div v-else>
    <UPageCTA
      :title="$t('home.welcome.title')"
      :description="$t('home.welcome.description')"
      orientation="horizontal"
      :links="links"
      :ui="{ root: 'rounded-none' }"
    >
      <img
        src="/presentation.webp"
        width="320"
        height="364"
        alt="Illustration"
        class="w-full rounded-lg"
      >
    </UPageCTA>

    <UPageSection
      id="features"
      :title="$t('home.features.title')"
      :description="$t('home.features.description')"
      :features="features"
    />

    <UPageSection>
      <UPageCTA
        :title="$t('home.cta.title')"
        :description="$t('home.cta.description')"
        variant="subtle"
        :links="ctaLinks"
      />
    </UPageSection>
  </div>
</template>

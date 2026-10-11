<script setup lang="ts">
definePageMeta({
  layout: 'landingpage'
})

const { t } = useI18n()

const cloudUrl = 'https://account.linkshelf.cloud'
const sponsorUrl = 'https://github.com/sponsors/m-mattia-m'

const features = computed(() => [
  { icon: 'i-lucide-rocket', title: t('cloud.feature.hosted.title'), description: t('cloud.feature.hosted.description') },
  { icon: 'i-lucide-globe', title: t('cloud.feature.domains.title'), description: t('cloud.feature.domains.description') },
  { icon: 'i-lucide-git-fork', title: t('cloud.feature.opensource.title'), description: t('cloud.feature.opensource.description') }
])
</script>

<template>
  <div class="relative overflow-hidden">
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-x-0 -top-24 mx-auto h-96 max-w-3xl rounded-full bg-primary/20 blur-3xl"
    />

    <section class="relative mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 pt-16 pb-12 text-center">
      <UBadge
        :label="t('cloud.badge')"
        icon="i-lucide-cloud"
        color="primary"
        variant="subtle"
        size="lg"
        class="rounded-full"
      />
      <h1 class="mt-6 text-4xl sm:text-6xl font-bold tracking-tight text-highlighted">
        {{ t('cloud.title') }}
      </h1>
      <p class="mt-6 mx-auto max-w-2xl text-lg text-muted">
        {{ t('cloud.description') }}
      </p>
      <div class="mt-8 flex flex-wrap items-center justify-center gap-3">
        <UButton
          :label="t('cloud.open')"
          :to="cloudUrl"
          target="_blank"
          size="xl"
          trailing-icon="i-lucide-arrow-up-right"
        />
        <UButton
          :label="t('cloud.sponsor')"
          :to="sponsorUrl"
          target="_blank"
          size="xl"
          color="neutral"
          variant="subtle"
          leading-icon="i-lucide-heart"
        />
      </div>
      <p class="mt-4 text-sm text-muted">
        <NuxtLink
          :to="cloudUrl"
          target="_blank"
          class="underline underline-offset-4 hover:text-highlighted"
        >{{ t('cloud.learnMore') }}</NuxtLink>
      </p>
    </section>

    <section class="relative mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 pb-12">
      <div class="grid gap-4 sm:grid-cols-3">
        <div
          v-for="feature in features"
          :key="feature.title"
          class="rounded-xl border border-default bg-elevated/50 p-6"
        >
          <div class="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <UIcon
              :name="feature.icon"
              class="size-5"
            />
          </div>
          <h2 class="mt-4 font-semibold text-highlighted">
            {{ feature.title }}
          </h2>
          <p class="mt-1 text-sm text-muted">
            {{ feature.description }}
          </p>
        </div>
      </div>
    </section>

    <section class="relative mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 pb-20">
      <div class="rounded-2xl border border-default bg-gradient-to-br from-primary/10 via-elevated/50 to-elevated/50 p-8 text-center">
        <UIcon
          name="i-lucide-heart-handshake"
          class="size-8 text-primary"
        />
        <h2 class="mt-3 text-2xl font-semibold text-highlighted">
          {{ t('cloud.sponsorTitle') }}
        </h2>
        <p class="mt-2 text-muted">
          {{ t('cloud.sponsorDescription') }}
        </p>
        <UButton
          :label="t('cloud.sponsor')"
          :to="sponsorUrl"
          target="_blank"
          class="mt-6"
          color="neutral"
          leading-icon="i-simple-icons-github"
        />
      </div>
    </section>
  </div>
</template>

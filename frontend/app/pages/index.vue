<script setup lang="ts">
import type {ButtonProps} from "#ui/components/Button.vue";

const links: ButtonProps[] = ref([
  {
    label: 'Get started',
    color: 'primary',
    to: '/app',
    trailingIcon: 'i-lucide-arrow-right',
    size: 'xl'
  },
  {
    label: 'Source code',
    color: 'neutral',
    variant: 'subtle',
    leadingIcon: 'i-uil-github',
    to: 'https://github.com/m-mattia-m/Linkshelf',
    size: 'xl'
  }
])

definePageMeta({
  layout: 'landingpage',
})
</script>

<template>
  <div>
    <UPageCTA
      :title="$t('home.welcome.title')"
      :description="$t('home.welcome.description')"
      orientation="horizontal"
      :links="links"
    >

      <img
        src="https://picsum.photos/640/728"
        width="320"
        height="364"
        alt="Illustration"
        class="w-full rounded-lg"
      />
    </UPageCTA>

    <UPageSection
      id="features"
      title="Everything you need to organize your links"
      description="LinkShelf helps you collect, manage, and share your favorite links with powerful organization tools, flexible customization, and rock-solid infrastructure."
      :features="[{
        icon: 'i-lucide-folder-plus',
        title: 'Unlimited collections',
        description: 'Create and organize unlimited collections of links without restrictions. Keep everything structured exactly the way you want.'
      }, {
        icon: 'i-lucide-user-circle',
        title: 'Accounts & sync',
        description: 'Access your links across all devices with user accounts. Your collections stay in sync wherever you go.'
      }, {
        icon: 'i-lucide-globe',
        title: 'Custom domains',
        description: 'Use your own domain for one or multiple collections and share them professionally with clients, teams, or the public.'
      }, {
        icon: 'i-lucide-paintbrush',
        title: 'Themes & customization',
        description: 'Personalize the look and layout of your collections with multiple themes and flexible customization options.'
      }, {
        icon: 'i-lucide-smartphone',
        title: 'Responsive by design',
        description: 'Optimized for desktop and mobile devices so your links are always accessible, fast, and easy to use.'
      }, {
        icon: 'i-lucide-file-text',
        title: 'Markdown support',
        description: 'Add rich descriptions and notes to your links using Markdown for better context and clarity.'
      }, {
        icon: 'i-lucide-server',
        title: 'Self-host & scale',
        description: 'Self-host LinkShelf with full control over your data. Deploy easily on Kubernetes with support for PostgreSQL and MySQL.'
      }, {
        icon: 'i-lucide-layout-dashboard',
        title: 'Admin dashboard',
        description: 'Manage users, links, collections, and settings from a powerful and intuitive admin interface.'
      }, {
        icon: 'i-lucide-shield-check',
        title: 'Secure & open',
        description: 'Built with security in mind, supporting OpenID Connect (OIDC) authentication and fully open source for transparency and extensibility.'
      }]"

    />

    <UPageSection>
      <UPageCTA
        title="Ready to organize your links?"
        description="Get started with LinkShelf in minutes. Use our managed cloud plan for zero setup, or self-host if you want full control."
        variant="subtle"
        :links="[
          {
            label: 'Start with LinkShelf Cloud',
            to: '/cloud',
            trailingIcon: 'i-lucide-cloud',
            color: 'primary'
          },
          {
            label: 'Self-host on GitHub',
            to: 'https://github.com/m-mattia-m/linkshelf',
            target: '_blank',
            icon: 'i-simple-icons-github',
            color: 'neutral',
            variant: 'outline'
          }
        ]"
      />
    </UPageSection>
  </div>
</template>

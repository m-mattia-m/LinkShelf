<script setup lang="ts">
import type { ContentNavigationLink } from '#ui/components/content/ContentNavigation.vue'

const { data: page } = await useAsyncData(() => queryCollection('docs').all())
if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Page not found', fatal: true })
}

// `exact` is forwarded to the underlying link: without it "/docs" counts as
// active on every "/docs/..." page.
type DocsNavLink = ContentNavigationLink & { exact?: boolean, _order?: number }

interface DocsNavItem {
  navigation?: boolean
  order?: number
  path: string
  title: string
  meta?: {
    icon?: string
    order?: number
    metadata?: { order?: number }
  }
}

const navigation = computed(() => {
  return buildNavigation((page.value || []) as DocsNavItem[])
})

function buildNavigation(items: DocsNavItem[]): ContentNavigationLink[] {
  const nav: Record<string, DocsNavLink> = {}

  const getOrder = (item?: DocsNavItem) =>
    Number(
      item?.order
      ?? item?.meta?.order
      ?? item?.meta?.metadata?.order
      ?? Infinity
    )

  const sorted = items
    .filter(item => item.navigation)
    .sort((a, b) => {
      const ao = getOrder(a)
      const bo = getOrder(b)
      if (ao !== bo) return ao - bo
      return a.title.localeCompare(b.title)
    })

  for (const item of sorted) {
    const segments = item.path.replace(/^\/docs\/?/, '').split('/')

    if (segments[0] === '') {
      nav.__root__ = {
        title: item.title,
        path: item.path,
        icon: item.meta?.icon,
        exact: true,
        _order: getOrder(item)
      }
      continue
    }

    const sectionKey = segments[0]!
    const isSectionRoot = segments.length === 1

    nav[sectionKey] ??= {
      title: item.title,
      path: `/docs/${sectionKey}`,
      icon: item.meta?.icon,
      exact: true,
      children: [],
      _order: Infinity
    }

    if (isSectionRoot) {
      nav[sectionKey].title = item.title
      nav[sectionKey].icon = item.meta?.icon
      nav[sectionKey]._order = getOrder(item)
      continue
    }

    nav[sectionKey].children!.push({
      title: item.title,
      path: item.path,
      exact: true
    } as DocsNavLink)
  }

  for (const section of Object.values(nav)) {
    section.children?.sort((a, b) => {
      const ao = getOrder(items.find(i => i.path === a.path))
      const bo = getOrder(items.find(i => i.path === b.path))
      if (ao !== bo) return ao - bo
      return a.title.localeCompare(b.title)
    })
  }

  return Object.values(nav)
    .sort((a, b) => {
      if (a.path === '/docs') return -1
      if (b.path === '/docs') return 1

      const ao = a._order ?? Infinity
      const bo = b._order ?? Infinity
      if (ao !== bo) return ao - bo

      return a.title.localeCompare(b.title)
    })
    .map(({ _order, ...item }) => item)
}
</script>

<template>
  <div class="m-0 lg:m-8">
    <UContentNavigation
      :navigation="navigation"
      color="neutral"
      link
    />
  </div>
</template>

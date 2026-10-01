<script setup lang="ts">
import type { SettingPageBody, Shelf } from '~~/api'

const props = defineProps<{
  shelf: Shelf
}>()

const { t } = useI18n()
const origin = useRequestURL().origin
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>

const userBasedPaths = computed(() => websiteSettings.value?.userBasedPaths ?? false)

// Shown when a path shelf's URL changed shape; a reserved path is reported first.
const state = computed<'nowUserBased' | 'noLongerUserBased' | 'reserved' | null>(() => {
  if (!props.shelf.path) return null
  if (!userBasedPaths.value && isRouteReservedPath(props.shelf.path)) return 'reserved'
  // Older backends don't send the creation mode.
  if (typeof props.shelf.createdWithUserBasedPaths !== 'boolean') return null
  if (props.shelf.createdWithUserBasedPaths === userBasedPaths.value) return null
  return userBasedPaths.value ? 'nowUserBased' : 'noLongerUserBased'
})
</script>

<template>
  <UAlert
    v-if="state === 'nowUserBased'"
    color="info"
    variant="subtle"
    icon="i-lucide-link"
    :title="t('app.shelf.urlAlert.nowUserBased.title')"
    :description="t('app.shelf.urlAlert.nowUserBased.description', { url: origin + publicShelfPath(shelf, true), path: shelf.path })"
  />
  <UAlert
    v-else-if="state === 'noLongerUserBased'"
    color="info"
    variant="subtle"
    icon="i-lucide-link"
    :title="t('app.shelf.urlAlert.noLongerUserBased.title')"
    :description="t('app.shelf.urlAlert.noLongerUserBased.description', { url: origin + publicShelfPath(shelf, false), username: shelf.username, path: shelf.path })"
  />
  <UAlert
    v-else-if="state === 'reserved'"
    color="warning"
    variant="subtle"
    icon="i-lucide-triangle-alert"
    :title="t('app.shelf.urlAlert.reserved.title')"
    :description="t('app.shelf.urlAlert.reserved.description', { path: shelf.path })"
  />
</template>

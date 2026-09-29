<script setup lang="ts">
import type { Theme } from '~~/api'

definePageMeta({
  layout: 'app',
  middleware: 'admin'
})

const { t } = useI18n()
const loading = ref(true)
const themes = ref<Theme[]>([])

async function load() {
  loading.value = true
  try {
    const api = useApi()
    themes.value = (await api.theme.listAllUserThemes()) ?? []
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
}

onMounted(load)

const deleteOpen = ref(false)
const deletingTheme = ref<Theme | null>(null)
const deleting = ref(false)

function openDelete(theme: Theme) {
  deletingTheme.value = theme
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deletingTheme.value) return
  deleting.value = true
  try {
    const api = useApi()
    await api.theme.deleteTheme({ themeId: deletingTheme.value.id })
    deleteOpen.value = false
    await load()
  } catch (err) {
    await handleApiError(err)
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl text-highlighted pb-2">
      {{ t('app.settings.themes.title') }}
    </h1>
    <p class="text-sm text-muted pb-6">
      {{ t('app.settings.themes.intro') }}
    </p>

    <div
      v-if="loading"
      class="space-y-2"
    >
      <USkeleton
        v-for="i in 3"
        :key="i"
        class="h-10 w-full"
      />
    </div>

    <div
      v-else-if="themes.length === 0"
      class="text-center text-muted py-16"
    >
      {{ t('app.settings.themes.empty') }}
    </div>

    <div
      v-else
      class="flex flex-col divide-y divide-default rounded-lg border border-default"
    >
      <div
        v-for="theme in themes"
        :key="theme.id"
        class="flex items-center justify-between px-4 py-2.5 gap-4"
      >
        <div class="min-w-0">
          <p class="font-medium truncate">
            {{ theme.name }}
          </p>
          <p class="text-xs text-dimmed truncate">
            {{ t('app.settings.themes.owner', { owner: theme.ownerUserId }) }}
          </p>
        </div>
        <UButton
          icon="i-lucide-trash-2"
          size="xs"
          color="error"
          variant="ghost"
          :aria-label="t('app.settings.themes.delete')"
          @click="openDelete(theme)"
        />
      </div>
    </div>

    <ConfirmDialog
      v-model:open="deleteOpen"
      :title="t('app.theme.deleteConfirm.title')"
      :description="t('app.theme.deleteConfirm.description', { name: deletingTheme?.name })"
      :loading="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>

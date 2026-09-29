<script setup lang="ts">
import type { Theme, ThemeBase } from '~~/api'
import { useThemeStore } from '~/stores/theme'
import ThemeFormDialog from '~/components/theme/ThemeFormDialog.vue'

definePageMeta({
  layout: 'app'
})

const { t } = useI18n()
const themeStore = useThemeStore()
const loading = ref(true)

onMounted(async () => {
  try {
    await callOnce(themeStore.fetch)
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
})

const editOpen = ref(false)
const editingTheme = ref<Theme>()

function openEdit(theme: Theme) {
  editingTheme.value = theme
  editOpen.value = true
}

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
    await themeStore.remove(deletingTheme.value.id)
    deleteOpen.value = false
  } catch (err) {
    await handleApiError(err)
  } finally {
    deleting.value = false
  }
}

function exportTheme(theme: Theme) {
  const blob = new Blob([theme.config], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${theme.name.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-') || 'theme'}.txt`
  a.click()
  URL.revokeObjectURL(url)
}

function actionItems(theme: Theme) {
  return [
    [{ label: t('common.edit'), icon: 'i-lucide-pencil', onSelect: () => openEdit(theme) }],
    [{ label: t('app.theme.export'), icon: 'i-lucide-download', onSelect: () => exportTheme(theme) }],
    [{ label: t('common.delete'), icon: 'i-lucide-trash-2', color: 'error' as const, onSelect: () => openDelete(theme) }]
  ]
}

// Import opens the create dialog pre-filled from an exported config file. It
// always creates a new theme.
const importFileInput = ref<HTMLInputElement>()
const importInitial = ref<ThemeBase>()
const importOpen = ref(false)

function triggerImport() {
  importFileInput.value?.click()
}

async function onImportFileSelected(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  const config = await file.text()
  const name = file.name.replace(/\.[^.]+$/, '')
  importInitial.value = { name, config }
  importOpen.value = true

  ;(event.target as HTMLInputElement).value = ''
}
</script>

<template>
  <div>
    <div class="flex justify-between items-center gap-2">
      <h1 class="text-2xl text-highlighted pb-4">
        {{ t('app.theme.title') }}
      </h1>

      <div class="flex items-center gap-2">
        <UButton
          :label="t('app.theme.import')"
          icon="i-lucide-upload"
          color="neutral"
          variant="outline"
          @click="triggerImport"
        />
        <input
          ref="importFileInput"
          type="file"
          accept=".txt,.css,text/plain"
          class="hidden"
          @change="onImportFileSelected"
        >
        <ThemeFormDialog mode="create" />
      </div>
    </div>

    <p class="text-sm text-muted pb-6">
      {{ t('app.theme.intro') }}
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

    <template v-else>
      <div
        v-if="themeStore.instance.length"
        class="pb-6"
      >
        <h2 class="text-sm font-medium text-dimmed pb-2">
          {{ t('app.shelf.form.instanceThemes') }}
        </h2>
        <p class="text-xs text-dimmed pb-2">
          {{ t('app.theme.instanceHelp') }}
        </p>
        <div class="flex flex-col divide-y divide-default rounded-lg border border-default">
          <div
            v-for="theme in themeStore.instance"
            :key="theme.id"
            class="flex items-center justify-between px-4 py-2.5"
          >
            <span class="font-medium">{{ theme.name }}</span>
            <UBadge
              color="neutral"
              variant="subtle"
            >
              {{ t('app.theme.instanceBadge') }}
            </UBadge>
          </div>
        </div>
      </div>

      <div>
        <h2 class="text-sm font-medium text-dimmed pb-2">
          {{ t('app.shelf.form.yourThemes') }}
        </h2>

        <div
          v-if="themeStore.mine.length === 0"
          class="flex flex-col items-center gap-4 py-16 text-center"
        >
          <p class="text-muted">
            {{ t('app.theme.empty') }}
          </p>
          <ThemeFormDialog mode="create" />
        </div>

        <div
          v-else
          class="flex flex-col divide-y divide-default rounded-lg border border-default"
        >
          <div
            v-for="theme in themeStore.mine"
            :key="theme.id"
            class="flex items-center justify-between px-4 py-2.5"
          >
            <span class="font-medium">{{ theme.name }}</span>
            <UDropdownMenu :items="actionItems(theme)">
              <UButton
                icon="i-lucide-ellipsis-vertical"
                color="neutral"
                variant="ghost"
                :aria-label="t('common.actions')"
              />
            </UDropdownMenu>
          </div>
        </div>
      </div>
    </template>

    <ThemeFormDialog
      v-model:open="editOpen"
      mode="edit"
      :theme="editingTheme"
    />

    <ThemeFormDialog
      v-model:open="importOpen"
      mode="create"
      hide-trigger
      :initial="importInitial"
    />

    <ConfirmDialog
      v-model:open="deleteOpen"
      :title="t('app.theme.deleteConfirm.title')"
      :description="t('app.theme.deleteConfirm.description', { name: deletingTheme?.name })"
      :loading="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>

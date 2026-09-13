<script setup lang="ts">
import type { Theme, ThemeBase } from '~~/api'
import { useThemeStore } from '~/stores/theme'
import ThemeFormDialog from '~/components/theme/ThemeFormDialog.vue'

definePageMeta({
  layout: 'app'
})

const themeStore = useThemeStore()
const loading = ref(true)

onMounted(async () => {
  await callOnce(themeStore.fetch)
  loading.value = false
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
    [{ label: 'Edit', icon: 'i-lucide-pencil', onSelect: () => openEdit(theme) }],
    [{ label: 'Export', icon: 'i-lucide-download', onSelect: () => exportTheme(theme) }],
    [{ label: 'Delete', icon: 'i-lucide-trash-2', color: 'error' as const, onSelect: () => openDelete(theme) }]
  ]
}

/**
 * Import reads a plain-text config file the user exported earlier (or wrote
 * by hand) and opens the create dialog pre-filled with it - it always
 * creates a NEW theme, never overwrites an existing one. Validation is the
 * same as the regular create flow (server-side, same schema as the editor).
 */
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
        Themes
      </h1>

      <div class="flex items-center gap-2">
        <UButton
          label="Import"
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
      Themes only affect a shelf's public page, never the app itself. Your themes are private - if you want to share one, export it and send the file to whoever wants to import it.
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
          Instance themes
        </h2>
        <p class="text-xs text-dimmed pb-2">
          Provided by your instance admin. Managed via server config, not here.
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
              Instance
            </UBadge>
          </div>
        </div>
      </div>

      <div>
        <h2 class="text-sm font-medium text-dimmed pb-2">
          Your themes
        </h2>

        <div
          v-if="themeStore.mine.length === 0"
          class="flex flex-col items-center gap-4 py-16 text-center"
        >
          <p class="text-muted">
            You haven't created any themes yet.
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
                aria-label="Actions"
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
      title="Delete theme?"
      :description="`Delete “${deletingTheme?.name}”? Any shelf using it will fall back to the default look. This cannot be undone.`"
      :loading="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import type { Theme } from '~~/api'

definePageMeta({
  layout: 'app',
  middleware: 'admin'
})

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
      User themes
    </h1>
    <p class="text-sm text-muted pb-6">
      Every user-created theme across the instance. You can remove one (a shelf using it falls back to the default look), but not edit its content - it belongs to its creator.
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
      No user-created themes yet.
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
            Owner: {{ theme.ownerUserId }}
          </p>
        </div>
        <UButton
          icon="i-lucide-trash-2"
          size="xs"
          color="error"
          variant="ghost"
          aria-label="Delete theme"
          @click="openDelete(theme)"
        />
      </div>
    </div>

    <ConfirmDialog
      v-model:open="deleteOpen"
      title="Delete theme?"
      :description="`Delete “${deletingTheme?.name}”? Any shelf using it will fall back to the default look. This cannot be undone.`"
      :loading="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>

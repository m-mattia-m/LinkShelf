<script setup lang="ts">
import type { Theme, ThemeBase } from '~~/api'
import { useThemeStore } from '~/stores/theme'

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  theme?: Theme
  initial?: ThemeBase
  // Suppresses the dialog's own default trigger button - for a usage that's
  // opened purely programmatically (e.g. the import flow, which only sets
  // `initial` once a file is picked, so `!initial` alone can't tell "the
  // import dialog" apart from "the create dialog" while no file is chosen
  // yet).
  hideTrigger?: boolean
}>(), {
  mode: 'create',
  hideTrigger: false
})

const emit = defineEmits<{
  (e: 'saved', theme: Theme): void
}>()

const open = defineModel<boolean>('open', { default: false })

const themeStore = useThemeStore()

const formModel = ref<ThemeBase>()
const saving = ref(false)

const formRef = ref<{
  validate: () => Promise<boolean>
  setErrors: (errs: { name?: string, message: string }[]) => void
} | null>(null)

async function save(close: () => void) {
  const valid = await formRef.value?.validate()
  if (!valid || !formModel.value) return

  saving.value = true
  try {
    const saved = props.mode === 'edit' && props.theme
      ? await themeStore.update(props.theme.id, formModel.value)
      : await themeStore.create(formModel.value)

    emit('saved', saved)
    close()
  } catch (err) {
    await handleApiError(err, formRef.value)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="mode === 'edit' ? 'Edit theme' : 'New theme'"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton
      v-if="mode === 'create' && !hideTrigger"
      icon="i-lucide-plus"
      label="New theme"
    />

    <template #body>
      <ThemeForm
        ref="formRef"
        :model-value="theme"
        :initial="initial"
        @update:model-value="(value) => (formModel = value)"
      />
    </template>

    <template #footer="{ close }">
      <UButton
        label="Cancel"
        color="neutral"
        variant="outline"
        @click="close"
      />
      <UButton
        label="Save"
        color="neutral"
        :loading="saving"
        @click="save(close)"
      />
    </template>
  </UModal>
</template>

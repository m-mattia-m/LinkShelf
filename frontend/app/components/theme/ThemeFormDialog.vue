<script setup lang="ts">
import type { Theme, ThemeBase } from '~~/api'
import { useThemeStore } from '~/stores/theme'

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  theme?: Theme
  initial?: ThemeBase
  // Hides the default trigger button for programmatic use (e.g. import).
  hideTrigger?: boolean
}>(), {
  mode: 'create',
  hideTrigger: false
})

const emit = defineEmits<{
  (e: 'saved', theme: Theme): void
}>()

const open = defineModel<boolean>('open', { default: false })

const { t } = useI18n()
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
    :title="mode === 'edit' ? t('app.theme.edit') : t('app.theme.new')"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton
      v-if="mode === 'create' && !hideTrigger"
      icon="i-lucide-plus"
      :label="t('app.theme.new')"
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
        :label="t('common.cancel')"
        color="neutral"
        variant="outline"
        @click="close"
      />
      <UButton
        :label="t('common.save')"
        color="neutral"
        :loading="saving"
        @click="save(close)"
      />
    </template>
  </UModal>
</template>

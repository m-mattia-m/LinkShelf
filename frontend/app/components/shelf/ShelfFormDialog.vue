<script setup lang="ts">
import type { Shelf, ShelfBase } from '~~/api'
import { useShelfStore } from '~/stores/shelf'

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  shelf?: Shelf
}>(), {
  mode: 'create'
})

const emit = defineEmits<{
  (e: 'saved', shelf: Shelf): void
}>()

const open = defineModel<boolean>('open', { default: false })

const { t } = useI18n()
const shelfStore = useShelfStore()
const { reached } = useShelfLimit()

const formModel = ref<ShelfBase>()
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
    const saved = props.mode === 'edit' && props.shelf
      ? await shelfStore.update(props.shelf.id, formModel.value)
      : await shelfStore.create(formModel.value)

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
    :title="t('app.shelf.title')"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton
      v-if="mode === 'create'"
      icon="i-lucide-plus"
      :label="t('common.new')"
      :disabled="reached"
    />

    <template #body>
      <ShelfForm
        ref="formRef"
        :model-value="shelf"
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
        :label="t('common.submit')"
        color="neutral"
        :loading="saving"
        @click="save(close)"
      />
    </template>
  </UModal>
</template>

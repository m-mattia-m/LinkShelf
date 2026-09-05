<script setup lang="ts">
const open = defineModel<boolean>('open', { default: false })

withDefaults(defineProps<{
  title: string
  description?: string
  confirmLabel?: string
  cancelLabel?: string
  color?: 'error' | 'primary' | 'neutral'
  loading?: boolean
}>(), {
  confirmLabel: 'Delete',
  cancelLabel: 'Cancel',
  color: 'error',
  loading: false
})

const emit = defineEmits<{
  (e: 'confirm'): void
}>()
</script>

<template>
  <UModal v-model:open="open" :title="title" :ui="{ footer: 'justify-end' }">
    <template #body>
      <p v-if="description" class="text-sm text-muted">{{ description }}</p>
    </template>

    <template #footer="{ close }">
      <UButton :label="cancelLabel" color="neutral" variant="outline" @click="close" />
      <UButton :label="confirmLabel" :color="color" :loading="loading" @click="emit('confirm')" />
    </template>
  </UModal>
</template>

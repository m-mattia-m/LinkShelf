<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import * as v from 'valibot'
import type { Link } from '~~/api'
import { useLinkStore } from '~/stores/link'

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  sectionId: string
  link?: Link
}>(), {
  mode: 'create'
})

const emit = defineEmits<{
  (e: 'saved'): void
}>()

const open = defineModel<boolean>('open', { default: false })

const linkStore = useLinkStore()
const saving = ref(false)

const form = reactive({
  title: props.link?.title ?? '',
  link: props.link?.link ?? '',
  icon: props.link?.icon ?? '',
  color: props.link?.color ?? ''
})

watch(open, (isOpen) => {
  if (!isOpen) return
  form.title = props.link?.title ?? ''
  form.link = props.link?.link ?? ''
  form.icon = props.link?.icon ?? ''
  form.color = props.link?.color ?? ''
})

const schema = v.object({
  title: v.pipe(v.string('Title is required'), v.nonEmpty('Title is required')),
  link: v.pipe(v.string('URL is required'), v.nonEmpty('URL is required')),
  icon: v.string(),
  // Empty is valid too - it means "no color set", so the shelf's theme (or
  // the default look) decides how the link renders instead.
  color: v.pipe(
    v.string(),
    v.check(
      (value) => value === '' || /^#[0-9a-fA-F]{6}$/.test(value),
      'Must be a hex color, e.g. #588157'
    )
  )
})

const formRef = ref<{
  validate: () => Promise<unknown>
  setErrors: (errs: { name?: string, message: string }[]) => void
} | null>(null)

async function save(close: () => void) {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    const linkBase = {
      title: form.title,
      link: form.link,
      icon: form.icon,
      color: form.color,
      sectionId: props.sectionId
    }

    if (props.mode === 'edit' && props.link) {
      await linkStore.update(props.link.id, linkBase)
    } else {
      await linkStore.create(linkBase)
    }

    emit('saved')
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
    title="Link"
    :ui="{ footer: 'justify-end' }"
  >
    <template #body>
      <UForm ref="formRef" :schema="schema" :state="form" class="flex flex-col gap-4">
        <UFormField label="Title" name="title" required>
          <UInput v-model="form.title" class="w-full" />
        </UFormField>

        <UFormField label="URL" name="link" required>
          <UInput v-model="form.link" class="w-full" placeholder="https://example.com" />
        </UFormField>

        <UFormField label="Icon" name="icon" help="Optional - leave empty for no icon.">
          <IconPicker v-model="form.icon" placeholder="i-lucide-link" />
        </UFormField>

        <UFormField label="Color" name="color" help="Optional - leave unset to use the shelf's theme (or the default look).">
          <div v-if="form.color" class="flex items-center gap-2">
            <UColorPicker v-model="form.color" format="hex" />
            <UInput v-model="form.color" class="w-full" placeholder="#000000" />
            <UButton icon="i-lucide-x" size="sm" color="neutral" variant="ghost" aria-label="Clear color" @click="form.color = ''" />
          </div>
          <UButton v-else label="Set a color" icon="i-lucide-palette" color="neutral" variant="outline" @click="form.color = '#000000'" />
        </UFormField>
      </UForm>
    </template>

    <template #footer="{ close }">
      <UButton label="Cancel" color="neutral" variant="outline" @click="close" />
      <UButton label="Submit" color="neutral" :loading="saving" @click="save(close)" />
    </template>
  </UModal>
</template>

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
  icon: props.link?.icon ?? 'i-lucide-link',
  color: props.link?.color ?? '#000000'
})

watch(open, (isOpen) => {
  if (!isOpen) return
  form.title = props.link?.title ?? ''
  form.link = props.link?.link ?? ''
  form.icon = props.link?.icon ?? 'i-lucide-link'
  form.color = props.link?.color ?? '#000000'
})

const schema = v.object({
  title: v.pipe(v.string(), v.nonEmpty('Required')),
  link: v.pipe(v.string(), v.nonEmpty('Required')),
  icon: v.pipe(v.string(), v.nonEmpty('Required')),
  color: v.pipe(v.string(), v.regex(/^#[0-9a-fA-F]{6}$/, 'Must be a hex color, e.g. #588157'))
})

const formRef = ref<{ setErrors: (errs: { name?: string, message: string }[]) => void } | null>(null)

async function save(close: () => void) {
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
        <UFormField label="Title" name="title">
          <UInput v-model="form.title" class="w-full" />
        </UFormField>

        <UFormField label="URL" name="link">
          <UInput v-model="form.link" class="w-full" placeholder="https://example.com" />
        </UFormField>

        <UFormField label="Icon" name="icon" help="An iconify icon name, e.g. i-lucide-link">
          <div class="flex items-center gap-2">
            <UIcon :name="form.icon || 'i-lucide-image'" class="size-5 shrink-0 text-muted" />
            <UInput v-model="form.icon" class="w-full" placeholder="i-lucide-link" />
          </div>
        </UFormField>

        <UFormField label="Color" name="color">
          <div class="flex items-center gap-2">
            <UColorPicker v-model="form.color" format="hex" />
            <UInput v-model="form.color" class="w-full" placeholder="#000000" />
          </div>
        </UFormField>
      </UForm>
    </template>

    <template #footer="{ close }">
      <UButton label="Cancel" color="neutral" variant="outline" @click="close" />
      <UButton label="Submit" color="neutral" :loading="saving" @click="save(close)" />
    </template>
  </UModal>
</template>

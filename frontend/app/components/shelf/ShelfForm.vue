<script setup lang="ts">
import { reactive, watch, ref, onMounted } from 'vue'
import type { Shelf, ShelfBase } from '~~/api'
import type { FormError } from '@nuxt/ui'
import * as v from 'valibot'

const props = defineProps<{
  modelValue?: Shelf
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ShelfBase): void
}>()

const { ensureUserId } = useCurrentUser()

const tabItems = [
  {
    label: 'Path',
    icon: 'i-lucide-link',
    slot: 'path'
  },
  {
    label: 'Domain',
    icon: 'i-lucide-globe',
    slot: 'domain'
  }
]

const form = reactive<ShelfBase>({
  title: props.modelValue?.title ?? '',
  description: props.modelValue?.description ?? '',
  domain: props.modelValue?.domain ?? '',
  path: props.modelValue?.path ?? '',
  icon: props.modelValue?.icon ?? '',
  theme: props.modelValue?.theme ?? '',
  userId: props.modelValue?.userId ?? ''
})

onMounted(async () => {
  if (!form.userId) {
    form.userId = await ensureUserId()
  }
})

const schema = v.pipe(
  v.object({
    title: v.pipe(v.string(), v.nonEmpty('Required')),
    description: v.string(),
    domain: v.pipe(
      v.string(),
      v.check(
        (value) =>
          value === '' ||
          /^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/.test(value),
        'Please enter a valid domain (e.g. example.com)'
      )
    ),
    path: v.pipe(
      v.string(),
      v.regex(
        /^[a-zA-Z0-9-]*$/,
        'Path may only contain letters, numbers, and hyphens'
      )
    ),
    icon: v.string()
  }),
  v.forward(
    v.check(
      (data) => data.domain.trim() !== '' || data.path.trim() !== '',
      'Either domain or path must be provided'
    ),
    ['domain']
  ),
  v.forward(
    v.check(
      (data) => data.domain.trim() !== '' || data.path.trim() !== '',
      'Either domain or path must be provided'
    ),
    ['path']
  )
)

/**
 * UForm ref
 */
const formRef = ref<any>()

/**
 * Expose validate() ONLY
 */
async function validate(): Promise<boolean> {
  try {
    await formRef.value.validate()
    return true
  } catch {
    return false
  }
}

function setErrors(errs: FormError[]) {
  formRef.value?.setErrors(errs)
}

defineExpose({ validate, setErrors })

/**
 * Sync parent → form
 */
watch(
  () => props.modelValue,
  (newShelf) => {
    if (!newShelf) return
    Object.assign(form, {
      title: newShelf.title,
      description: newShelf.description,
      domain: newShelf.domain,
      path: newShelf.path,
      icon: newShelf.icon,
      theme: newShelf.theme,
      userId: newShelf.userId
    })
  },
  { immediate: true }
)

/**
 * Sync form → parent
 */
watch(
  form,
  () => emit('update:modelValue', { ...form }),
  { deep: true }
)
</script>


<template>
  <UForm
    ref="formRef"
    :schema="schema"
    :state="form"
  >
    <UFormField label="Title" name="title" required>
      <UInput v-model="form.title" class="w-full" />
    </UFormField>

    <UFormField label="Description" name="description" class="pt-4">
      <UTextarea v-model="form.description" class="w-full" />
    </UFormField>

    <UFormField label="Icon" name="icon" help="An iconify icon name, e.g. i-lucide-book-open" class="pt-4">
      <div class="flex items-center gap-2">
        <UIcon :name="form.icon || 'i-lucide-image'" class="size-5 shrink-0 text-muted" />
        <UInput v-model="form.icon" class="w-full" placeholder="i-lucide-book-open" />
      </div>
    </UFormField>

    <UTabs :items="tabItems" class="pt-4 w-full">
      <template #domain>
        <UFormField label="Domain" name="domain" :help="'https://' + form.domain">
          <UInput v-model="form.domain" class="w-full" />
        </UFormField>
      </template>

      <template #path>
        <UFormField label="Path" name="path" :help="'https://linkshelf.com/' + form.path">
          <UInput v-model="form.path" class="w-full" />
        </UFormField>
      </template>
    </UTabs>
  </UForm>
</template>

<script setup lang="ts">
import { reactive, watch, ref, computed, onMounted } from 'vue'
import type { SettingPageBody, Shelf, ShelfBase } from '~~/api'
import type { FormError, SelectItem } from '@nuxt/ui'
import * as v from 'valibot'
import { useThemeStore } from '~/stores/theme'

const props = defineProps<{
  modelValue?: Shelf
}>()

const { user: currentUser } = useCurrentUser()
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>
const userBasedPaths = computed(() => websiteSettings.value?.userBasedPaths ?? false)
const origin = useRequestURL().origin

const themeStore = useThemeStore()
onMounted(() => {
  if (!themeStore.loaded) themeStore.fetch()
})

// Reka UI's Select reserves an empty string to mean "cleared" and throws if
// an item actually uses it as a value, so "no theme" needs a real sentinel
// value here - selectedThemeId below converts it back to "" for the form.
const NO_THEME_VALUE = '__none__'

const themeItems = computed<SelectItem[]>(() => [
  { label: 'No theme (default look)', value: NO_THEME_VALUE },
  ...(themeStore.instance.length
    ? [{ type: 'label' as const, label: 'Instance themes' }, ...themeStore.instance.map(t => ({ label: t.name, value: t.id }))]
    : []),
  ...(themeStore.mine.length
    ? [{ type: 'label' as const, label: 'Your themes' }, ...themeStore.mine.map(t => ({ label: t.name, value: t.id }))]
    : [])
])

const emit = defineEmits<{
  (e: 'update:modelValue', value: ShelfBase): void
}>()

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

const form = reactive({
  title: props.modelValue?.title ?? '',
  description: props.modelValue?.description ?? '',
  domain: props.modelValue?.domain ?? '',
  path: props.modelValue?.path ?? '',
  icon: props.modelValue?.icon ?? '',
  themeId: props.modelValue?.themeId ?? ''
})

const schema = v.pipe(
  v.object({
    title: v.pipe(v.string(), v.nonEmpty('Required')),
    description: v.string(),
    domain: v.pipe(
      v.string(),
      v.check(
        value =>
          value === ''
          || /^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/.test(value),
        'Please enter a valid domain (e.g. example.com)'
      )
    ),
    path: v.pipe(
      v.string(),
      v.regex(
        /^[a-zA-Z0-9-]*$/,
        'Path may only contain letters, numbers, and hyphens'
      ),
      // Behind a username nothing is off limits. Top-level, the words the app
      // itself answers can't be used - except on a shelf that already has
      // one, so its other fields stay editable (the backend does the same).
      v.check(
        value => userBasedPaths.value || value === props.modelValue?.path || !isRouteReservedPath(value),
        'This path is used by a page of this app. Please choose another one'
      )
    ),
    icon: v.string()
  }),
  v.forward(
    v.check(
      data => data.domain.trim() !== '' || data.path.trim() !== '',
      'Either domain or path must be provided'
    ),
    ['domain']
  ),
  v.forward(
    v.check(
      data => data.domain.trim() !== '' || data.path.trim() !== '',
      'Either domain or path must be provided'
    ),
    ['path']
  )
)

// The URL the shelf will get: /<username>/<path> with user-based paths, else
// /<path>. A shelf being edited keeps its owner's username - an admin may be
// editing someone else's shelf.
const ownerUsername = computed(() => props.modelValue?.username || currentUser.value?.username || '')
const pathHelp = computed(() => userBasedPaths.value
  ? `${origin}/${ownerUsername.value || '<username>'}/${form.path}`
  : `${origin}/${form.path}`)

const selectedThemeId = computed({
  get: () => form.themeId || NO_THEME_VALUE,
  set: (value: string) => {
    form.themeId = value === NO_THEME_VALUE ? '' : value
  }
})

/**
 * UForm ref
 */
const formRef = ref<{ validate: () => Promise<unknown>, setErrors: (errs: FormError[]) => void }>()

/**
 * Expose validate() ONLY
 */
async function validate(): Promise<boolean> {
  try {
    await formRef.value!.validate()
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
      // A missing theme's id no longer matches any picker option, which
      // would otherwise show the raw stale id as the selection - the alert
      // above already says it's gone, so just clear it instead.
      themeId: newShelf.themeMissing ? '' : newShelf.themeId
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
    <UFormField
      label="Title"
      name="title"
      required
    >
      <UInput
        v-model="form.title"
        class="w-full"
      />
    </UFormField>

    <UFormField
      label="Description"
      name="description"
      class="pt-4"
    >
      <UTextarea
        v-model="form.description"
        class="w-full"
      />
    </UFormField>

    <UFormField
      label="Icon"
      name="icon"
      class="pt-4"
      help="Optional - leave empty for no icon."
    >
      <IconPicker
        v-model="form.icon"
        placeholder="i-lucide-book-open"
      />
    </UFormField>

    <UAlert
      v-if="modelValue?.themeMissing"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Theme unavailable"
      description="The theme this shelf was using no longer exists. Pick another one below."
      class="mt-4"
    />

    <UFormField
      label="Theme"
      name="themeId"
      class="pt-4"
      help="Only affects this shelf's public page, never the app."
    >
      <USelect
        v-model="selectedThemeId"
        :items="themeItems"
        value-key="value"
        class="w-full"
      />
    </UFormField>

    <UTabs
      :items="tabItems"
      class="pt-4 w-full"
    >
      <template #domain>
        <UFormField
          label="Domain"
          name="domain"
          :help="'https://' + form.domain"
        >
          <UInput
            v-model="form.domain"
            class="w-full"
          />
        </UFormField>
      </template>

      <template #path>
        <UFormField
          label="Path"
          name="path"
          :help="pathHelp"
        >
          <UInput
            v-model="form.path"
            class="w-full"
          />
        </UFormField>
      </template>
    </UTabs>
  </UForm>
</template>

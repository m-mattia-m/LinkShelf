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

// A shelf is reached through a path or through a domain of its own, never
// both. The tab that is open when the form is saved decides which one is kept
// and the other is sent empty, so the backend's "exactly one" rule can never
// surprise anyone who filled in both.
type Mode = 'path' | 'domain'

const tabItems = [
  {
    label: 'Path',
    icon: 'i-lucide-link',
    slot: 'path',
    value: 'path'
  },
  {
    label: 'Domain',
    icon: 'i-lucide-globe',
    slot: 'domain',
    value: 'domain'
  }
]

// A shelf that only has a domain opens on the Domain tab. One that has both
// (created before a shelf had to choose) opens on Path, like everywhere else.
function modeOf(shelf?: Pick<Shelf, 'path' | 'domain'>): Mode {
  return shelf?.domain && !shelf?.path ? 'domain' : 'path'
}

const mode = ref<Mode>(modeOf(props.modelValue))
const hadBoth = computed(() => Boolean(props.modelValue?.path && props.modelValue?.domain))

// The host this app is being used on: a shelf can't take over the instance
// itself. The backend also rejects its configured frontend host.
const ownHost = normalizeShelfDomain(useRequestURL().host)

const form = reactive({
  title: props.modelValue?.title ?? '',
  description: props.modelValue?.description ?? '',
  domain: props.modelValue?.domain ?? '',
  path: props.modelValue?.path ?? '',
  icon: props.modelValue?.icon ?? '',
  themeId: props.modelValue?.themeId ?? ''
})

// Each of path and domain is only checked while its tab is the one in use - a
// half-typed value on the other tab is thrown away on save anyway.
const schema = v.object({
  title: v.pipe(v.string(), v.nonEmpty('Required')),
  description: v.string(),
  domain: v.pipe(
    v.string(),
    v.check(
      value => mode.value !== 'domain' || normalizeShelfDomain(value) !== '',
      'Please enter a domain (e.g. profile.example.com)'
    ),
    v.check(
      value => mode.value !== 'domain' || validateShelfDomain(normalizeShelfDomain(value)) === null,
      issue => validateShelfDomain(normalizeShelfDomain(String(issue.input))) ?? 'Please enter a valid domain'
    ),
    v.check(
      value => mode.value !== 'domain' || normalizeShelfDomain(value) !== ownHost,
      'This is the address of this LinkShelf instance itself. Please choose another domain'
    )
  ),
  path: v.pipe(
    v.string(),
    v.check(
      value => mode.value !== 'path' || value.trim() !== '',
      'Please enter a path'
    ),
    v.check(
      value => mode.value !== 'path' || /^[a-zA-Z0-9-]*$/.test(value),
      'Path may only contain letters, numbers, and hyphens'
    ),
    // Behind a username nothing is off limits. Top-level, the words the app
    // itself answers can't be used - except on a shelf that already has
    // one, so its other fields stay editable (the backend does the same).
    v.check(
      value => mode.value !== 'path' || userBasedPaths.value || value === props.modelValue?.path || !isRouteReservedPath(value),
      'This path is used by a page of this app. Please choose another one'
    )
  ),
  icon: v.string()
})

// The URL the shelf will get: /<username>/<path> with user-based paths, else
// /<path>. A shelf being edited keeps its owner's username - an admin may be
// editing someone else's shelf.
const ownerUsername = computed(() => props.modelValue?.username || currentUser.value?.username || '')
const pathHelp = computed(() => userBasedPaths.value
  ? `${origin}/${ownerUsername.value || '<username>'}/${form.path}`
  : `${origin}/${form.path}`)

const domainHelp = computed(() => {
  const domain = normalizeShelfDomain(form.domain)
  const shown = domain ? `https://${domain}` : 'https://<domain>'
  return `${shown} - point the domain's DNS at this LinkShelf instance and route it to the frontend.`
})

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
    mode.value = modeOf(newShelf)
  },
  { immediate: true }
)

/**
 * Sync form → parent
 */
watch(
  [form, mode],
  () => emit('update:modelValue', {
    ...form,
    path: mode.value === 'path' ? form.path.trim() : '',
    domain: mode.value === 'domain' ? normalizeShelfDomain(form.domain) : ''
  }),
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

    <UAlert
      v-if="hadBoth"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Path and domain"
      :description="`This shelf has both a path and a domain, and a shelf can only have one of them. Saving keeps the ${mode} and clears the other.`"
      class="mt-4"
    />

    <p class="pt-4 text-sm text-muted">
      A shelf is reached through either a path on this site or a domain of its own. The tab that is open when you save
      is used, the other one is cleared.
    </p>

    <UTabs
      v-model="mode"
      :items="tabItems"
      class="pt-2 w-full"
    >
      <template #domain>
        <UFormField
          label="Domain"
          name="domain"
          :help="domainHelp"
        >
          <UInput
            v-model="form.domain"
            class="w-full"
            placeholder="profile.example.com"
            @blur="form.domain = normalizeShelfDomain(form.domain)"
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

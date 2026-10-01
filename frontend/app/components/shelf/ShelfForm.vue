<script setup lang="ts">
import { reactive, watch, ref, computed, onMounted } from 'vue'
import type { SettingPageBody, Shelf, ShelfBase } from '~~/api'
import type { FormError, SelectItem } from '@nuxt/ui'
import * as v from 'valibot'
import { useThemeStore } from '~/stores/theme'

const props = defineProps<{
  modelValue?: Shelf
}>()

const { t } = useI18n()
const { user: currentUser } = useCurrentUser()
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>
const userBasedPaths = computed(() => websiteSettings.value?.userBasedPaths ?? false)
const origin = useRequestURL().origin

const themeStore = useThemeStore()
onMounted(() => {
  if (!themeStore.loaded) themeStore.fetch()
})

// Reka UI's Select reserves "" for cleared, so "no theme" needs a sentinel.
const NO_THEME_VALUE = '__none__'

const themeItems = computed<SelectItem[]>(() => [
  { label: t('app.shelf.form.noTheme'), value: NO_THEME_VALUE },
  ...(themeStore.instance.length
    ? [{ type: 'label' as const, label: t('app.shelf.form.instanceThemes') }, ...themeStore.instance.map(t => ({ label: t.name, value: t.id }))]
    : []),
  ...(themeStore.mine.length
    ? [{ type: 'label' as const, label: t('app.shelf.form.yourThemes') }, ...themeStore.mine.map(t => ({ label: t.name, value: t.id }))]
    : [])
])

const emit = defineEmits<{
  (e: 'update:modelValue', value: ShelfBase): void
}>()

// A shelf uses a path or a domain; the open tab decides which is saved.
type Mode = 'path' | 'domain'

const tabItems = computed(() => [
  {
    label: t('app.shelf.form.path'),
    icon: 'i-lucide-link',
    slot: 'path',
    value: 'path'
  },
  {
    label: t('app.shelf.form.domain'),
    icon: 'i-lucide-globe',
    slot: 'domain',
    value: 'domain'
  }
])

// Domain-only shelves open on the Domain tab, all others on Path.
function modeOf(shelf?: Pick<Shelf, 'path' | 'domain'>): Mode {
  return shelf?.domain && !shelf?.path ? 'domain' : 'path'
}

const mode = ref<Mode>(modeOf(props.modelValue))
const hadBoth = computed(() => Boolean(props.modelValue?.path && props.modelValue?.domain))

// The current host, which a shelf can't claim.
const ownHost = normalizeShelfDomain(useRequestURL().host)

const form = reactive({
  title: props.modelValue?.title ?? '',
  description: props.modelValue?.description ?? '',
  domain: props.modelValue?.domain ?? '',
  path: props.modelValue?.path ?? '',
  icon: props.modelValue?.icon ?? '',
  themeId: props.modelValue?.themeId ?? '',
  noIndex: props.modelValue?.noIndex ?? false,
  // New shelves default to showing the default footer.
  footerEnabled: props.modelValue?.footerEnabled ?? true,
  footerCustomText: props.modelValue?.footerCustomText ?? ''
})

// Only the active tab's field is validated.
const schema = computed(() => v.object({
  title: v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  description: v.string(),
  domain: v.pipe(
    v.string(),
    v.check(
      value => mode.value !== 'domain' || normalizeShelfDomain(value) !== '',
      t('validation.domain.missing')
    ),
    v.check(
      value => mode.value !== 'domain' || validateShelfDomain(normalizeShelfDomain(value)) === null,
      issue => t(validateShelfDomain(normalizeShelfDomain(String(issue.input))) ?? 'validation.domain.invalid')
    ),
    v.check(
      value => mode.value !== 'domain' || normalizeShelfDomain(value) !== ownHost,
      t('validation.domain.ownHost')
    )
  ),
  path: v.pipe(
    v.string(),
    v.check(
      value => mode.value !== 'path' || value.trim() !== '',
      t('validation.path.missing')
    ),
    v.check(
      value => mode.value !== 'path' || /^[a-zA-Z0-9-]*$/.test(value),
      t('validation.path.invalid')
    ),
    // Without user-based paths, reserved words are blocked unless the shelf already has one.
    v.check(
      value => mode.value !== 'path' || userBasedPaths.value || value === props.modelValue?.path || !isRouteReservedPath(value),
      t('validation.path.reserved')
    )
  ),
  icon: v.string(),
  footerCustomText: v.pipe(v.string(), v.maxLength(500, t('validation.maxLength', { max: 500 })))
}))

// The shelf's URL; keeps the owner's username when an admin edits it.
const ownerUsername = computed(() => props.modelValue?.username || currentUser.value?.username || '')
const pathHelp = computed(() => userBasedPaths.value
  ? `${origin}/${ownerUsername.value || '<username>'}/${form.path}`
  : `${origin}/${form.path}`)

const domainHelp = computed(() => {
  const domain = normalizeShelfDomain(form.domain)
  const shown = domain ? `https://${domain}` : 'https://<domain>'
  return t('app.shelf.form.domainHelp', { url: shown })
})

const selectedThemeId = computed({
  get: () => form.themeId || NO_THEME_VALUE,
  set: (value: string) => {
    form.themeId = value === NO_THEME_VALUE ? '' : value
  }
})

const formRef = ref<{ validate: () => Promise<unknown>, setErrors: (errs: FormError[]) => void }>()

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

// Sync parent → form
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
      // Clear a missing theme instead of showing its stale id.
      themeId: newShelf.themeMissing ? '' : newShelf.themeId,
      // Fall back to defaults for shelves missing these fields.
      noIndex: newShelf.noIndex ?? false,
      footerEnabled: newShelf.footerEnabled ?? true,
      footerCustomText: newShelf.footerCustomText ?? ''
    })
    mode.value = modeOf(newShelf)
  },
  { immediate: true }
)

// Sync form → parent
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
      :label="t('app.shelf.form.title')"
      name="title"
      required
    >
      <UInput
        v-model="form.title"
        class="w-full"
      />
    </UFormField>

    <UFormField
      :label="t('app.shelf.form.description')"
      name="description"
      class="pt-4"
    >
      <UTextarea
        v-model="form.description"
        class="w-full"
      />
    </UFormField>

    <UFormField
      :label="t('app.shelf.form.icon')"
      name="icon"
      class="pt-4"
      :help="t('app.shelf.form.iconHelp')"
    >
      <IconPicker
        v-model="form.icon"
        placeholder="i-lucide-book-open"
      />
    </UFormField>

    <UCheckbox
      v-model="form.noIndex"
      name="noIndex"
      class="pt-4"
      :label="t('app.shelf.form.noIndex')"
      :description="t('app.shelf.form.noIndexHelp')"
    />

    <UCheckbox
      v-model="form.footerEnabled"
      name="footerEnabled"
      class="pt-4"
      :label="t('app.shelf.form.showFooter')"
      :description="t('app.shelf.form.showFooterHelp')"
    />

    <UFormField
      v-if="form.footerEnabled"
      :label="t('app.shelf.form.footerText')"
      name="footerCustomText"
      class="pt-4"
      :help="t('app.shelf.form.footerTextHelp')"
    >
      <UTextarea
        v-model="form.footerCustomText"
        class="w-full"
        :rows="2"
        :placeholder="t('linkpage.poweredBy')"
      />
    </UFormField>

    <UAlert
      v-if="modelValue?.themeMissing"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      :title="t('app.shelf.form.themeMissing.title')"
      :description="t('app.shelf.form.themeMissing.description')"
      class="mt-4"
    />

    <UFormField
      :label="t('app.shelf.form.theme')"
      name="themeId"
      class="pt-4"
      :help="t('app.shelf.form.themeHelp')"
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
      :title="t('app.shelf.form.hadBoth.title')"
      :description="t(mode === 'path' ? 'app.shelf.form.hadBoth.keepsPath' : 'app.shelf.form.hadBoth.keepsDomain')"
      class="mt-4"
    />

    <p class="pt-4 text-sm text-muted">
      {{ t('app.shelf.form.pathOrDomain') }}
    </p>

    <UTabs
      v-model="mode"
      :items="tabItems"
      class="pt-2 w-full"
    >
      <template #domain>
        <UFormField
          :label="t('app.shelf.form.domain')"
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
          :label="t('app.shelf.form.path')"
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

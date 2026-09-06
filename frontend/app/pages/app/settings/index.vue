<script setup lang="ts">
import { useSettingStore } from '~/stores/setting'

definePageMeta({
  layout: 'app',
  middleware: 'admin'
})

const { t, locales } = useI18n()

const settingStore = useSettingStore()
const loading = ref(true)
const saving = ref(false)

const languageCode = ref('en')
const availableLocales = computed(() => locales.value.map((l) => ({ label: l.name ?? l.code, value: l.code })))

const loginOptionChoices = ['Local', 'Zitadel', 'Microsoft', 'Google']

const form = reactive({
  about: '',
  aboutShow: false,
  contact: '',
  contactShow: false,
  imprint: '',
  imprintShow: false,
  termsOfUse: '',
  termsOfUseShow: false,
  privacyPolicy: '',
  privacyPolicyShow: false,
  redirectToDashboard: false,
  loginOptions: [] as string[]
})

function applyPageToForm() {
  const page = settingStore.page
  if (!page) return
  form.about = page.about
  form.aboutShow = page.aboutShow
  form.contact = page.contact
  form.contactShow = page.contactShow
  form.imprint = page.imprint
  form.imprintShow = page.imprintShow
  form.termsOfUse = page.termsOfUse
  form.termsOfUseShow = page.termsOfUseShow
  form.privacyPolicy = page.privacyPolicy
  form.privacyPolicyShow = page.privacyPolicyShow
  form.redirectToDashboard = page.redirectToDashboard
  form.loginOptions = page.loginOptions ?? []
}

async function loadLanguage(code: string) {
  loading.value = true
  try {
    await settingStore.fetch(code)
    applyPageToForm()
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => loadLanguage(languageCode.value))
watch(languageCode, (code) => loadLanguage(code))

async function save() {
  saving.value = true
  try {
    const failures = await settingStore.updateMany(languageCode.value, [
      { key: 'about', value: form.about },
      { key: 'about_show', value: String(form.aboutShow) },
      { key: 'contact', value: form.contact },
      { key: 'contact_show', value: String(form.contactShow) },
      { key: 'imprint', value: form.imprint },
      { key: 'imprint_show', value: String(form.imprintShow) },
      { key: 'terms_of_use', value: form.termsOfUse },
      { key: 'terms_of_use_show', value: String(form.termsOfUseShow) },
      { key: 'privacy_policy', value: form.privacyPolicy },
      { key: 'privacy_policy_show', value: String(form.privacyPolicyShow) },
      { key: 'redirect_to_dashboard', value: String(form.redirectToDashboard) },
      { key: 'login_options', value: JSON.stringify(form.loginOptions) }
    ])

    applyPageToForm()

    const toast = useToast()
    if (failures.length === 0) {
      toast.add({ title: t('app.settings.saveSuccess'), color: 'success' })
    } else {
      toast.add({
        title: t('app.settings.savePartialFailure'),
        description: failures.map((f) => `${f.key}: ${f.message}`).join(', '),
        color: 'error'
      })
    }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="flex justify-between items-center pb-4">
    <h1 class="text-2xl text-highlighted">{{ t('app.settings.title') }}</h1>

    <USelectMenu
      v-model="languageCode"
      :items="availableLocales"
      value-key="value"
      class="w-48"
    />
  </div>

  <div v-if="loading" class="space-y-4">
    <USkeleton v-for="i in 4" :key="i" class="h-24 w-full" />
  </div>

  <div v-else class="flex flex-col gap-6">
    <UFormField :label="t('app.settings.about.label')">
      <UTextarea v-model="form.about" class="w-full" :rows="6" />
      <template #hint>
        <USwitch v-model="form.aboutShow" :label="t('app.settings.showOnSite')" />
      </template>
    </UFormField>

    <UFormField :label="t('app.settings.contact.label')">
      <UTextarea v-model="form.contact" class="w-full" :rows="6" />
      <template #hint>
        <USwitch v-model="form.contactShow" :label="t('app.settings.showOnSite')" />
      </template>
    </UFormField>

    <UFormField :label="t('app.settings.imprint.label')">
      <UTextarea v-model="form.imprint" class="w-full" :rows="6" />
      <template #hint>
        <USwitch v-model="form.imprintShow" :label="t('app.settings.showOnSite')" />
      </template>
    </UFormField>

    <UFormField :label="t('app.settings.termsOfUse.label')">
      <UTextarea v-model="form.termsOfUse" class="w-full" :rows="6" />
      <template #hint>
        <USwitch v-model="form.termsOfUseShow" :label="t('app.settings.showOnSite')" />
      </template>
    </UFormField>

    <UFormField :label="t('app.settings.privacyPolicy.label')">
      <UTextarea v-model="form.privacyPolicy" class="w-full" :rows="6" />
      <template #hint>
        <USwitch v-model="form.privacyPolicyShow" :label="t('app.settings.showOnSite')" />
      </template>
    </UFormField>

    <UFormField :label="t('app.settings.loginOptions.label')" :help="t('app.settings.loginOptions.help')">
      <USelectMenu v-model="form.loginOptions" :items="loginOptionChoices" multiple class="w-full" />
    </UFormField>

    <UFormField :label="t('app.settings.redirectToDashboard.label')" :help="t('app.settings.redirectToDashboard.help')">
      <USwitch v-model="form.redirectToDashboard" />
    </UFormField>

    <div>
      <UButton :label="t('app.settings.save')" color="neutral" :loading="saving" @click="save" />
    </div>
  </div>
</template>

<script setup lang="ts">
import * as v from 'valibot'
import type { FormSubmitEvent } from '@nuxt/ui'
import UserPasswordDialog from '~/components/user/UserPasswordDialog.vue'

definePageMeta({
  layout: 'app'
})

const { t } = useI18n()
const { user, ensureUser } = useCurrentUser()
const authStore = useAuthStore()

const loading = ref(true)
const saving = ref(false)
const passwordOpen = ref(false)

const form = reactive({
  firstName: '',
  lastName: '',
  email: ''
})

onMounted(async () => {
  try {
    await ensureUser()
    if (user.value) {
      form.firstName = user.value.firstName
      form.lastName = user.value.lastName
      form.email = user.value.email
    }
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
})

const schema = v.object({
  firstName: v.pipe(v.string('First name is required'), v.nonEmpty('First name is required')),
  lastName: v.pipe(v.string('Last name is required'), v.nonEmpty('Last name is required')),
  email: v.pipe(
    v.string('Email is required'),
    v.nonEmpty('Email is required'),
    v.email('Please enter a valid email')
  )
})

type Schema = v.InferOutput<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  if (!user.value) return
  saving.value = true
  try {
    const api = useApi()
    const updated = await api.user.putUpdateUser({
      userId: user.value.id,
      userBase: {
        firstName: payload.data.firstName,
        lastName: payload.data.lastName,
        email: payload.data.email
      }
    })
    authStore.user = updated
    const toast = useToast()
    toast.add({ title: t('app.profile.saveSuccess'), color: 'success' })
  } catch (err) {
    await handleApiError(err)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl text-highlighted pb-4">{{ t('app.profile.title') }}</h1>

    <div v-if="loading" class="flex flex-col gap-4 max-w-md">
      <USkeleton v-for="i in 3" :key="i" class="h-10 w-full" />
    </div>

    <UForm v-else :schema="schema" :state="form" class="flex flex-col gap-4 max-w-md" @submit="onSubmit">
      <UFormField :label="t('app.profile.firstName')" name="firstName" required>
        <UInput v-model="form.firstName" class="w-full" />
      </UFormField>

      <UFormField :label="t('app.profile.lastName')" name="lastName" required>
        <UInput v-model="form.lastName" class="w-full" />
      </UFormField>

      <UFormField :label="t('app.profile.email')" name="email" required>
        <UInput v-model="form.email" type="email" class="w-full" />
      </UFormField>

      <UFormField :label="t('app.profile.role')">
        <UBadge :label="user?.role" color="neutral" variant="subtle" />
      </UFormField>

      <div class="flex gap-2 pt-2">
        <UButton type="submit" :label="t('app.profile.save')" color="neutral" :loading="saving" />
        <UButton
          :label="t('app.profile.changePassword')"
          color="neutral"
          variant="outline"
          @click="passwordOpen = true"
        />
      </div>
    </UForm>

    <UserPasswordDialog v-model:open="passwordOpen" :user="user ?? undefined" />
  </div>
</template>

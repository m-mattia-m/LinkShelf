<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import * as v from 'valibot'
import type { User } from '~~/api'
import { useUserStore } from '~/stores/user'

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  user?: User
}>(), {
  mode: 'create'
})

const emit = defineEmits<{
  (e: 'saved'): void
}>()

const open = defineModel<boolean>('open', { default: false })

const userStore = useUserStore()
const saving = ref(false)

const form = reactive({
  firstName: props.user?.firstName ?? '',
  lastName: props.user?.lastName ?? '',
  email: props.user?.email ?? '',
  password: ''
})

watch(open, (isOpen) => {
  if (!isOpen) return
  form.firstName = props.user?.firstName ?? ''
  form.lastName = props.user?.lastName ?? ''
  form.email = props.user?.email ?? ''
  form.password = ''
})

const createSchema = v.object({
  firstName: v.pipe(v.string(), v.nonEmpty('Required')),
  lastName: v.pipe(v.string(), v.nonEmpty('Required')),
  email: v.pipe(v.string(), v.nonEmpty('Required'), v.email('Must be a valid email address')),
  password: v.pipe(v.string(), v.nonEmpty('Required'))
})

const editSchema = v.object({
  firstName: v.pipe(v.string(), v.nonEmpty('Required')),
  lastName: v.pipe(v.string(), v.nonEmpty('Required')),
  email: v.pipe(v.string(), v.nonEmpty('Required'), v.email('Must be a valid email address')),
  password: v.string()
})

const schema = computed(() => (props.mode === 'create' ? createSchema : editSchema))

const formRef = ref<{ setErrors: (errs: { name?: string, message: string }[]) => void } | null>(null)

async function save(close: () => void) {
  saving.value = true
  try {
    if (props.mode === 'edit' && props.user) {
      await userStore.update(props.user.id, {
        firstName: form.firstName,
        lastName: form.lastName,
        email: form.email
      })
    } else {
      await userStore.create({
        firstName: form.firstName,
        lastName: form.lastName,
        email: form.email,
        password: form.password
      })
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
    :title="mode === 'edit' ? 'Edit user' : 'New user'"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton v-if="mode === 'create'" icon="i-lucide-plus" label="New" />

    <template #body>
      <UForm ref="formRef" :schema="schema" :state="form" class="flex flex-col gap-4">
        <UFormField label="First name" name="firstName" required>
          <UInput v-model="form.firstName" class="w-full" />
        </UFormField>

        <UFormField label="Last name" name="lastName" required>
          <UInput v-model="form.lastName" class="w-full" />
        </UFormField>

        <UFormField label="Email" name="email" required>
          <UInput v-model="form.email" type="email" class="w-full" />
        </UFormField>

        <UFormField v-if="mode === 'create'" label="Password" name="password" required>
          <UInput v-model="form.password" type="password" class="w-full" />
        </UFormField>
      </UForm>
    </template>

    <template #footer="{ close }">
      <UButton label="Cancel" color="neutral" variant="outline" @click="close" />
      <UButton label="Submit" color="neutral" :loading="saving" @click="save(close)" />
    </template>
  </UModal>
</template>

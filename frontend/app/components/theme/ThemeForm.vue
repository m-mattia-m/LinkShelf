<script setup lang="ts">
import { reactive, watch, ref } from 'vue'
import type { Theme, ThemeBase } from '~~/api'
import type { FormError } from '@nuxt/ui'
import * as v from 'valibot'

const props = defineProps<{
  modelValue?: Theme
  initial?: ThemeBase
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ThemeBase): void
}>()

const { t } = useI18n()

const form = reactive<ThemeBase>({
  name: props.modelValue?.name ?? props.initial?.name ?? '',
  config: props.modelValue?.config ?? props.initial?.config ?? ''
})

const schema = computed(() => v.object({
  name: v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  config: v.pipe(v.string(), v.nonEmpty(t('validation.required')))
}))

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

watch(
  () => props.modelValue,
  (newTheme) => {
    if (!newTheme) return
    Object.assign(form, { name: newTheme.name, config: newTheme.config })
  }
)

watch(
  () => props.initial,
  (initial) => {
    if (!initial) return
    Object.assign(form, { name: initial.name, config: initial.config })
  }
)

watch(
  form,
  () => emit('update:modelValue', { ...form }),
  { deep: true, immediate: true }
)

const propertyReference = computed(() => [
  ['--shelf-bg', t('app.theme.form.properties.bg')],
  ['--shelf-text', t('app.theme.form.properties.text')],
  ['--shelf-link-bg', t('app.theme.form.properties.linkBg')],
  ['--shelf-link-text', t('app.theme.form.properties.linkText')],
  ['--shelf-link-radius', t('app.theme.form.properties.linkRadius')],
  ['--shelf-font-family', t('app.theme.form.properties.fontFamily')],
  ['--shelf-bg-image', t('app.theme.form.properties.bgImage')]
])
</script>

<template>
  <UForm
    ref="formRef"
    :schema="schema"
    :state="form"
  >
    <UFormField
      :label="t('app.theme.form.name')"
      name="name"
      required
    >
      <UInput
        v-model="form.name"
        class="w-full"
      />
    </UFormField>

    <UFormField
      :label="t('app.theme.form.config')"
      name="config"
      class="pt-4"
      :help="t('app.theme.form.configHelp')"
    >
      <UTextarea
        v-model="form.config"
        :rows="8"
        class="w-full font-mono text-sm"
        placeholder="--shelf-bg: #1c274c;&#10;--shelf-text: #ffffff;"
      />
    </UFormField>

    <div class="pt-3 text-xs text-muted space-y-1">
      <p class="font-medium text-dimmed">
        {{ t('app.theme.form.availableProperties') }}
      </p>
      <p
        v-for="[prop, desc] in propertyReference"
        :key="prop"
      >
        <code class="text-highlighted">{{ prop }}</code> — {{ desc }}
      </p>
    </div>
  </UForm>
</template>

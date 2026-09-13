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

const form = reactive<ThemeBase>({
  name: props.modelValue?.name ?? props.initial?.name ?? '',
  config: props.modelValue?.config ?? props.initial?.config ?? ''
})

const schema = v.object({
  name: v.pipe(v.string(), v.nonEmpty('Required')),
  config: v.pipe(v.string(), v.nonEmpty('Required'))
})

const formRef = ref<any>()

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

const propertyReference = [
  ['--shelf-bg', 'background color or gradient, e.g. #1c274c'],
  ['--shelf-text', 'text color, e.g. #ffffff'],
  ['--shelf-link-bg', 'link button background color'],
  ['--shelf-link-text', 'link button text color'],
  ['--shelf-link-radius', 'link button corner radius, e.g. 12px'],
  ['--shelf-font-family', "font stack, e.g. 'Inter', sans-serif"],
  ['--shelf-bg-image', 'background image URL (https:// or /images/...)']
]
</script>

<template>
  <UForm ref="formRef" :schema="schema" :state="form">
    <UFormField label="Name" name="name" required>
      <UInput v-model="form.name" class="w-full" />
    </UFormField>

    <UFormField label="Config" name="config" class="pt-4" help="One '--property: value;' declaration per line.">
      <UTextarea v-model="form.config" :rows="8" class="w-full font-mono text-sm" placeholder="--shelf-bg: #1c274c;&#10;--shelf-text: #ffffff;" />
    </UFormField>

    <div class="pt-3 text-xs text-muted space-y-1">
      <p class="font-medium text-dimmed">Available properties</p>
      <p v-for="[prop, desc] in propertyReference" :key="prop">
        <code class="text-highlighted">{{ prop }}</code> — {{ desc }}
      </p>
    </div>
  </UForm>
</template>

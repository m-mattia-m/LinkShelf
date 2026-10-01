<script setup lang="ts">
import type { DotType, CornerDotType, CornerSquareType } from 'qr-code-styling'

const { t } = useI18n()

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  url: string
  title?: string
}>()

const shapeOptions = computed<{ label: string, value: DotType }[]>(() => [
  { label: t('app.shelf.qr.shapes.square'), value: 'square' },
  { label: t('app.shelf.qr.shapes.rounded'), value: 'rounded' },
  { label: t('app.shelf.qr.shapes.extraRounded'), value: 'extra-rounded' },
  { label: t('app.shelf.qr.shapes.dots'), value: 'dots' }
])

const color = ref('#000000')
const shape = ref<DotType>('square')
const downloading = ref<'svg' | 'png' | null>(null)

// Map each dot style to a matching corner style.
const cornerStylesByShape: Record<DotType, { square: CornerSquareType, dot: CornerDotType }> = {
  'square': { square: 'square', dot: 'square' },
  'rounded': { square: 'extra-rounded', dot: 'dot' },
  'extra-rounded': { square: 'extra-rounded', dot: 'dot' },
  'dots': { square: 'dot', dot: 'dot' },
  'classy': { square: 'extra-rounded', dot: 'dot' },
  'classy-rounded': { square: 'extra-rounded', dot: 'dot' }
}

const previewEl = ref<HTMLElement>()
// qr-code-styling renders straight to the DOM, so the instance isn't reactive.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let qrCode: any

function buildOptions() {
  const corners = cornerStylesByShape[shape.value]
  return {
    width: 240,
    height: 240,
    margin: 8,
    data: props.url,
    dotsOptions: { color: color.value, type: shape.value },
    cornersSquareOptions: { color: color.value, type: corners.square },
    cornersDotOptions: { color: color.value, type: corners.dot },
    backgroundOptions: { color: '#ffffff' }
  }
}

async function render() {
  if (!open.value || !props.url) return
  await nextTick()
  if (!previewEl.value) return

  if (!qrCode) {
    const { default: QRCodeStyling } = await import('qr-code-styling')
    // The modal may have closed while loading.
    if (!previewEl.value) return
    qrCode = new QRCodeStyling(buildOptions())
    previewEl.value.replaceChildren()
    qrCode.append(previewEl.value)
    return
  }

  qrCode.update(buildOptions())
}

watch([open, color, shape, () => props.url], render, { immediate: true })

// A fresh instance per open.
watch(open, (isOpen) => {
  if (!isOpen) qrCode = undefined
})

async function download(extension: 'svg' | 'png') {
  if (!qrCode) return
  downloading.value = extension
  try {
    await qrCode.download({ name: 'qr-code', extension })
  } finally {
    downloading.value = null
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="t('app.shelf.detail.qrCode')"
    :description="title ? t('app.shelf.qr.scanTitle', { title }) : t('app.shelf.qr.scan')"
  >
    <template #body>
      <div class="flex flex-col items-center gap-6">
        <div
          ref="previewEl"
          class="flex items-center justify-center rounded-lg bg-white p-4"
        />

        <div class="w-full flex flex-col gap-4">
          <UFormField :label="t('app.link.form.color')">
            <div class="flex items-center gap-2">
              <UColorPicker
                v-model="color"
                format="hex"
              />
              <UInput
                v-model="color"
                class="w-full"
                placeholder="#000000"
              />
            </div>
          </UFormField>

          <UFormField :label="t('app.shelf.qr.shape')">
            <USelect
              v-model="shape"
              :items="shapeOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>
        </div>
      </div>
    </template>

    <template #footer="{ close }">
      <UButton
        :label="t('common.close')"
        color="neutral"
        variant="outline"
        @click="close"
      />
      <UButton
        :label="t('app.shelf.qr.downloadSvg')"
        icon="i-lucide-download"
        color="neutral"
        variant="outline"
        :loading="downloading === 'svg'"
        @click="download('svg')"
      />
      <UButton
        :label="t('app.shelf.qr.downloadPng')"
        icon="i-lucide-download"
        :loading="downloading === 'png'"
        @click="download('png')"
      />
    </template>
  </UModal>
</template>

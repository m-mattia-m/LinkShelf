<script setup lang="ts">
import type { DotType, CornerDotType, CornerSquareType } from 'qr-code-styling'

const open = defineModel<boolean>('open', { default: false })

const props = defineProps<{
  url: string
  title?: string
}>()

const shapeOptions: { label: string, value: DotType }[] = [
  { label: 'Square', value: 'square' },
  { label: 'Rounded', value: 'rounded' },
  { label: 'Extra rounded', value: 'extra-rounded' },
  { label: 'Dots', value: 'dots' }
]

const color = ref('#000000')
const shape = ref<DotType>('square')
const downloading = ref<'svg' | 'png' | null>(null)

// The big "eyes" (corner squares/dots) don't share the --dots-- style names,
// so each shape maps to the closest-looking corner pair rather than a 1:1
// property - otherwise picking "Dots" would still leave sharp square eyes,
// which reads as unfinished rather than as one cohesive style.
const cornerStylesByShape: Record<DotType, { square: CornerSquareType, dot: CornerDotType }> = {
  'square': { square: 'square', dot: 'square' },
  'rounded': { square: 'extra-rounded', dot: 'dot' },
  'extra-rounded': { square: 'extra-rounded', dot: 'dot' },
  'dots': { square: 'dot', dot: 'dot' },
  'classy': { square: 'extra-rounded', dot: 'dot' },
  'classy-rounded': { square: 'extra-rounded', dot: 'dot' }
}

const previewEl = ref<HTMLElement>()
// qr-code-styling renders straight to the DOM (canvas/svg) rather than
// through Vue's own reactivity, so the instance is plain, unreactive state
// kept alongside the component instead of in a ref.
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
    // The modal may have closed (or the component unmounted) while the
    // chunk was loading - nothing left to render into.
    if (!previewEl.value) return
    qrCode = new QRCodeStyling(buildOptions())
    previewEl.value.replaceChildren()
    qrCode.append(previewEl.value)
    return
  }

  qrCode.update(buildOptions())
}

watch([open, color, shape, () => props.url], render, { immediate: true })

// A fresh instance per time the modal opens, rather than reusing one across
// opens - simpler than tracking whether the old preview element (torn down
// with the modal's content) is still attached.
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
    title="QR code"
    :description="title ? `Scan to open &quot;${title}&quot;` : 'Scan to open this shelf'"
  >
    <template #body>
      <div class="flex flex-col items-center gap-6">
        <div
          ref="previewEl"
          class="flex items-center justify-center rounded-lg bg-white p-4"
        />

        <div class="w-full flex flex-col gap-4">
          <UFormField label="Color">
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

          <UFormField label="Shape">
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
        label="Close"
        color="neutral"
        variant="outline"
        @click="close"
      />
      <UButton
        label="Download SVG"
        icon="i-lucide-download"
        color="neutral"
        variant="outline"
        :loading="downloading === 'svg'"
        @click="download('svg')"
      />
      <UButton
        label="Download PNG"
        icon="i-lucide-download"
        :loading="downloading === 'png'"
        @click="download('png')"
      />
    </template>
  </UModal>
</template>

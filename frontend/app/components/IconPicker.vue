<script setup lang="ts">
// Icon names are generated from the installed @iconify-json/lucide and
// @iconify-json/simple-icons packages - see public/icon-names.json. Loaded
// lazily (only once the picker is actually opened) and cached across every
// instance of this component via useState, so opening a second picker on the
// same page is instant.
interface IconNameSets {
  lucide: string[]
  'simple-icons': string[]
}

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
}>(), {
  placeholder: 'i-lucide-link'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const open = ref(false)
const search = ref('')
const allIcons = useState<string[]>('icon-picker-names', () => [])
const loading = ref(false)
const loadFailed = ref(false)

async function ensureIconsLoaded() {
  if (allIcons.value.length > 0 || loading.value) return
  loading.value = true
  loadFailed.value = false
  try {
    const data = await $fetch<IconNameSets>('/icon-names.json')
    allIcons.value = [...data.lucide, ...data['simple-icons']]
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
}

watch(open, (isOpen) => {
  if (isOpen) ensureIconsLoaded()
})

const MAX_RESULTS = 120

const matchingIcons = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return allIcons.value
  return allIcons.value.filter((name) => name.toLowerCase().includes(query))
})

const visibleIcons = computed(() => matchingIcons.value.slice(0, MAX_RESULTS))

function select(name: string) {
  emit('update:modelValue', name)
  open.value = false
}
</script>

<template>
  <UPopover v-model:open="open">
    <UButton color="neutral" variant="outline" class="justify-start">
      <UIcon :name="modelValue || placeholder" class="size-5 shrink-0" />
      <span class="truncate max-w-40">{{ modelValue || placeholder }}</span>
      <template #trailing>
        <UIcon name="i-lucide-chevron-down" class="size-4 shrink-0" />
      </template>
    </UButton>

    <template #content>
      <div class="p-3 w-80">
        <UInput
          v-model="search"
          icon="i-lucide-search"
          placeholder="Search icons..."
          class="w-full mb-2"
          autofocus
        />

        <div v-if="loading" class="flex justify-center py-6">
          <UIcon name="i-lucide-loader-circle" class="size-5 animate-spin text-muted" />
        </div>

        <p v-else-if="loadFailed" class="text-xs text-error py-2">
          Could not load the icon list. You can still type an icon name directly.
        </p>

        <template v-else>
          <div class="grid grid-cols-8 gap-1 max-h-64 overflow-y-auto">
            <button
              v-for="name in visibleIcons"
              :key="name"
              type="button"
              class="flex items-center justify-center rounded p-2 hover:bg-elevated cursor-pointer"
              :class="{ 'bg-elevated': name === modelValue }"
              :title="name"
              @click="select(name)"
            >
              <UIcon :name="name" class="size-5" />
            </button>
          </div>

          <p v-if="matchingIcons.length > visibleIcons.length" class="text-xs text-muted pt-2">
            Showing {{ visibleIcons.length }} of {{ matchingIcons.length }} - keep typing to narrow it down.
          </p>
          <p v-else-if="visibleIcons.length === 0" class="text-xs text-muted pt-2">
            No icons found for "{{ search }}".
          </p>
        </template>
      </div>
    </template>
  </UPopover>
</template>

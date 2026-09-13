<script setup lang="ts">
import { reactive } from 'vue'
import draggable from 'vuedraggable'
import type { Shelf, Section, Link } from '~~/api'
import { useShelfStore } from '~/stores/shelf'
import { useSectionStore } from '~/stores/section'
import { useLinkStore } from '~/stores/link'
import ShelfFormDialog from '~/components/shelf/ShelfFormDialog.vue'
import SectionCard from '~/components/shelf/ShelfSectionCard.vue'

definePageMeta({
  layout: 'app'
})

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const shelfId = route.params.shelfId as string

const shelfStore = useShelfStore()
const sectionStore = useSectionStore()
const linkStore = useLinkStore()

const shelf = ref<Shelf>()
const loading = ref(true)
const notFound = ref(false)

// Local, unsaved drag order - dragging a section or link only ever mutates
// these; nothing is persisted until "Save order" is clicked. Re-derived from
// the stores whenever they refetch (create/delete/initial load), which also
// discards any in-progress unsaved drag in the rare case one of those
// happens concurrently.
const orderedSections = ref<Section[]>([])
const linkOrders = reactive<Record<string, Link[]>>({})
const orderDirty = ref(false)
const savingOrder = ref(false)

function resetLocalOrder() {
  orderedSections.value = [...sectionStore.sections].sort((a, b) => a.order - b.order)
  for (const key of Object.keys(linkOrders)) Reflect.deleteProperty(linkOrders, key)
  for (const section of orderedSections.value) {
    linkOrders[section.id] = [...(linkStore.bySectionId.get(section.id) ?? [])].sort((a, b) => a.order - b.order)
  }
  orderDirty.value = false
}

watch(() => sectionStore.sections, resetLocalOrder)
watch(() => linkStore.links, resetLocalOrder)

async function saveOrder() {
  savingOrder.value = true
  try {
    const api = useApi()
    const sections = orderedSections.value.map((s, i) => ({ id: s.id, order: i }))
    const links = orderedSections.value.flatMap(s => (linkOrders[s.id] ?? []).map((l, i) => ({ id: l.id, order: i })))

    await Promise.all([
      sections.length ? api.section.putUpdateSectionsOrder({ sectionOrderRequestBody: { sections } }) : undefined,
      links.length ? api.link.putUpdateLinksOrder({ linkOrderRequestBody: { links } }) : undefined
    ])

    await Promise.all([sectionStore.fetch(shelfId), linkStore.fetch(shelfId)])
  } catch (err) {
    await handleApiError(err)
  } finally {
    savingOrder.value = false
  }
}

// Unsaved order changes are easy to lose silently (tab close, in-app nav) -
// warn before either. Deliberately no autosave-on-drop: with several active
// users this would fire a request per drag frame instead of one on demand.
function onBeforeUnload(e: BeforeUnloadEvent) {
  if (!orderDirty.value) return
  e.preventDefault()
  e.returnValue = ''
}
onMounted(() => window.addEventListener('beforeunload', onBeforeUnload))
onBeforeUnmount(() => window.removeEventListener('beforeunload', onBeforeUnload))
onBeforeRouteLeave(() => {
  if (!orderDirty.value) return true
  return window.confirm('You have unsaved order changes. Leave without saving?')
})

async function loadAll() {
  loading.value = true
  try {
    shelf.value = await shelfStore.getById(shelfId)
    await Promise.all([
      sectionStore.fetch(shelfId),
      linkStore.fetch(shelfId)
    ])
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)

const editOpen = ref(false)

function onShelfSaved(updated: Shelf) {
  shelf.value = updated
}

const deleteOpen = ref(false)
const deleting = ref(false)

async function confirmDeleteShelf() {
  if (!shelf.value) return
  deleting.value = true
  try {
    await shelfStore.remove(shelf.value.id)
    deleteOpen.value = false
    await router.push('/app/shelf')
  } catch (err) {
    await handleApiError(err)
  } finally {
    deleting.value = false
  }
}

const newSectionTitle = ref('')
const creatingSection = ref(false)

async function createSection() {
  if (!newSectionTitle.value.trim() || !shelf.value) return
  creatingSection.value = true
  try {
    await sectionStore.create({ title: newSectionTitle.value, shelfId: shelf.value.id })
    newSectionTitle.value = ''
  } catch (err) {
    await handleApiError(err)
  } finally {
    creatingSection.value = false
  }
}
</script>

<template>
  <div>
    <div
      v-if="loading"
      class="space-y-4"
    >
      <USkeleton class="h-8 w-1/3" />
      <USkeleton class="h-24 w-full" />
      <USkeleton class="h-40 w-full" />
    </div>

    <div
      v-else-if="notFound"
      class="text-center py-16"
    >
      <p class="text-muted">
        {{ t('app.shelf.detail.notFound') }}
      </p>
      <ULink
        to="/app/shelf"
        class="text-primary"
      >{{ t('app.shelf.detail.backToList') }}</ULink>
    </div>

    <template v-else-if="shelf">
      <div class="flex justify-between items-start gap-4 pb-4">
        <div>
          <h1 class="text-2xl text-highlighted">
            {{ shelf.title }}
          </h1>
          <p
            v-if="shelf.description"
            class="text-muted"
          >
            {{ shelf.description }}
          </p>
          <p class="text-sm text-dimmed">
            <span v-if="shelf.path">/{{ shelf.path }}</span>
            <span v-if="shelf.domain">{{ shelf.path ? ' · ' : '' }}{{ shelf.domain }}</span>
          </p>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <UButton
            :label="t('app.shelf.detail.open')"
            icon="i-lucide-external-link"
            color="neutral"
            variant="outline"
            :href="'/' + shelf.path"
            target="_blank"
          />
          <UButton
            :label="t('app.shelf.detail.edit')"
            icon="i-lucide-pencil"
            color="neutral"
            variant="outline"
            @click="editOpen = true"
          />
          <UButton
            :label="t('app.shelf.detail.delete')"
            icon="i-lucide-trash-2"
            color="error"
            variant="outline"
            @click="deleteOpen = true"
          />
        </div>
      </div>

      <div class="flex items-center gap-2 pb-4">
        <UInput
          v-model="newSectionTitle"
          :placeholder="t('app.section.newPlaceholder')"
          class="max-w-xs"
          @keyup.enter="createSection"
        />
        <UButton
          :label="t('app.section.new')"
          icon="i-lucide-plus"
          color="neutral"
          :loading="creatingSection"
          :disabled="!newSectionTitle.trim()"
          @click="createSection"
        />
      </div>

      <div
        v-if="sectionStore.sections.length === 0"
        class="text-center text-muted py-12"
      >
        {{ t('app.section.empty') }}
      </div>

      <template v-else>
        <div class="flex justify-end pb-2">
          <UButton
            label="Save order"
            icon="i-lucide-save"
            color="primary"
            :disabled="!orderDirty"
            :loading="savingOrder"
            @click="saveOrder"
          />
        </div>

        <draggable
          v-model="orderedSections"
          item-key="id"
          handle=".drag-handle"
          tag="div"
          class="flex flex-col gap-4"
          @end="orderDirty = true"
        >
          <template #item="{ element: section }">
            <SectionCard
              v-model:links="linkOrders[section.id]!"
              :section="section"
              @reordered="orderDirty = true"
            />
          </template>
        </draggable>
      </template>

      <ShelfFormDialog
        v-model:open="editOpen"
        mode="edit"
        :shelf="shelf"
        @saved="onShelfSaved"
      />

      <ConfirmDialog
        v-model:open="deleteOpen"
        :title="t('app.shelf.deleteConfirm.title')"
        :description="t('app.shelf.deleteConfirm.descriptionWithCounts', {
          title: shelf.title,
          sections: sectionStore.sections.length,
          links: linkStore.links.length
        })"
        :loading="deleting"
        @confirm="confirmDeleteShelf"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import draggable from 'vuedraggable'
import type { Section, Link } from '~~/api'
import { useSectionStore } from '~/stores/section'
import { useLinkStore } from '~/stores/link'

const props = defineProps<{
  section: Section
}>()

const emit = defineEmits<{
  (e: 'reordered'): void
}>()

// Two-way bound with the parent's local, unsaved order for this section's
// links - dragging here only ever touches this local copy; nothing is
// persisted until the parent's single "Save order" button is clicked.
const links = defineModel<Link[]>('links', { required: true })

const sectionStore = useSectionStore()
const linkStore = useLinkStore()

const renaming = ref(false)
const titleDraft = ref(props.section.title)
const renameLoading = ref(false)

function startRename() {
  titleDraft.value = props.section.title
  renaming.value = true
}

async function confirmRename() {
  if (!titleDraft.value.trim()) return
  renameLoading.value = true
  try {
    await sectionStore.update(props.section.id, { title: titleDraft.value, shelfId: props.section.shelfId })
    renaming.value = false
  } catch (err) {
    await handleApiError(err)
  } finally {
    renameLoading.value = false
  }
}

const deleteSectionOpen = ref(false)
const deletingSection = ref(false)

async function confirmDeleteSection() {
  deletingSection.value = true
  try {
    await sectionStore.remove(props.section.id, props.section.shelfId)
    deleteSectionOpen.value = false
  } catch (err) {
    await handleApiError(err)
  } finally {
    deletingSection.value = false
  }
}

const linkDialogOpen = ref(false)
const linkDialogMode = ref<'create' | 'edit'>('create')
const editingLink = ref<Link | undefined>()

function openCreateLink() {
  linkDialogMode.value = 'create'
  editingLink.value = undefined
  linkDialogOpen.value = true
}

function openEditLink(link: Link) {
  linkDialogMode.value = 'edit'
  editingLink.value = link
  linkDialogOpen.value = true
}

const deleteLinkOpen = ref(false)
const deletingLinkTarget = ref<Link | null>(null)
const deletingLink = ref(false)

function openDeleteLink(link: Link) {
  deletingLinkTarget.value = link
  deleteLinkOpen.value = true
}

async function confirmDeleteLink() {
  if (!deletingLinkTarget.value) return
  deletingLink.value = true
  try {
    await linkStore.remove(deletingLinkTarget.value.id)
    deleteLinkOpen.value = false
  } catch (err) {
    await handleApiError(err)
  } finally {
    deletingLink.value = false
  }
}
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between gap-2 flex-wrap">
        <div v-if="!renaming" class="flex items-center gap-1">
          <UIcon name="i-lucide-grip-vertical" class="drag-handle size-4 text-dimmed cursor-grab shrink-0" aria-label="Drag to reorder section" />
          <h3 class="font-medium">{{ section.title }}</h3>
          <UButton icon="i-lucide-pencil" size="xs" color="neutral" variant="ghost" aria-label="Rename section" @click="startRename" />
        </div>
        <div v-else class="flex items-center gap-2">
          <UInput v-model="titleDraft" size="sm" autofocus @keyup.enter="confirmRename" @keyup.escape="renaming = false" />
          <UButton icon="i-lucide-check" size="xs" color="primary" :loading="renameLoading" aria-label="Save section title" @click="confirmRename" />
          <UButton icon="i-lucide-x" size="xs" color="neutral" variant="ghost" aria-label="Cancel rename" @click="renaming = false" />
        </div>

        <div class="flex items-center gap-2">
          <UButton icon="i-lucide-plus" size="xs" label="New link" color="neutral" variant="outline" @click="openCreateLink" />
          <UButton icon="i-lucide-trash-2" size="xs" color="error" variant="ghost" aria-label="Delete section" @click="deleteSectionOpen = true" />
        </div>
      </div>
    </template>

    <p v-if="links.length === 0" class="text-center text-muted text-sm py-6">
      No links yet.
      <UButton label="Add the first link" variant="link" @click="openCreateLink" />
    </p>

    <draggable
      v-else
      v-model="links"
      item-key="id"
      handle=".drag-handle"
      tag="div"
      class="flex flex-col divide-y divide-default"
      @end="emit('reordered')"
    >
      <template #item="{ element: link }">
        <div class="flex items-center gap-2 py-2">
          <UIcon name="i-lucide-grip-vertical" class="drag-handle size-4 text-dimmed cursor-grab shrink-0" aria-label="Drag to reorder link" />
          <UIcon :name="link.icon || 'i-lucide-link'" class="size-4 shrink-0" :style="{ color: link.color }" />
          <div class="min-w-0 flex-1">
            <p class="font-medium truncate">{{ link.title }}</p>
            <ULink :href="link.link" target="_blank" class="text-xs text-dimmed truncate block">{{ link.link }}</ULink>
          </div>
          <UButton icon="i-lucide-pencil" size="xs" color="neutral" variant="ghost" aria-label="Edit link" @click="openEditLink(link)" />
          <UButton icon="i-lucide-trash-2" size="xs" color="error" variant="ghost" aria-label="Delete link" @click="openDeleteLink(link)" />
        </div>
      </template>
    </draggable>

    <ShelfLinkFormDialog
      v-model:open="linkDialogOpen"
      :mode="linkDialogMode"
      :section-id="section.id"
      :link="editingLink"
    />

    <ConfirmDialog
      v-model:open="deleteSectionOpen"
      title="Delete section?"
      :description="`This also deletes ${links.length} link(s) in “${section.title}”. This cannot be undone.`"
      :loading="deletingSection"
      @confirm="confirmDeleteSection"
    />

    <ConfirmDialog
      v-model:open="deleteLinkOpen"
      title="Delete link?"
      :description="`Delete “${deletingLinkTarget?.title}”? This cannot be undone.`"
      :loading="deletingLink"
      @confirm="confirmDeleteLink"
    />
  </UCard>
</template>

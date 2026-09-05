<script setup lang="ts">
import { ref } from 'vue'
import type { Section, Link } from '~~/api'
import type { TableColumn } from '#ui/components/Table.vue'
import { useSectionStore } from '~/stores/section'
import { useLinkStore } from '~/stores/link'

const props = defineProps<{
  section: Section
  links: Link[]
}>()

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

const columns: TableColumn<Link>[] = [
  { accessorKey: 'title', header: 'Title' },
  { accessorKey: 'link', header: 'URL' },
  { id: 'action' }
]
</script>

<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between gap-2 flex-wrap">
        <div v-if="!renaming" class="flex items-center gap-2">
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

    <UTable :columns="columns" :data="links">
      <template #link-cell="{ row }">
        <ULink :href="row.original.link" target="_blank" class="truncate block max-w-xs">{{ row.original.link }}</ULink>
      </template>
      <template #action-cell="{ row }">
        <div class="flex items-center justify-end gap-1">
          <UIcon :name="row.original.icon || 'i-lucide-link'" class="size-4" :style="{ color: row.original.color }" />
          <UButton icon="i-lucide-pencil" size="xs" color="neutral" variant="ghost" aria-label="Edit link" @click="openEditLink(row.original)" />
          <UButton icon="i-lucide-trash-2" size="xs" color="error" variant="ghost" aria-label="Delete link" @click="openDeleteLink(row.original)" />
        </div>
      </template>
      <template #empty>
        <div class="text-center text-muted text-sm py-6">
          No links yet.
          <UButton label="Add the first link" variant="link" @click="openCreateLink" />
        </div>
      </template>
    </UTable>

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

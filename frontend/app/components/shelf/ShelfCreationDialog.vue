<script setup lang="ts">
import type {Shelf, PostCreateShelfRequest} from "~~/api";
import {useShelfStore} from "~/stores/shelf";

const shelfStore = useShelfStore()

const openCreationModalState = ref<boolean>()

const shelf = ref<Shelf>();
const isValid = ref(false)

const formRef = ref<{
  validate: () => Promise<boolean>
} | null>(null)


async function save() {
  const valid = await formRef.value?.validate()
  if (!valid) return

  const request: PostCreateShelfRequest = {
    shelfBase: shelf.value!!
  }

  await shelfStore.create(request)
}


</script>

<template>
  <UModal
    :open="openCreationModalState"
    title="Shelf"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton
      @click="openCreationModalState = true"
      icon="i-lucide-plus"
      label="New"
    />

    <template #body>
      <ShelfForm
        ref="formRef"
        v-model="shelf"
        v-model:validate="isValid"
      />
    </template>

    <template #footer="{ close }">
      <UButton label="Cancel" color="neutral" variant="outline" @click="close"/>
      <UButton label="Submit" color="neutral" @click="save"/>
    </template>
  </UModal>
</template>

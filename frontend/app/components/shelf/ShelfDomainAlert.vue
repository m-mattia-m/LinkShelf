<script setup lang="ts">
import type { Shelf } from '~~/api'

const props = defineProps<{
  shelf: Shelf
  checking?: boolean
}>()

const emit = defineEmits<{
  check: []
}>()

const { t } = useI18n()

// A domain which is ready needs nothing. One which is not has to be seen on the main screen and not only
// in the form: the user has something to do, and until then the shelf is not reachable on it.
const state = computed<'pending' | 'error' | null>(() => {
  if (!props.shelf.domain || !props.shelf.domainStatus) return null
  const current = props.shelf.domainStatus.state
  return current === 'pending' || current === 'error' ? current : null
})

// The records are long, so they are folded away: the notice itself stays a line or two.
const open = ref(false)
const recordsId = useId()
const recordCount = computed(() => props.shelf.domainStatus?.records?.length ?? 0)
</script>

<template>
  <UAlert
    v-if="state && shelf.domainStatus"
    :color="state === 'error' ? 'error' : 'warning'"
    variant="subtle"
    :icon="state === 'error' ? 'i-lucide-circle-alert' : 'i-lucide-clock'"
    :title="t(`app.shelf.domainAlert.${state}.title`)"
    :description="t(`app.shelf.domainAlert.${state}.description`, { domain: shelf.domain })"
    :ui="{ root: 'p-3', title: 'text-sm', description: 'text-sm' }"
    data-testid="domain-alert"
  >
    <template #actions>
      <div class="flex flex-col gap-2">
        <div class="flex flex-wrap items-center gap-2">
          <UButton
            v-if="recordCount"
            :label="open ? t('app.shelf.domainAlert.hide') : t('app.shelf.domainAlert.show', { count: recordCount })"
            :trailing-icon="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
            color="neutral"
            variant="outline"
            size="xs"
            :aria-expanded="open"
            :aria-controls="recordsId"
            data-testid="domain-alert-toggle"
            @click="open = !open"
          />
          <UButton
            v-if="state === 'pending'"
            :label="t('app.shelf.domainAlert.check')"
            icon="i-lucide-refresh-cw"
            color="neutral"
            variant="ghost"
            size="xs"
            :loading="checking"
            data-testid="domain-alert-check"
            @click="emit('check')"
          />
        </div>

        <!-- The text is the normal one, the yellow of the notice is for its title. -->
        <div
          v-if="open"
          :id="recordsId"
          class="text-default"
          data-testid="domain-alert-records"
        >
          <ShelfDomainStatus
            :status="shelf.domainStatus"
            embedded
          />
        </div>
      </div>
    </template>
  </UAlert>
</template>

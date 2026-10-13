<script setup lang="ts">
import type { DomainRecord, DomainStatus } from '~~/api'

const props = defineProps<{
  status: DomainStatus
  // Without its own border and title, for use inside another box (the alert on the shelf page).
  embedded?: boolean
}>()

const { t, te } = useI18n()
const toast = useToast()

// LinkShelf attaches no meaning to these values: the state is a badge, the reason is
// translated when known and the controller's own message is the fallback.
const badge = computed(() => {
  switch (props.status.state) {
    case 'ready':
      return { color: 'success' as const, icon: 'i-lucide-circle-check' }
    case 'error':
      return { color: 'error' as const, icon: 'i-lucide-circle-alert' }
    default:
      return { color: 'warning' as const, icon: 'i-lucide-clock' }
  }
})

const stateLabel = computed(() => te(`app.shelf.domainStatus.state.${props.status.state}`)
  ? t(`app.shelf.domainStatus.state.${props.status.state}`)
  : props.status.state)

const reasonLabel = computed(() => {
  const key = `app.shelf.domainStatus.reason.${props.status.reason}`
  if (props.status.reason && te(key)) return t(key)
  return props.status.message || props.status.reason
})

// What a record is for, by its type. The values of the record stay exactly as the controller wrote
// them; only the explanation around them is known here, with the type as the fallback.
function recordTitle(record: DomainRecord): string {
  const key = `app.shelf.domainStatus.record.${record.type}.title`
  return te(key) ? t(key) : record.type
}

function recordHint(record: DomainRecord): string {
  const key = `app.shelf.domainStatus.record.${record.type}.hint`
  return te(key) ? t(key) : ''
}

async function copy(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    toast.add({ title: t('app.shelf.domainStatus.copied'), color: 'success', icon: 'i-lucide-clipboard-check' })
  } catch {
    // Without clipboard access the value can still be selected by hand, it is select-all.
  }
}
</script>

<template>
  <div
    :class="embedded ? 'space-y-3 text-sm' : 'mt-3 space-y-3 rounded-md border border-default p-3 text-sm'"
    data-testid="domain-status"
  >
    <div
      v-if="!embedded || reasonLabel"
      class="flex flex-wrap items-center gap-2"
    >
      <template v-if="!embedded">
        <span class="font-medium">{{ t('app.shelf.domainStatus.title') }}</span>
        <UBadge
          :color="badge.color"
          :icon="badge.icon"
          variant="subtle"
          data-testid="domain-status-state"
        >
          {{ stateLabel }}
        </UBadge>
      </template>
      <span
        v-if="reasonLabel"
        :class="embedded ? 'text-sm' : 'text-muted'"
        data-testid="domain-status-reason"
      >{{ reasonLabel }}</span>
    </div>

    <template v-if="status.records?.length">
      <p
        v-if="!embedded"
        class="text-muted"
      >
        {{ t('app.shelf.domainStatus.recordsHelp') }}
      </p>
      <ol :class="embedded ? 'space-y-2' : 'space-y-3'">
        <li
          v-for="(record, index) in status.records"
          :key="index"
          class="rounded-md bg-elevated"
          :class="embedded ? 'p-2' : 'p-3'"
          data-testid="domain-status-record"
        >
          <p class="flex flex-wrap items-center gap-2 font-medium text-highlighted">
            <span>{{ index + 1 }}. {{ recordTitle(record) }}</span>
            <UBadge
              color="neutral"
              variant="outline"
              size="sm"
              class="font-mono"
              data-testid="domain-status-record-type"
            >
              {{ record.type }}
            </UBadge>
          </p>
          <p
            v-if="recordHint(record)"
            class="pb-1 text-xs text-muted"
          >
            {{ recordHint(record) }}
          </p>

          <dl class="grid grid-cols-[3rem_1fr_auto] items-center gap-x-2 gap-y-0.5 text-xs">
            <dt class="text-muted">
              {{ t('app.shelf.domainStatus.name') }}
            </dt>
            <dd
              class="break-all font-mono select-all"
              data-testid="domain-status-record-name"
            >
              {{ record.name }}
            </dd>
            <dd>
              <UButton
                icon="i-lucide-copy"
                color="neutral"
                variant="ghost"
                size="xs"
                :aria-label="`${t('app.shelf.domainStatus.copy')}: ${t('app.shelf.domainStatus.name')}`"
                @click="copy(record.name)"
              />
            </dd>

            <dt class="text-muted">
              {{ t('app.shelf.domainStatus.value') }}
            </dt>
            <dd
              class="break-all font-mono select-all"
              data-testid="domain-status-record-value"
            >
              {{ record.value }}
            </dd>
            <dd>
              <UButton
                icon="i-lucide-copy"
                color="neutral"
                variant="ghost"
                size="xs"
                :aria-label="`${t('app.shelf.domainStatus.copy')}: ${t('app.shelf.domainStatus.value')}`"
                @click="copy(record.value)"
              />
            </dd>
          </dl>
        </li>
      </ol>
    </template>
  </div>
</template>

<script setup lang="ts">
import { mailStatus } from '@/composables/mail'

// A rejected mail is a red badge that tells the error on hover, an accepted one is plain text.
defineProps<{ data: SmtpEventData }>()
</script>

<template>
  <span v-if="mailStatus(data) === 'Rejected'" class="badge bg-danger" :title="data.error">Rejected</span>
  <template v-else>Accepted</template>
</template>

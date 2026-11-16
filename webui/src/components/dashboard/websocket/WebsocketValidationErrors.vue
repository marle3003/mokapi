<script setup lang="ts">
import ValidationErrorsTable from './message/ValidationErrorsTable.vue'

defineProps<{
    channel?: WebsocketChannel
    validationErrors: Record<string, string>
    value?: string
}>()
</script>

<template>
    <div class="card-group">
        <section class="card" aria-labelledby="validation-errors">
            <div class="card-body">
                <h2 id="validation-errors" class="card-title text-center">Validation Errors</h2>

                <validation-errors-table :channel="channel" :validation-errors="validationErrors" :value="value" />
            </div>
        </section>
    </div>
</template>

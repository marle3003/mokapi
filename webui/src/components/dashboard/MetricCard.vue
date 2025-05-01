<script setup lang="ts">

const props = defineProps<{
    title: string
    value: string | number
    additional?: string
    live?: "off" | "assertive" | "polite" | undefined

}>()

const labelId = 'metric-'+props.title.replaceAll(' ', '-')
</script>

<template>
    <div class="card text-center">
        <div class="card-body">
            <div class="card-title" :id="labelId">{{ title }}</div>
            <div class="card-text" role="status" :aria-live="live ? live : undefined" :aria-labelledby="labelId">
                {{ value }}
            </div>
            <div class="card-additional" v-if="additional">{{ additional }}</div>
        </div>
    </div>
</template>
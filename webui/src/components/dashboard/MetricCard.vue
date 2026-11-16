<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'

const props = defineProps<{
    title: string
    value?: string | number
    additional?: string
    additionalLabel?: string
    live?: "off" | "assertive" | "polite" | undefined
    // makes the whole card a link to this route
    to?: RouteLocationRaw
}>()

const labelId = 'metric-'+props.title.replaceAll(' ', '-')
// a link should not announce every refresh of its value
const ariaLive = computed(() => props.to ? 'off' : props.live)
</script>

<template>
    <component :is="to ? RouterLink : 'div'" :to="to" class="card text-center" :class="{ 'metric-tile': to }">
        <div class="card-body">
            <div class="card-title" :id="labelId">{{ title }}</div>
            <div class="card-text metric-value" role="status" :aria-live="ariaLive" :aria-labelledby="labelId">
                <slot name="value">{{ value }}</slot>
            </div>
            <div class="card-additional" v-if="additional" :aria-label="additionalLabel">{{ additional }}</div>
        </div>
    </component>
</template>

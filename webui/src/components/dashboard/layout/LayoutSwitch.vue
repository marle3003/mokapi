<script setup lang="ts">
import { useDashboardLayout } from '@/composables/dashboard-layout'

const { layout, setLayout } = useDashboardLayout()
</script>

<template>
    <button v-if="layout === 'new'" type="button" class="btn layout-switch" @click="setLayout('classic')">Classic view</button>
    <button v-else type="button" class="btn layout-switch" @click="setLayout('new')">New view</button>
</template>

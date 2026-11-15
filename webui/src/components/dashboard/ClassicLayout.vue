<script setup lang="ts">
import DashboardViews from '@/components/dashboard/DashboardViews.vue'
import Tabs from '@/components/dashboard/Tabs.vue'
import type { AppInfo } from '@/types/dashboard'
import type { PropType } from 'vue'

defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
})
</script>

<template>
    <div class="dashboard-tabs">
        <tabs :app-info="appInfo" />
    </div>
    <div class="dashboard-content">
        <dashboard-views :key="$route.fullPath" :app-info="appInfo" />
    </div>
</template>

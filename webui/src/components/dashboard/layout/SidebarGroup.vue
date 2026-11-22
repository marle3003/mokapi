<script setup lang="ts">
import type { PropType } from 'vue'
import { getRouteName } from '@/composables/dashboard'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import type { SidebarGroup } from '@/composables/sidebar-groups'
import { useRoute } from '@/router'

defineProps({
    group: { type: Object as PropType<SidebarGroup>, required: true },
})

const { isGroupCollapsed, toggleGroup } = useDashboardLayout()
const { service } = useRoute()
</script>

<template>
    <div class="sidebar-group">
        <div class="sidebar-group-header">
            <button type="button" class="sidebar-chevron" :aria-expanded="!isGroupCollapsed(group.type)"
                :aria-label="`${isGroupCollapsed(group.type) ? 'Expand' : 'Collapse'} ${group.title}`"
                @click="toggleGroup(group.type)">
                <i class="bi" :class="isGroupCollapsed(group.type) ? 'bi-chevron-right' : 'bi-chevron-down'" aria-hidden="true"></i>
            </button>
            <router-link class="sidebar-group-title" :to="{ name: getRouteName(group.type).value }">{{ group.title }}</router-link>
        </div>
        <ul v-if="!isGroupCollapsed(group.type)" class="sidebar-services" :aria-label="`${group.title} services`">
            <li v-for="s in group.services" :key="s.name">
                <router-link class="sidebar-entry" :to="service(s.name, group.type)">
                    <span class="sidebar-entry-name">{{ s.name }}</span>
                    <i v-if="s.hasErrors" class="bi bi-circle-fill sidebar-error-dot" role="img" aria-label="Service has errors"></i>
                    <span class="sidebar-count">{{ s.count }}</span>
                </router-link>
            </li>
        </ul>
    </div>
</template>

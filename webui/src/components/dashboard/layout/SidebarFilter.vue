<script setup lang="ts">
import { computed, ref, type PropType } from 'vue'
import { getRouteName } from '@/composables/dashboard'
import { filterSidebar, isFilterActive, type FilterTarget } from '@/composables/sidebar-filter'

const props = defineProps({
    services: { type: Array as PropType<Service[]>, required: true },
})
const query = defineModel<string>({ required: true })

const input = ref<HTMLInputElement>()
const result = computed(() => filterSidebar(props.services, query.value))

function link(target: FilterTarget) {
    return { name: getRouteName(target.name).value, params: target.params }
}

function clear() {
    query.value = ''
    input.value?.focus()
}
</script>

<template>
    <div class="sidebar-filter">
        <div class="sidebar-filter-field">
            <input ref="input" v-model="query" type="search" class="form-control form-control-sm" placeholder="Filter services"
                aria-label="Filter services" autocomplete="off" @keydown.esc="query = ''" />
            <button v-if="query" type="button" class="sidebar-filter-clear" aria-label="Clear filter" @click="clear">
                <i class="bi bi-x-lg" aria-hidden="true"></i>
            </button>
        </div>
        <div v-if="isFilterActive(query)" class="sidebar-filter-results">
            <p class="sidebar-hint" data-testid="filter-count">{{ result.summary }}</p>
            <div v-for="group in result.groups" :key="group.service.name" class="sidebar-filter-group" role="group" :aria-label="group.service.name">
                <router-link v-if="group.service.matched" class="sidebar-group-title" :to="link(group.service.to)">{{ group.service.name }}</router-link>
                <span v-else class="sidebar-group-title">{{ group.service.name }}</span>
                <ul>
                    <li v-for="item in group.items" :key="item.method + item.name">
                        <router-link class="sidebar-entry" :to="link(item.to)">
                            <span v-if="item.method" class="badge operation" :class="item.method">{{ item.method.toUpperCase() }}</span>
                            <span class="sidebar-entry-name">{{ item.name }}</span>
                        </router-link>
                    </li>
                    <li v-if="group.more">
                        <router-link class="sidebar-entry sidebar-more" :to="link(group.moreTo)">{{ group.more }} more in {{ group.service.name }}</router-link>
                    </li>
                </ul>
            </div>
        </div>
    </div>
</template>

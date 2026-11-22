<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch, type Component } from 'vue'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import { facetTitle, fullSearchLocation, searchTotalLabel } from '@/composables/search'
import { useSearchPanel } from '@/composables/search-panel'
import SearchTips from '@/components/dashboard/SearchTips.vue'
import ConfigRow from '@/components/dashboard/search/ConfigRow.vue'
import EventRow from '@/components/dashboard/search/EventRow.vue'
import HttpRow from '@/components/dashboard/search/HttpRow.vue'
import KafkaRow from '@/components/dashboard/search/KafkaRow.vue'
import LdapRow from '@/components/dashboard/search/LdapRow.vue'
import MailRow from '@/components/dashboard/search/MailRow.vue'

const rowByType: { [type: string]: Component } = {
    http: HttpRow, kafka: KafkaRow, mail: MailRow, ldap: LdapRow, config: ConfigRow, event: EventRow,
}

const { queryText, facets, result, errorMessage, loading, hasMore, searchNow, loadMore } = useSearchPanel()
const { escapeSearch, searchFocusRequest } = useDashboardLayout()

const input = ref<HTMLInputElement>()
const showTips = ref(false)
const fullSearch = computed(() => fullSearchLocation(queryText.value, facets.value))

// the text is saved, so there can be text to search for that nobody typed yet
function search() {
    if (queryText.value !== '') {
        searchNow()
    }
}

// The panel mounts every time the sidebar shows it, so it also shows the current results.
onMounted(() => {
    input.value?.focus()
    search()
})
watch(searchFocusRequest, () => nextTick(() => input.value?.focus()))
</script>

<template>
    <aside class="dashboard-sidebar search-panel" aria-label="Search">
        <div class="sidebar-filter">
            <label class="sidebar-section-title search-panel-label" for="search-panel-input">Search</label>
            <div class="sidebar-filter-field">
                <input id="search-panel-input" ref="input" v-model="queryText" type="text" class="form-control form-control-sm"
                    placeholder='Search e.g. "petstore" method:GET statusCode:>=400' autocomplete="off"
                    @keydown.enter="search" @keydown.esc="escapeSearch(queryText)" />
            </div>
        </div>

        <template v-if="queryText">
            <div v-if="result?.facets" class="search-panel-facets">
                <select v-for="name in Object.keys(result.facets)" :key="name" v-model="facets[name]"
                    class="form-select form-select-sm" :aria-label="name" @change="search">
                    <option value="">{{ facetTitle(name) }}</option>
                    <option v-for="option in result.facets[name]" :key="option.value" :value="option.value">
                        {{ option.value }} ({{ option.count }})
                    </option>
                </select>
            </div>

            <div v-if="result" class="search-panel-count">
                <span data-testid="search-count">{{ searchTotalLabel(result.results.length, result.total) }}</span>
                <button type="button" class="search-panel-action" @click="search">Refresh</button>
            </div>

            <p v-if="loading.isLoading.value" class="sidebar-hint">{{ loading.statusText.value }}</p>
            <p v-if="errorMessage" class="search-panel-error" role="alert">{{ errorMessage }}</p>
            <p v-if="result && result.total === 0 && !errorMessage" class="sidebar-hint">No results found</p>

            <ul v-if="result && result.total > 0" class="search-results" aria-label="Search results">
                <component :is="rowByType[item.params.type]" v-for="(item, index) in result.results" :key="index"
                    :item="item" />
            </ul>
            <button v-if="hasMore" type="button" class="search-panel-action search-panel-more" :disabled="loading.isLoading.value"
                @click="loadMore">Load more</button>
        </template>

        <div class="search-panel-footer">
            <button type="button" class="search-panel-action" :aria-expanded="showTips" @click="showTips = !showTips">
                Search tips <i class="bi" :class="showTips ? 'bi-chevron-up' : 'bi-chevron-down'" aria-hidden="true"></i>
            </button>
            <div v-if="showTips" class="search-panel-tips">
                <search-tips />
            </div>
            <router-link v-if="fullSearch" class="search-panel-action" :to="fullSearch">Open all results</router-link>
        </div>
    </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { isOpenResult } from '@/composables/search'
import { resultRoute } from '@/composables/search-result'
import { usePrettyDates } from '@/composables/usePrettyDate'

const props = defineProps<{
    item: SearchItem
    badge: string
    // the badge is an HTTP method, which has its own colors, otherwise a neutral type pill
    isMethod?: boolean
    status?: string
}>()

const route = useRoute()
const router = useRouter()
const { format } = usePrettyDates()

const to = computed(() => resultRoute(props.item))
const isOpen = computed(() => to.value !== undefined && isOpenResult(router.resolve(to.value).fullPath, route.fullPath))
const isError = computed(() => Number(props.status) >= 500)
const context = computed(() => [props.item.domain, props.item.time && format(props.item.time)].filter(Boolean).join(', '))
const matches = computed(() => props.item.fragments?.join(' ... '))
</script>

<template>
    <li>
        <component :is="to ? 'router-link' : 'div'" :to="to" class="search-result" :class="{ open: isOpen }"
            :aria-current="isOpen ? 'page' : undefined">
            <span class="search-result-head">
                <span class="badge" :class="isMethod ? `operation ${badge.toLowerCase()}` : 'search-type'">{{ badge }}</span>
                <span class="search-result-title" :title="item.title">{{ item.title }}</span>
                <span v-if="status" class="search-result-status" :class="isError ? 'error' : 'other'">{{ status }}</span>
            </span>
            <span class="search-result-detail">
                {{ context }}<template v-if="context && matches">, </template><span v-html="matches"></span>
            </span>
        </component>
    </li>
</template>

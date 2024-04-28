<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import type { PropType } from 'vue'
import { useRoute } from 'vue-router'

defineProps({
    actions: { type: Object as PropType<Action[]>, required: true },
})

const route = useRoute()
const { duration } = usePrettyDates()

function getName(action: Action){
    for (const key in action.tags){
        if (key == 'name'){
            return action.tags[key]
        }
    }
    return null
}
</script>

<template>
    <table class="table dataTable actions">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 30px;"></th>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">Duration</th>
            </tr>
        </thead>
        <tbody>
            <template v-for="(action, index) of actions">
                <tr data-bs-toggle="collapse" :data-bs-target="'#action_'+index" aria-expanded=false>
                    <td><i class="bi bi-chevron-right"></i><i class="bi bi-chevron-down"></i></td>
                    <td>{{ getName(action) }}</td>
                    <td>{{ duration(action.duration) }}</td>
                </tr>
                <tr class="collapse-row">
                    <td colspan="3">
                        <div class="collapse" :id="'action_'+index" style="padding: 2rem;">
                            <div class="row">
                                <div class="col-1">Tags</div>
                                <div class="col">
                            <table class="table dataTable">
                                <thead>
                                    <tr>
                                        <th scope="col" class="text-left">Name</th>
                                        <th scope="col" class="text-left w-75">Value</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-for="(value, key) of action.tags">
                                        <tr>
                                            <td>{{ key }}</td>
                                            <td>
                                                <router-link v-if="key === 'file' && action.tags.fileKey" :to="{ name: 'config', params: { id: action.tags.fileKey }, query: { refresh: route.query.refresh }}">
                                                {{ value }}
                                                </router-link>
                                                <span v-else>{{ value }}</span>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </div>
                        </div>
                        </div>
                    </td>
                </tr>
            </template>
        </tbody>
    </table>
</template>

<style scoped>
.actions tbody td{
    border-bottom-width: 0;
}
.actions tbody tr.collapse-row td{
    border-bottom-width: 2px;
    border-top-width: 0;
    background-color: var(--color-background-soft);
}
.collapse {
    padding: 2rem;
}

tr[aria-expanded=true] .bi-chevron-right{
    display:none;
}
tr[aria-expanded=false] .bi-chevron-down{
    display:none;
}
</style>
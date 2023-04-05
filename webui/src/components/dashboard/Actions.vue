<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import type { PropType } from 'vue';

defineProps({
    actions: { type: Object as PropType<Action[]>, required: true },
})

const {duration} = usePrettyDates()

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
                <tr data-bs-toggle="collapse" :data-bs-target="'#action_'+index">
                    <td><i class="bi bi-chevron-right"></i></td>
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
                                            <td>{{ value }}</td>
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
    padding: 0;
    background-color: var(--color-background-soft);
}
.collapse {
    padding: 2rem;
}
</style>
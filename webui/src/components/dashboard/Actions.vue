<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import type { PropType } from 'vue'
import { useRoute } from 'vue-router'
import SourceView from './SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'

defineProps({
    actions: { type: Object as PropType<Action[]>, required: true },
})

const { formatLanguage } = usePrettyLanguage()

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
function formatParameters(action: Action): {name?: string, value: string}[] {
    if (action.tags.event === 'http') {
        return [
            {
                name: 'request',
                value: formatLanguage(action.parameters[0], 'application/json')
            },
            {
                name: 'response',
                value: formatLanguage(action.parameters[1], 'application/json')
            }
        ]
    }
    
    let list = []
    for (let p of action.parameters) {
        list.push({ value: formatLanguage(p, 'application/json') })
    }
    return list
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
                            <div v-if="action.parameters.length > 0">
                                <h5>Event Handler Parameters</h5>
                                <div class="accordion mb-3" id="parametersAccordion">
      
                                    <div class="accordion-item" v-for="(item, index) in formatParameters(action)">
                                        <h2 class="accordion-header mt-0" :id="'param-heading-'+index">
                                        <button class="accordion-button p-2 collapsed" type="button" data-bs-toggle="collapse" :data-bs-target="'#param-'+index" aria-expanded="false" :aria-controls="'param-'+index">
                                            {{ item.name ?? index }}
                                        </button>
                                        </h2>
                                        <div :id="'param-'+index" class="accordion-collapse collapse p-0" :aria-labelledby="'param-heading-'+index" data-bs-parent="#parametersAccordion">
                                            <div class="accordion-body p-0">
                                                <source-view :source="{ preview: { content: item.value, contentType: 'application/json' }}" :hide-header="true" :max-height="250"></source-view>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                            </div>
                            <h5>Logs</h5>
                            <ul class="list-group mb-3">
                                <li v-for="log in action.logs" class="list-group-item">
                                    <span class="text-body log">
                                        <i class="bi bi-exclamation-triangle-fill text-warning" v-if="log.level == 'warn'"></i>
                                        <i class="bi bi-x-circle-fill text-danger" v-else-if="log.level == 'error'"></i>
                                        <i class="bi bi-bug-fill text-info" v-else-if="log.level == 'debug'"></i>
                                        <i class="bi bi-chat-dots text-primary" v-else></i>
                                        {{ log.message }}
                                    </span>
                                </li>
                            </ul>
                            <div class="alert alert-danger" role="alert" v-if="action.error">
                                <h5 class="alert-heading">Error</h5>
                                <p>{{ action.error?.message }}</p>
                            </div>
                            <h5>Tags</h5>
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
                    </td>
                </tr>
            </template>
        </tbody>
    </table>
</template>

<style scoped>
.actions tbody td {
    border-bottom-width: 0;
    padding-bottom: 1px;
    margin-bottom: 0;
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
.log {
    font-family: 'Fira Code', monospace;
}

.accordion-button{
    color: var(--color-text);
    border: 1px solid transparent;
}
.accordion-button::after {
    color: var(--color-text);
    filter: hue-rotate(130deg) brightness(2);
}
.accordion-button:not(.collapsed) {
    background-color: transparent;
}

.accordion-button:hover {
  border: 1px solid;
  border-color: var(--color-button-link) !important;
  background-color: transparent;
}
.accordion-button:focus {
    border-color: var(--color-button-link);
    box-shadow: none;
}
</style>
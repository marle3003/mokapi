<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import type { PropType } from 'vue'
import { useRoute } from 'vue-router'
import SourceView from './SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyHttp } from '@/composables/http'
import { usePrettyText } from '@/composables/usePrettyText'

defineProps({
    actions: { type: Object as PropType<Action[]>, required: true },
})

const { formatLanguage } = usePrettyLanguage()

const route = useRoute()
const { duration } = usePrettyDates()
const { parseUrls } = usePrettyText()
let status: string | null

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
function getStatus(action: Action) {
    if (action.error) {
        return 'error'
    }
    let hasError = false
    let hasWarning = false
    if (!action.logs) {
        return 'success'
    }
    for (const log of action.logs) {
        switch (log.level) {
            case 'error': 
                hasError = true
                break
            case 'warn':
                hasWarning = true
                break
        }
    }
    if (hasError) {
        return 'error'
    }
    if (hasWarning) {
        return 'warning'
    }
    return 'success'
}
</script>

<template>
    <table class="table dataTable actions">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 30px;"></th>
                <th scope="col" class="text-left" style="width: 30px;">Status</th>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-center" style="width: 10%">Duration</th>
            </tr>
        </thead>
        <tbody>
            <template v-for="(action, index) of actions">
                <tr data-bs-toggle="collapse" :data-bs-target="'#action_'+index" aria-expanded=false>
                    <td><i class="bi bi-chevron-right"></i><i class="bi bi-chevron-down"></i></td>
                    <td :set="status = getStatus(action)">
                        <span class="badge bg-danger me-2" v-if="status === 'error'">Error</span>
                        <span class="badge bg-warning me-2" v-else-if="status === 'warning'">Warning</span>
                        <span class="badge bg-success me-2" v-else>Success</span>
                    </td>
                    <td>{{ getName(action) }}</td>
                    <td class="text-center">{{ duration(action.duration) }}</td>
                </tr>
                <tr class="collapse-row">
                    <td colspan="4">
                        <div class="collapse" :id="'action_'+index" style="padding: 2rem;">
                            <div class="alert alert-danger" role="alert" v-if="action.error">
                                <h5 class="alert-heading">Error</h5>
                                <p>{{ action.error?.message }}</p>
                            </div>
                            <div v-if="action.parameters && action.parameters.length > 0">
                                <h5>Parameters</h5>
                                <div class="accordion mb-3" id="parametersAccordion">
      
                                    <div class="accordion-item" v-for="(item, paramIndex) in formatParameters(action)">
                                        <h2 class="accordion-header mt-0" :id="'action-'+index+'param-heading-'+paramIndex">
                                        <button class="accordion-button p-2 collapsed" type="button" data-bs-toggle="collapse" :data-bs-target="'#action-'+index+'-param-'+paramIndex" aria-expanded="false" :aria-controls="'action-'+index+'-param-'+paramIndex">
                                            {{ item.name ?? paramIndex }}
                                        </button>
                                        </h2>
                                        <div :id="'action-'+index+'-param-'+paramIndex" class="accordion-collapse collapse p-0" :aria-labelledby="'action-'+index+'-param-heading-'+paramIndex" data-bs-parent="#parametersAccordion">
                                            <div class="accordion-body p-0">
                                                <source-view :source="{ preview: { content: item.value, contentType: 'application/json' }}" :hide-header="true" :max-height="250"></source-view>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                            </div>
                            <h5>Logs</h5>
                            <div class="mb-3">
                                <ul class="list-group">
                                    <li v-for="log in action.logs" class="list-group-item">
                                        <span class="text-body log">
                                            <i class="bi bi-exclamation-triangle-fill text-warning" v-if="log.level == 'warn'"></i>
                                            <i class="bi bi-x-circle-fill text-danger" v-else-if="log.level == 'error'"></i>
                                            <i class="bi bi-bug-fill text-info" v-else-if="log.level == 'debug'"></i>
                                            <i class="bi bi-chat-dots text-primary" v-else></i>
                                            <span class="ms-2" v-html="parseUrls(log.message)"></span>
                                        </span>
                                    </li>
                                </ul>
                                <p v-if="!action.logs || action.logs.length == 0">This event handler did not produce any logs. You can use <code>console.log()</code> or <code>console.error()</code> in your script to output information.</p>
                            </div>
                            <h5>
                                Tags
                                <a href="/docs/javascript-api/mokapi/eventhandler/eventargs" class="ms-1" title="Learn how to define tags for your event handler" rel="noopener">
                                    <i class="bi bi-question-circle" aria-hidden="true"></i>
                                </a>
                            </h5>
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
h5 i {
    font-size: 0.8rem;
    vertical-align: top;
}
</style>
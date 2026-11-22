<script setup lang="ts">
import { computed, ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import LogTabs from '../../log/LogTabs.vue'
import JobSummary from '../blocks/JobSummary.vue'
import JobLogs from '../blocks/JobLogs.vue'
import JobDetails from '../blocks/JobDetails.vue'
import JobTags from '../blocks/JobTags.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { jobTabs, type JobTab } from '@/composables/job'
import { useRequestEvent } from '@/composables/request-event'

const props = defineProps<{ eventId: string }>()
defineEmits<{ close: [] }>()

// the event holds the whole run, so there is nothing else to load
const { event, data, isLoading } = useRequestEvent<JobExecution>(() => props.eventId)
const fullView = computed(() => ({ name: getRouteName('job').value, params: { id: props.eventId } }))

// The sub tab belongs to the panel, not to the run: the next run opens on the same tab.
const activeTab = ref<JobTab>('Logs')
</script>

<template>
    <log-detail-panel label="Job details" :full-view="fullView" @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <job-summary :event="event" :data="data" />
                </div>
            </div>

            <log-tabs v-model="activeTab" :tabs="jobTabs" />
            <div class="log-detail-content" role="tabpanel">
                <job-logs v-if="activeTab === 'Logs'" :data="data" />
                <div v-else class="card">
                    <div class="card-body">
                        <job-details :data="data" />
                        <job-tags :tags="data.tags" />
                    </div>
                </div>
            </div>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="Job not found" />
    </log-detail-panel>
</template>

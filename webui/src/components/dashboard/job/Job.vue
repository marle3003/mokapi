<script setup lang="ts">
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { useRoute } from '@/router'
import { useRequestEvent } from '@/composables/request-event'
import JobSummary from './blocks/JobSummary.vue'
import JobLogs from './blocks/JobLogs.vue'
import JobDetails from './blocks/JobDetails.vue'
import JobTags from './blocks/JobTags.vue'

const route = useRoute()
const { event, data, isLoading } = useRequestEvent<JobExecution>(() => route.params.id as string)
</script>

<template>
    <div v-if="event && data">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <job-summary :event="event" :data="data" />
                </div>
            </div>
        </div>
        <div class="card-group">
            <section class="card" aria-labelledby="job-logs-title">
                <div class="card-body">
                    <h2 id="job-logs-title" class="card-title text-center">Logs</h2>
                    <job-logs :data="data" />
                </div>
            </section>
        </div>
        <div class="card-group">
            <section class="card" aria-labelledby="job-details-title">
                <div class="card-body">
                    <h2 id="job-details-title" class="card-title text-center">Details</h2>
                    <job-details :data="data" />
                </div>
            </section>
        </div>
        <div class="card-group">
            <section class="card">
                <div class="card-body">
                    <job-tags :tags="data.tags" />
                </div>
            </section>
        </div>
    </div>
    <loading v-if="!event && isLoading" />
    <message v-if="!event && !isLoading" message="Job not found" />
</template>

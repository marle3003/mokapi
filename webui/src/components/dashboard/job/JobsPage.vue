<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from '@/router'
import { activeJobsTab, type JobsTab } from '@/composables/job'
import { filterByJobName } from '@/composables/job-log-filter'
import { withSelection } from '@/composables/log-selection'
import Jobs from './Jobs.vue'
import JobOverview from './JobOverview.vue'

const route = useRoute()
const router = useRouter()

const activeTab = computed(() => activeJobsTab(route.hash, route.query))

// Switching to the overview drops the selected run, as the Kafka view drops the selected message.
function setTab(tab: JobsTab) {
    router.replace({
        query: withSelection(route.query, 'job', undefined),
        hash: `#${tab}`
    })
}

function showRunsOf(name: string) {
    filterByJobName(name)
    setTab('tab-runs')
}
</script>

<template>
    <div class="card-group">
        <section class="card" aria-label="Jobs">
            <div class="card-body">
                <div class="nav card-tabs" role="tablist">
                    <button :class="{ active: activeTab === 'tab-jobs' }" id="jobs-tab" type="button" role="tab"
                        :aria-selected="activeTab === 'tab-jobs'" aria-controls="jobs-pane" @click="setTab('tab-jobs')">
                        Jobs
                    </button>
                    <button :class="{ active: activeTab === 'tab-runs' }" id="runs-tab" type="button" role="tab"
                        :aria-selected="activeTab === 'tab-runs'" aria-controls="runs-pane" @click="setTab('tab-runs')">
                        Runs
                    </button>
                </div>
                <div class="tab-content">
                    <!-- v-if: the log reloads and the countdown ticks while mounted, so a hidden one would keep running -->
                    <div v-if="activeTab === 'tab-jobs'" class="tab-pane fade show active" id="jobs-pane"
                        role="tabpanel" aria-labelledby="jobs-tab">
                        <job-overview @select="showRunsOf" />
                    </div>
                    <div v-else class="tab-pane fade show active" id="runs-pane"
                        role="tabpanel" aria-labelledby="runs-tab">
                        <div class="card-group">
                            <jobs />
                        </div>
                    </div>
                </div>
            </div>
        </section>
    </div>
</template>

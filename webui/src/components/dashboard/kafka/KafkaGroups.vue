<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { onMounted, ref } from 'vue'
import { Modal, Popover, Tab } from 'bootstrap'
import { useMetrics } from '@/composables/metrics'

const props = defineProps<{
    service: KafkaService,
    topicName?: string
}>()

const { format } = usePrettyDates()
const { sum } = useMetrics()

function memberInfo(member: KafkaMember): string {
    let addition = ''
    if (props.topicName) {
        addition = `<p id="${member.name}-partitions" class="label">Partitions</p><p aria-labelledby="${member.name}-partitions">${member.partitions[props.topicName].join(', ')}</p>`
    } else {
        addition = `<p id="${member.name}-topics" class="label">Topics</p><p aria-labelledby="${member.name}-topics">${Object.keys(member.partitions).join(',<br />')}</p>`
    }
    return `<div aria-label="${member.name}">
            <p id="${member.name}-address" class="label">Address</p>
            <p aria-labelledby="${member.name}-address">${member.addr}</p>
            <p id="${member.name}-client-software" class="label">Client Software</p>
            <p aria-labelledby="${member.name}-client-software">${member.clientSoftwareName} ${member.clientSoftwareVersion}</p>
            <p id="${member.name}-last-heartbeat" class="label">Last Heartbeat</p>
            <p aria-labelledby="${member.name}-last-heartbeat">${format(member.heartbeat)}</p>
            ${addition}
            </div>`
}

function getGroups(): KafkaGroup[] {
    if (!props.topicName) {
        return props.service.groups
    }
    if (!props.service.groups) {
        return []
    }

    let result = []
    for (let group of props.service.groups) {
        if (group.topics?.includes(props.topicName)) {
            result.push(group)
        }
    }
    return result
}

onMounted(()=> {
    const elements = document.querySelectorAll('.has-popover')
    const popovers = [...elements].map(x => {
        new Popover(x, {
            customClass: 'custom-popover',
            trigger: 'hover',
            html: true,
            placement: 'left',
            content: () => x.querySelector('span')?.innerHTML ?? '',
        })
    })
    dialog = new Modal(groupDialog.value)
    tab = new Tab(tabDetailGroup.value)
})

const groupDialog = ref<any>(null)
const tabDetailGroup = ref<any>(null)
const memberButtonList = ref<any>(null)
let dialog:  Modal
let tab: Tab
let selectedGroup = ref<KafkaGroup | null>(null)
function showGroup(group: KafkaGroup){
    if (getSelection()?.toString()) {
        return
    }

    selectedGroup.value = group
    tab.show()
    dialog.show()
    if (group.members.length > 0 && memberButtonList.value) {
        new Tab(memberButtonList.value.children[0]).show()
    }
}
</script>

<template>
    <table class="table dataTable selectable">
        <caption class="visually-hidden">{{ props.topicName ? 'Topic Groups' : 'Cluster Groups' }}</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left">Name</th>
                <th scope="col" class="text-left">State</th>
                <th scope="col" class="text-left">Protocol</th>
                <th scope="col" class="text-left">Coordinator</th>
                <th scope="col" class="text-left">Leader</th>
                <th scope="col" class="text-left">Members</th>
                <th scope="col" class="text-center" v-if="topicName">Lag</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="group in getGroups()" :key="group.name" @click="showGroup(group)">
                <td>{{ group.name }}</td>
                <td>{{ group.state }}</td>
                <td>{{ group.protocol }}</td>
                <td>{{ group.coordinator }}</td>
                <td>{{ group.leader }}</td>
                <td>
                    <ul class="members">
                        <li v-for="member in group.members" class="has-popover">
                            {{ member.name }} <i class="bi bi-info-circle"></i>
                            <span style="display:none" v-html="memberInfo(member)"></span>
                        </li>
                        
                    </ul>
                </td>
                <td v-if="topicName" class="text-center">
                    {{ sum(service.metrics, 'kafka_consumer_group_lag', { name: 'topic', value: topicName }, { name: 'group', value: group.name }) }}
                </td>
            </tr>
        </tbody>
    </table>
    <div class="modal fade" id="dialogGroup" ref="groupDialog" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <div class="card-group" >
                        <div class="card">
                            <div class="card-body">
                                <div class="row">
                                    <ul class="nav nav-pills tab-sm mb-3" role="tablist">
                                        <li class="nav-link show active" style="padding-left: 12px;" ref="tabDetailGroup" id="detail-group-tab" data-bs-toggle="tab" data-bs-target="#detail-group" type="button" role="tab" aria-controls="detail-group" aria-selected="true">Group</li>
                                        <li class="nav-link" id="detail-topics-tab" data-bs-toggle="tab" data-bs-target="#detail-topics" type="button" role="tab" aria-controls="detail-topics" aria-selected="false">Topics</li>
                                        <li class="nav-link" id="detail-members-tab" data-bs-toggle="tab" data-bs-target="#detail-members" type="button" role="tab" aria-controls="detail-members" aria-selected="false">Members</li>
                                    </ul>

                                    <div class="tab-content" v-if="selectedGroup">
                                        <div class="tab-pane fade show active" id="detail-group" role="tabpanel">
                                            <div class="row mb-3">
                                                <p id="dialog-group-name" class="label">Name</p>
                                                    <p aria-labelledby="dialog-group-name">{{ selectedGroup.name }}</p>
                                            </div>
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-group-state" class="label">State</p>
                                                    <p aria-labelledby="dialog-group-statet">{{ selectedGroup.state }}</p>
                                                </div>
                                                <div class="col">
                                                    <p id="dialog-group-protocol" class="label">Protocol</p>
                                                    <p aria-labelledby="dialog-group-protocol">{{ selectedGroup.protocol }}</p>
                                                </div>
                                            </div>
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-group-coordinator" class="label">Coordinator</p>
                                                    <p aria-labelledby="dialog-group-coordinator">{{ selectedGroup.coordinator }}</p>
                                                </div>
                                                <div class="col">
                                                    <p id="dialog-group-leader" class="label">Leader</p>
                                                    <p aria-labelledby="dialog-group-leader">{{ selectedGroup.leader }}</p>
                                                </div>
                                            </div>
                                        </div>
                                        <div class="tab-pane fade" id="detail-topics" role="tabpanel">
                                            <table class="table dataTable">
                                                <caption class="visually-hidden">Topics</caption>
                                                <thead>
                                                    <tr>
                                                        <th scope="col" class="text-left">Topic</th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    <tr v-for="topicName of selectedGroup.topics" :key="topicName">
                                                        <td>{{ topicName }}</td>
                                                    </tr>
                                                </tbody>
                                            </table>
                                        </div>
                                        <div class="tab-pane fade members" id="detail-members" role="tabpanel">
                                            <div class="d-flex align-items-start align-items-stretch">
                                                <div class="nav flex-column nav-pills" id="v-pills-tab" role="tablist" aria-orientation="vertical" ref="memberButtonList">
                                                    <button v-for="(member, index) of selectedGroup.members" class="badge member" :class="(index==0 ? ' active' : '')" :id="'v-pills-'+member.name+'-tab'" data-bs-toggle="pill" :data-bs-target="'#v-pills-'+member.name" type="button" role="tab" :aria-controls="'v-pills-'+member.name" :aria-selected="index === 0">
                                                        {{ member.name }}
                                                        <i class="bi bi-stars" v-if="member.name === selectedGroup.leader"></i>
                                                    </button>
                                                </div>
                                                <div class="tab-content ms-3 ps-3 members-tab" style="width: 100%" id="v-pills-tabContent">
                                                    <div v-for="(member, index) of selectedGroup.members" class="tab-pane fade" :class="index==0 ? 'show active' : ''" :id="'v-pills-'+member.name" role="tabpanel" :aria-labelledby="'v-pills-'+member.name+'-tab'">
                                                        <div class="row mb-3">
                                                            <div class="col">
                                                                <p id="dialog-group-member-addr" class="label">Address</p>
                                                                <p aria-labelledby="dialog-group-member-addr">{{ member.addr }}</p>
                                                            </div>
                                                            <div class="col">
                                                                <p id="dialog-group-member-client-sw" class="label">Client Software</p>
                                                                <p aria-labelledby="dialog-group-member-client-sw">{{ `${member.clientSoftwareName} ${member.clientSoftwareVersion}` }}</p>
                                                            </div>
                                                        </div>
                                                        <div class="row mb-3">
                                                            <div class="col">
                                                                <p id="dialog-group-member-heartbeat" class="label">Heartbeat</p>
                                                                <p aria-labelledby="dialog-group-member-heartbeat">{{ format(member.heartbeat) }}</p>
                                                            </div>
                                                        </div>
                                                        <table class="table dataTable">
                                                            <caption class="visually-hidden">Member Partitions</caption>
                                                            <thead>
                                                                <tr>
                                                                    <th scope="col" class="text-left">Topic</th>
                                                                    <th scope="col" class="text-left">Partitions</th>
                                                                </tr>
                                                            </thead>
                                                            <tbody>
                                                                <tr v-for="(partition, topicName) of member.partitions" :key="topicName">
                                                                    <td>{{ topicName }}</td>
                                                                    <td>{{ partition.join(', ') }}</td>
                                                                </tr>
                                                            </tbody>
                                                        </table>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
ul.members {
    list-style: none; 
    padding: 0;
    margin: 0;
}
ul.members li {
    padding-right: 0.5em;
}
.tab-pane {
    padding: 0;
}
.members .tab-content.members-tab {
    border-style: solid;
    border-width: 0;
    border-left-width: 2px;
    border-color: var(--color-datatable-border);
    min-width: 0;
}
.members .nav.nav-pills button.badge {
    font-size: 0.75rem;
    border-color: var(--color-datatable-border);
    border: 0;
    margin-bottom: 15px;
}
.members .nav.nav-pills button.badge.active {
    font-size: 0.75rem;
    border: 0;
    outline-width: 2px;
    outline-style: solid;
}
.members .tab-pane {
    padding: 0;
    margin-top: -5px;
}
</style>
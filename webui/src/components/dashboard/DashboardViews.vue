<script setup lang="ts">
import AppStartCard from './AppStartCard.vue'
import MemoryUsageCard from './MemoryCard.vue'
import JobCountCard from '@/components/dashboard/JobCountCard.vue'
import Search from './Search.vue'

import HttpRequestCard from './http/HttpRequestMetricCard.vue'
import HttpServicesCard from './http/HttpServicesCard.vue'
import HttpService from './http/HttpService.vue'
import HttpRequests from './http/Requests.vue'

import KafkaMessageMetricCard from './kafka/KafkaMessageMetricCard.vue'
import KafkaClustersCard from './kafka/KafkaServicesCard.vue'
import KafkaService from './kafka/KafkaService.vue'
import KafkaMessages from '@/components/dashboard/kafka/KafkaMessages.vue'

import LdapServicesCard from '@/components/dashboard/ldap/LdapServicesCard.vue'
import LdapService from './ldap/Service.vue'
import LdapSearchMetricCard from './ldap/LdapRequestMetricCard.vue'
import LdapRequests from '@/components/dashboard/ldap/Requests.vue'

import SmtpMessageMetricCard from './mail/SmtpMessageMetricCard.vue'
import MailServicesCard from './mail/MailServicesCard.vue'
import MailService from './mail/Service.vue'
import Mails from '@/components/dashboard/mail/Mails.vue'

import MqttMessageMetricCard from './mqtt/MqttMessageMetricCard.vue'
import MqttClustersCard from './mqtt/MqttServicesCard.vue'
import MqttService from './mqtt/MqttService.vue'
import MqttMessages from '@/components/dashboard/mqtt/MqttMessages.vue'

import WebsocketMessageMetricCard from './websocket/WebsocketMessageMetricCard.vue'
import WebsocketServicesCard from './websocket/WebsocketServicesCard.vue'
import WebsocketMessages from '@/components/dashboard/websocket/WebsocketMessages.vue'

import Message from '@/components/Message.vue'

import Jobs from '@/components/dashboard/job/Jobs.vue'
import Job from '@/components/dashboard/job/Job.vue'
import ConfigCard from '@/components/dashboard/ConfigCard.vue'
import FakerTree from '@/components/dashboard/FakerTree.vue'

import Config from '@/components/dashboard/Config.vue'
import WebsocketService from '@/components/dashboard/websocket/WebsocketService.vue'

import { getRouteName } from '@/composables/dashboard'
import type { AppInfo } from '@/types/dashboard'
import type { PropType } from 'vue'

const props = defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
})

function isServiceAvailable(service: string): boolean {
    if (!props.appInfo.activeServices) {
        return false
    }
    return props.appInfo.activeServices.includes(service)
}

function isAnyServiceAvailable() {
    return props.appInfo.activeServices && props.appInfo.activeServices.length > 0
}
</script>

<template>
    <div v-if="$route.name === getRouteName('dashboard').value">
        <div class="card-group">
            <app-start-card />
            <memory-usage-card />
        </div>
        <div class="card-group">
            <http-request-card v-if="isServiceAvailable('http')" includeError />
            <kafka-message-metric-card v-if="isServiceAvailable('kafka')" />
            <mqtt-message-metric-card v-if="isServiceAvailable('mqtt')" />
            <websocket-message-metric-card v-if="isServiceAvailable('websocket')" />
            <smtp-message-metric-card v-if="isServiceAvailable('mail')" />
            <ldap-search-metric-card v-if="isServiceAvailable('ldap')" />
            <job-count-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('http')">
            <http-services-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('kafka')">
            <kafka-clusters-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('mqtt')">
            <mqtt-clusters-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('websocket')">
            <websocket-services-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('mail')">
            <mail-services-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('ldap')">
            <ldap-services-card />
        </div>

        <message message="No service available" v-if="!isAnyServiceAvailable()"></message>
    </div>

    <div v-if="$route.name === getRouteName('http').value">
        <div class="card-group">
            <http-request-card v-if="isServiceAvailable('http')" />
            <http-request-card v-if="isServiceAvailable('http')" onlyError />
        </div>
        <div class="card-group" v-if="isServiceAvailable('http')">
            <http-services-card />
        </div>
        <div class="card-group">
            <http-requests />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('httpLog').value">
        <div class="card-group">
            <http-requests />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('kafkaLog').value">
        <div class="card-group">
            <kafka-messages />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('kafka').value">
        <div class="card-group">
            <kafka-message-metric-card v-if="isServiceAvailable('kafka')" />
        </div>
        <div class="card-group" v-if="isServiceAvailable('kafka')">
            <kafka-clusters-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('kafka')">
            <kafka-messages />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('mqttLog').value">
        <div class="card-group">
            <mqtt-messages />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('mqtt').value">
        <div class="card-group">
            <mqtt-message-metric-card v-if="isServiceAvailable('mqtt')" />
        </div>
        <div class="card-group" v-if="isServiceAvailable('mqtt')">
            <mqtt-clusters-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('mqtt')">
            <mqtt-messages />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('websocketLog').value">
        <div class="card-group">
            <websocket-messages />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('websocket').value">
        <div class="card-group">
            <websocket-message-metric-card v-if="isServiceAvailable('websocket')" />
        </div>
        <div class="card-group" v-if="isServiceAvailable('websocket')">
            <websocket-services-card />
        </div>
        <div class="card-group" v-if="isServiceAvailable('websocket')">
            <websocket-messages />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('mail').value">
        <div class="card-group">
            <smtp-message-metric-card v-if="isServiceAvailable('mail')" />
        </div>
        <div class="card-group" v-if="isServiceAvailable('mail')">
            <mail-services-card />
        </div>
        <div class="card-group">
            <mails />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('mailLog').value">
        <div class="card-group">
            <mails />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('ldapLog').value">
        <div class="card-group">
            <ldap-requests />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('ldap').value">
        <div class="card-group">
            <ldap-search-metric-card v-if="isServiceAvailable('ldap')" />
        </div>
        <div class="card-group" v-if="isServiceAvailable('ldap')">
            <ldap-services-card />
        </div>
        <div class="card-group">
            <ldap-requests />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('jobs').value">
        <div class="card-group">
            <jobs />
        </div>
    </div>

    <div v-if="$route.name === getRouteName('job').value">
        <job />
    </div>

    <div v-if="$route.name === getRouteName('configs').value">
        <div class="card-group">
            <config-card />
        </div>
    </div>
    <div v-if="$route.name === getRouteName('tree').value">
        <div class="card-group">
            <faker-tree v-if="$route.name === 'tree'"></faker-tree>
        </div>
    </div>

    <div v-if="$route.name === getRouteName('search').value">
        <search v-if="$route.name === 'search'"></search>
    </div>

    <http-service v-if="$route.meta.service === 'http'" />
    <kafka-service v-if="$route.meta.service === 'kafka'" />
    <mqtt-service v-if="$route.meta.service === 'mqtt'" />
    <websocket-service v-if="$route.meta.service === 'websocket'" />
    <mail-service v-if="$route.meta.service === 'mail'" />
    <ldap-service v-if="$route.meta.service === 'ldap'" />
    <config v-if="$route.name === getRouteName('config').value"></config>
</template>

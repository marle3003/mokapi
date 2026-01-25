<script setup lang="ts">
import AppStartCard from '../components/dashboard/AppStartCard.vue'
import MemoryUsageCard from '../components/dashboard/MemoryCard.vue'
import JobCountCard from '@/components/dashboard/JobCountCard.vue'
import Search from '../components/dashboard/Search.vue'

import HttpRequestCard from '../components/dashboard/http/HttpRequestMetricCard.vue'
import HttpServicesCard from '../components/dashboard/http/HttpServicesCard.vue'
import HttpService from '../components/dashboard/http/HttpService.vue'
import HttpRequests from '../components/dashboard/http/Requests.vue'

import KafkaMessageMetricCard from '../components/dashboard/kafka/KafkaMessageMetricCard.vue'
import KafkaClustersCard from '../components/dashboard/kafka/KafkaServicesCard.vue'
import KafkaService from '../components/dashboard/kafka/KafkaService.vue'
import KafkaMessagesCard from '@/components/dashboard/kafka/KafkaMessagesCard.vue'

import LdapServicesCard from '@/components/dashboard/ldap/LdapServicesCard.vue'
import LdapService from '../components/dashboard/ldap/Service.vue'
import LdapSearchMetricCard from '../components/dashboard/ldap/LdapRequestMetricCard.vue'
import Searches from '@/components/dashboard/ldap/Requests.vue'

import SmtpMessageMetricCard from '../components/dashboard/mail/SmtpMessageMetricCard.vue'
import MailServicesCard from '../components/dashboard/mail/MailServicesCard.vue'
import MailService from '../components/dashboard/mail/Service.vue'
import Mails from '@/components/dashboard/mail/Mails.vue'

import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'

import JobCard from '@/components/dashboard/JobCard.vue'
import ConfigCard from '@/components/dashboard/ConfigCard.vue'
import FakerTree from '@/components/dashboard/FakerTree.vue'

import '@/assets/dashboard.css'
import { computed, onMounted, onUnmounted, ref, type Ref } from 'vue'

import { useMeta } from '@/composables/meta'
import Config from '@/components/dashboard/Config.vue'
import { useRoute } from 'vue-router'
import { useRefreshManager } from '@/composables/refresh-manager'
import { useDashboard, getRouteName } from '@/composables/dashboard'
import type { AppInfoResponse } from '@/types/dashboard'
import Tabs from '@/components/dashboard/Tabs.vue'

const { progress, start, isActive } = useRefreshManager();
const transitionRefresh = computed(() => {
    return progress.value > 0.1 && progress.value < 99.9;
})
const { dashboard, setMode, getMode } = useDashboard();

const route = useRoute()
const appInfo = ref<AppInfoResponse | undefined>()
onMounted(() => {
    if (route.meta.mode === 'live') {
        start();
    }
    if (route.meta.mode) {
        setMode(route.meta.mode as 'live' | 'demo')
    }
    appInfo.value = dashboard.value.getAppInfo()
    onUnmounted(() => {
        appInfo.value?.close()
    })
})

function isServiceAvailable(service: string): boolean{
    if (!appInfo.value?.data || !appInfo.value?.data.activeServices){
        return false
    }
    return appInfo.value?.data.activeServices.includes(service)
}

function isAnyServiceAvailable() {
    return appInfo.value?.data.activeServices && appInfo.value?.data.activeServices.length> 0
}

function isInitLoading() {
    return appInfo.value?.isLoading && !appInfo.value?.data
}

const description = `Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mock APIs are used.`
useMeta('Dashboard | mokapi.io', description, '')
</script>

<template>
    <div v-if="isActive" class="position-fixed top-0 start-0 w-100" style="height: 2px; z-index: 1000;">
        <div class="refresh-progress-bar h-100" :style="{
                width: progress + '%',
                height: '3px',
                transition: transitionRefresh ? 'width 0.12s linear' : 'none'
            }">
        </div>
    </div>
    <main>
        <div class="dashboard">
            <h1 v-if="getMode() === 'live'" class="visually-hidden">Dashboard</h1>
            <div v-if="getMode() === 'demo'" class="header-demo">
                <h1 style="font-size: 2rem; margin-bottom: 10px;">Demo Dashboard</h1>
                <p>
                Get a feel for the interface and explore recorded data.
                </p>
            </div>

            <div class="dashboard-tabs" v-if="appInfo?.data" :class="{ demo: getMode() === 'demo' }">
                <tabs :app-info="appInfo.data" />
            </div>
            <div v-if="appInfo?.data" class="dashboard-content" :class="{ demo: getMode() === 'demo' }">
                <div v-if="$route.name === getRouteName('dashboard').value && appInfo?.data">
                    <div class="card-group">
                        <app-start-card />
                        <memory-usage-card />
                    </div>
                    <div class="card-group">
                        <http-request-card v-if="isServiceAvailable('http')" includeError />
                        <kafka-message-metric-card v-if="isServiceAvailable('kafka')" />
                        <smtp-message-metric-card v-if="isServiceAvailable('mail')" />
                        <ldap-search-metric-card v-if="isServiceAvailable('ldap')" />
                        <job-count-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('http')">
                        <http-services-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('kafka')">
                        <kafka-clusters-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('mail')">
                        <mail-services-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('ldap')">
                        <ldap-services-card />
                    </div>

                    <message message="No service available" v-if="appInfo.data && !isAnyServiceAvailable()"></message>
                </div>

                <div v-if="$route.name === getRouteName('http').value">
                    <div class="card-group">
                        <http-request-card v-if="isServiceAvailable('http')" />
                        <http-request-card v-if="isServiceAvailable('http')" onlyError />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('http')">
                        <http-services-card />
                    </div>
                    <div class="card-group">
                        <http-requests />
                    </div>
                </div>

                <div v-if="$route.name === getRouteName('kafka').value">
                    <div class="card-group">
                        <kafka-message-metric-card v-if="isServiceAvailable('kafka')" />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('kafka')">
                        <kafka-clusters-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('kafka')">
                        <kafka-messages-card />
                    </div>
                </div>

                <div v-if="$route.name === getRouteName('mail').value">
                    <div class="card-group">
                        <smtp-message-metric-card v-if="isServiceAvailable('mail')" />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('mail')">
                        <mail-services-card />
                    </div>
                    <div class="card-group">
                        <mails />
                    </div>
                </div>

                <div v-if="$route.name === getRouteName('ldap').value">
                    <div class="card-group">
                        <ldap-search-metric-card v-if="isServiceAvailable('ldap')" />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('ldap')">
                        <ldap-services-card />
                    </div>
                    <div class="card-group">
                        <searches />
                    </div>
                </div>

                <div v-if="$route.name === getRouteName('jobs').value">
                    <div class="card-group">
                        <job-card />
                    </div>
                </div>

                <div v-if="$route.name === getRouteName('configs').value">
                    <div class="card-group">
                        <config-card />
                    </div>
                </div>
                <div v-if="$route.name === getRouteName('tree').value" >
                    <div class="card-group">
                        <faker-tree v-if="$route.name === 'tree'"></faker-tree>
                    </div>
                </div>

                <div v-if="$route.name === getRouteName('search').value" >
                    <search v-if="$route.name === 'search'"></search>
                </div>

                <http-service v-if="$route.meta.service === 'http'" />
                <kafka-service v-if="$route.meta.service === 'kafka'" />
                <mail-service v-if="$route.meta.service === 'mail'" />
                <ldap-service v-if="$route.meta.service === 'ldap'" />
                <config v-if="$route.name === getRouteName('config').value"></config>
            </div>
        </div>

        <message :message="appInfo.error" v-if="!appInfo?.data && !appInfo?.isLoading && appInfo?.error"></message>
        <loading v-if="isInitLoading()"></loading>
    </main>
</template>

<style scoped>
.refresh-progress-bar {
  box-shadow: 0 0 4px rgba(0, 123, 255, 0.4);
  background-color: var(--color-refresh-background);
}
</style>
<script setup lang="ts">
import { useAppInfo } from '@/composables/appInfo'
import AppStartCard from '../components/dashboard/AppStartCard.vue'
import MemoryUsageCard from '../components/dashboard/MemoryCard.vue'
import MetricCard from '../components/dashboard/MetricCard.vue'
import Search from '../components/dashboard/Search.vue'

import HttpRequestCard from '../components/dashboard/http/HttpRequestMetricCard.vue'
import HttpServicesCard from '../components/dashboard/http/HttpServicesCard.vue'
import HttpService from '../components/dashboard/http/HttpService.vue'
import HttpRequests from '../components/dashboard/http/Requests.vue'

import KafkaMessageCard from '../components/dashboard/KafkaMessageMetricCard.vue'
import KafkaClustersCard from '../components/dashboard/kafka/KafkaServicesCard.vue'
import KafkaService from '../components/dashboard/kafka/KafkaService.vue'

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
import { onMounted, onUnmounted, ref, watchEffect, type Ref } from 'vue'

import { useMeta } from '@/composables/meta'
import Config from '@/components/dashboard/Config.vue'
import { useFetch, type Response } from '@/composables/fetch'
import { useRoute } from 'vue-router'
import { useMetrics } from '../composables/metrics'

const appInfo = useAppInfo()
const configs: Ref<Response | undefined> = ref<Response>()

const { query, sum } = useMetrics()
const count = ref<number>(0)
const response = query('app')
watchEffect(() => {
    count.value = 0
    if (!response.data) {
        return
    }
    
    const n = sum(response.data, 'app_job_run_total')
    count.value = n
})

onUnmounted(() => {
    appInfo.close()
    if (configs.value) {
        configs.value.close()
    }
    response.close()
})

const route = useRoute()
onMounted(() => {
    if (route.name === 'configs' && configs.value == null) {
        configs.value = useFetch('/api/configs')
    }
})

function isServiceAvailable(service: string): boolean{
    if (!appInfo.data || !appInfo.data.activeServices){
        return false
    }
    return appInfo.data.activeServices.includes(service)
}

function isAnyServiceAvailable() {
    return appInfo.data.activeServices && appInfo.data.activeServices.length> 0
}

function isInitLoading() {
    return appInfo.isLoading && !appInfo.data
}

function hasJobs() {
    return count.value > 0
}

const description = `Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mock APIs are used.`
useMeta('Dashboard | mokapi.io', description, '')
</script>

<template>
    <main>
        <div class="dashboard">
            <h1 class="visually-hidden">Dashboard</h1>
            <div class="dashboard-tabs" v-if="appInfo.data">
                <nav class="navbar navbar-expand-lg pb-1" aria-label="Services">
                    <div>
                        <ul class="navbar-nav me-auto mb-2 mb-lg-0">
                            <li class="nav-item">
                                <router-link class="nav-link overview" :to="{ name: 'dashboard', query: {refresh: $route.query.refresh} }">Overview</router-link>
                            </li>
                            <li class="nav-item" v-if="isServiceAvailable('http')">
                                <router-link class="nav-link" :to="{ name: 'http', query: {refresh: $route.query.refresh} }">HTTP</router-link>
                            </li>
                            <li class="nav-item" v-if="isServiceAvailable('kafka')">
                                <router-link class="nav-link" :to="{ name: 'kafka', query: {refresh: $route.query.refresh} }">Kafka</router-link>
                            </li>
                            <li class="nav-item" v-if="isServiceAvailable('mail')">
                                <router-link class="nav-link" :to="{ name: 'mail', query: {refresh: $route.query.refresh} }">Mail</router-link>
                            </li>
                            <li class="nav-item" v-if="isServiceAvailable('ldap')">
                                <router-link class="nav-link" :to="{ name: 'ldap', query: {refresh: $route.query.refresh} }">LDAP</router-link>
                            </li>
                            <li class="nav-item" v-if="hasJobs()">
                                <router-link class="nav-link" :to="{ name: 'jobs', query: {refresh: $route.query.refresh} }">Jobs</router-link>
                            </li>
                            <li class="nav-item">
                                <router-link class="nav-link" :to="{ name: 'configs', query: {refresh: $route.query.refresh} }">Configs</router-link>
                            </li>
                            <li class="nav-item">
                                <router-link class="nav-link" :to="{ name: 'tree', query: {refresh: $route.query.refresh} }">Faker</router-link>
                            </li>
                            <li class="nav-item" v-if="appInfo.data.search.enabled">
                                <router-link class="nav-link" :to="{ name: 'search', query: {refresh: $route.query.refresh} }">Search</router-link>
                            </li>
                        </ul>
                    </div>
                </nav>
            </div>
            <div v-if="appInfo.data" class="dashboard-content">
                <div v-if="$route.name === 'dashboard' && appInfo.data">
                    <div class="card-group">
                        <app-start-card />
                        <memory-usage-card />
                    </div>
                    <div class="card-group">
                        <http-request-card v-if="isServiceAvailable('http')" includeError />
                        <kafka-message-card v-if="isServiceAvailable('kafka')" />
                        <smtp-message-metric-card v-if="isServiceAvailable('mail')" />
                        <ldap-search-metric-card v-if="isServiceAvailable('ldap')" />
                        <metric-card title="Jobs" :value="count" data-testid="metric-job-count" v-if="count > 0"></metric-card>
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

                <div v-if="$route.name === 'http'">
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

                <div v-if="$route.name === 'kafka'">
                    <div class="card-group">
                        <kafka-message-card v-if="isServiceAvailable('kafka')" />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('kafka')">
                        <kafka-clusters-card />
                    </div>
                </div>

                <div v-if="$route.name === 'mail'">
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

                <div v-if="$route.name === 'ldap'">
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

                <div v-if="$route.name === 'jobs'">
                    <div class="card-group">
                        <job-card />
                    </div>
                </div>

                <div v-if="$route.name === 'configs'">
                    <div class="card-group">
                        <config-card v-if="configs" :configs="configs.data" />
                    </div>
                </div>
                <div v-if="$route.name === 'tree'" >
                    <div class="card-group">
                        <faker-tree v-if="$route.name === 'tree'"></faker-tree>
                    </div>
                </div>

                <div v-if="$route.name === 'search'" >
                    <search v-if="$route.name === 'search'"></search>
                </div>

                <http-service v-if="$route.meta.service === 'http'" />
                <kafka-service v-if="$route.meta.service === 'kafka'" />
                <mail-service v-if="$route.meta.service === 'mail'" />
                <ldap-service v-if="$route.meta.service === 'ldap'" />
                <config v-if="$route.name === 'config'"></config>
            </div>
        </div>

        <message :message="appInfo.error" v-if="!appInfo.data && !appInfo.isLoading"></message>
        <loading v-if="isInitLoading()"></loading>
    </main>
</template>
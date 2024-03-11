<script setup lang="ts">
import { useAppInfo } from '@/composables/appInfo'
import AppStartCard from '../components/dashboard/AppStartCard.vue'
import MemoryUsageCard from '../components/dashboard/MemoryCard.vue'

import HttpRequestCard from '../components/dashboard/http/HttpRequestMetricCard.vue'
import HttpServicesCard from '../components/dashboard/http/HttpServicesCard.vue'
import HttpService from '../components/dashboard/http/HttpService.vue'
import HttpRequests from '../components/dashboard/http/Requests.vue'

import KafkaMessageCard from '../components/dashboard/KafkaMessageMetricCard.vue'
import KafkaClustersCard from '../components/dashboard/kafka/KafkaServicesCard.vue'
import KafkaService from '../components/dashboard/kafka/KafkaService.vue'

import LdapServicesCard from '@/components/dashboard/ldap/LdapServicesCard.vue'
import LdapService from '../components/dashboard/ldap/Service.vue'
import LdapSearchMetricCard from '../components/dashboard/ldap/LdapSearchMetricCard.vue'
import Searches from '@/components/dashboard/ldap/Searches.vue'

import SmtpMessageMetricCard from '../components/dashboard/smtp/SmtpMessageMetricCard.vue'
import SmtpServicesCard from '../components/dashboard/smtp/SmtpServicesCard.vue'
import SmtpService from '../components/dashboard/smtp/Service.vue'
import Mails from '@/components/dashboard/smtp/Mails.vue'

import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'

import ConfigCard from '@/components/dashboard/ConfigCard.vue'

import '@/assets/dashboard.css'
import { onMounted, onUnmounted, watch, ref, type Ref } from 'vue'

import { useMeta } from '@/composables/meta'
import Config from '@/components/dashboard/Config.vue'
import { useFetch, type Response } from '@/composables/fetch'
import { useRoute } from 'vue-router'

const appInfo = useAppInfo()
const configs: Ref<Response | undefined> = ref<Response>()
onUnmounted(() => {
    appInfo.close()
    if (configs.value) {
        configs.value.close()
    }
})

const route = useRoute()
onMounted(() => {
    if (route.name === 'configs' && configs.value == null) {
        configs.value = useFetch('/api/configs')
    }
})

function isServiceAvailable(service: string): Boolean{
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

const description = `Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mock APIs are used.`
useMeta('Dashboard | mokapi.io', description, "https://mokapi.io/smtp")
</script>

<template>
    <main>
        <div class="dashboard">
            <h1 class="visually-hidden">Dashboard</h1>
            <div class="dashboard-tabs" v-if="appInfo.data">
                <nav class="navbar navbar-expand-lg" aria-label="Services">
                    <ul class="navbar-nav me-auto mb-2 mb-lg-0">
                        <li class="nav-item">
                            <router-link class="nav-link overview" :to="{ name: 'dashboard', query: {refresh: $route.query.refresh} }">Overview</router-link>
                        </li>
                        <li class="nav-item">
                            <router-link class="nav-link" :to="{ name: 'http', query: {refresh: $route.query.refresh} }" v-if="isServiceAvailable('http')">HTTP</router-link>
                        </li>
                        <li class="nav-item">
                            <router-link class="nav-link" :to="{ name: 'kafka', query: {refresh: $route.query.refresh} }" v-if="isServiceAvailable('kafka')">Kafka</router-link>
                        </li>
                        <li class="nav-item">
                            <router-link class="nav-link" :to="{ name: 'smtp', query: {refresh: $route.query.refresh} }" v-if="isServiceAvailable('smtp')">SMTP</router-link>
                        </li>
                        <li class="nav-item">
                            <router-link class="nav-link" :to="{ name: 'ldap', query: {refresh: $route.query.refresh} }" v-if="isServiceAvailable('ldap')">LDAP</router-link>
                        </li>
                        <li class="nav-item">
                            <router-link class="nav-link" :to="{ name: 'configs', query: {refresh: $route.query.refresh} }">Configs</router-link>
                        </li>
                    </ul>
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
                        <smtp-message-metric-card v-if="isServiceAvailable('smtp')" />
                        <ldap-search-metric-card v-if="isServiceAvailable('ldap')" />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('http')">
                        <http-services-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('kafka')">
                        <kafka-clusters-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('smtp')">
                        <smtp-services-card />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('ldap')">
                        <ldap-services-card />
                    </div>
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

                <div v-if="$route.name === 'smtp'">
                    <div class="card-group">
                        <smtp-message-metric-card v-if="isServiceAvailable('smtp')" />
                    </div>
                    <div class="card-group"  v-if="isServiceAvailable('smtp')">
                        <smtp-services-card />
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

                <div v-if="$route.name === 'configs'">
                    <div class="card-group">
                        <config-card v-if="configs" :configs="configs.data" />
                    </div>
                </div>

                <http-service v-if="$route.meta.service === 'http'" />
                <kafka-service v-if="$route.meta.service === 'kafka'" />
                <smtp-service v-if="$route.meta.service === 'smtp'" />
                <ldap-service v-if="$route.meta.service === 'ldap'" />
                <config v-if="$route.name === 'config'"></config>
            </div>
        </div>

        <message :message="appInfo.error" v-if="!appInfo.data && !appInfo.isLoading"></message>
        <message message="No service available" v-if="appInfo.data && !isAnyServiceAvailable()"></message>
        <loading v-if="isInitLoading()"></loading>
    </main>
</template>

<style scoped>
.router-link-active.overview {
    background-color: transparent;
    color: var(--color-text);
}
.router-link-exact-active.overview {
    background-color: var(--color-background-mute);
    color: var(--color-nav-link-active);
}
</style>
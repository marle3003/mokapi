<script setup lang="ts">
import { useAppInfo } from '@/composables/appInfo';
import AppStartCard from '../components/dashboard/AppStartCard.vue'
import MemoryUsageCard from '../components/dashboard/MemoryCard.vue'

import HttpRequestCard from '../components/dashboard/http/HttpRequestMetricCard.vue'
import HttpServicesCard from '../components/dashboard//http/HttpServicesCard.vue'
import HttpService from '../components/dashboard//http/HttpService.vue'
import HttpRequests from '../components/dashboard/http/Requests.vue'

import KafkaMessageCard from '../components/dashboard/KafkaMessageCard.vue'
import SmtpMailCard from '../components/dashboard/SmtpMailCard.vue'

import '@/assets/dashboard.css'

const appInfo = useAppInfo()

function isServiceAvailable(service: string): Boolean{
    if (!appInfo.data){
        return false
    }
    return appInfo.data.activeServices.includes(service)
}
</script>

<template>
    <div class="dashboard">
        <div class="dashboard-tabs">
            <nav class="navbar navbar-expand-lg">
                <ul class="navbar-nav me-auto mb-2 mb-lg-0">
                    <li class="nav-item">
                        <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: $route.query.refresh} }">Overview</router-link>
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
                </ul>
            </nav>
        </div>
        <div v-if="$route.name == 'dashboard'">
            <div class="card-group">
                <app-start-card />
                <memory-usage-card />
            </div>
            <div class="card-group">
                <http-request-card v-if="isServiceAvailable('http')" includeError />
                <kafka-message-card v-if="isServiceAvailable('kafka')" />
                <smtp-mail-card v-if="isServiceAvailable('smtp')" />
            </div>
            <div class="card-group"  v-if="isServiceAvailable('http')">
                <http-services-card />
            </div>
        </div>

        <div v-if="$route.name == 'http'">
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

        <http-service v-if="$route.meta.service == 'http'" />
    </div>
</template>

<style scoped>
.dashboard {
    width: 90%;
    margin: 0 auto auto;
}
.dashboard-tabs .nav-link {
  color: var(--color-text);
  position: relative;
  border-radius: 6px;
  margin-right: 5px;
}
.dashboard-tabs .nav-link:hover {
  color: var(--color-text);
  text-decoration: none;
  background-color: var(--color-background-mute);
  opacity: 0.8;
}
.dashboard-tabs .nav-link.router-link-exact-active {
  color: var(--color-text);
  text-decoration: none;
  background-color: var(--color-background-mute);
  opacity: 0.8;
}
</style>
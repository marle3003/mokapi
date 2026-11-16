<script setup lang="ts">
import { useRoute } from 'vue-router'
import Servers from './Server.vue'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import LdapRequests from './Requests.vue'
import Configs from '../Configs.vue'
import Request from './Request.vue'
import { getRouteName } from '@/composables/dashboard';
import { useRouter } from '@/router'
import { withSelection } from '@/composables/log-selection'
import '@/assets/ldap.css'
import { useService } from '@/composables/service.ts'
import { provideBreadcrumb } from '@/composables/breadcrumb.ts'
import { computed, ref, watch } from 'vue'
import Breadcrumb from '../Breadcrumb.vue'

const route = useRoute();
const router = useRouter();
const { service } = useService<LdapService>('ldap')

const activeTab = ref('tab-requests')

function setTab(tab: string) {
  router.replace({
    query: withSelection(route.query, 'request', undefined),
    hash: `#${tab}`
  })
}
watch(() => route.hash, (hash) => {
    activeTab.value = hash ? hash.slice(1) : 'tab-requests'
  },
  { immediate: true }
)

const tail = provideBreadcrumb()
const breadcrumb = computed<BreadcrumbItem[]>(() => {
  const base: BreadcrumbItem[] = [
    { text: 'LDAP', to: { name: getRouteName('ldap').value } },
  ]
  if (!service.value) return base
  const isOverview = route.name === getRouteName('ldapService').value
  base.push(isOverview
    ? { text: service.value.name }
    : { text: service.value.name, to: { name: getRouteName('ldapService').value,
        params: { service: service.value.name } } })
  return [...base, ...tail.value]
})
</script>

<template>
  <Breadcrumb :items="breadcrumb" />

  <div v-if="$route.name == getRouteName('ldapService').value && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="LDAP" />
      </div>
      <div class="card-group">
        <section class="card" aria-label="Service Data">
          <div class="card-body">
            <div class="nav card-tabs" role="tablist">
              <button :class="{ active: activeTab === 'tab-requests' }" id="requests-tab" type="button" role="tab"
                aria-controls="requests-pane" @click="setTab('tab-requests')">
                Requests
              </button>
              <button :class="{ active: activeTab === 'tab-servers' }" id="servers-tab" type="button" role="tab"
                aria-controls="servers" @click="setTab('tab-servers')">
                Servers
              </button>
              <button :class="{ active: activeTab === 'tab-configs' }" id="configs-tab" type="button" role="tab"
                aria-controls="configs" @click="setTab('tab-configs')">
                Configs
              </button>
            </div>
            <div class="tab-content">
              <!-- v-if: the log reloads while it is mounted, so a hidden one would poll -->
              <div v-if="activeTab === 'tab-requests'" class="tab-pane fade show active" id="requests-pane"
                role="tabpanel" aria-labelledby="requests-tab">
                <div class="d-flex justify-content-end mb-2">
                  <router-link :to="{ name: getRouteName('ldapServiceLog').value, params: { service: service.name }, query: { refresh: route.query.refresh } }">
                    Open as tab
                  </router-link>
                </div>
                <div class="card-group">
                  <ldap-requests :service="service" />
                </div>
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-servers' }" id="servers"
                role="tabpanel" aria-labelledby="servers-tab">
                <servers :service="service" />
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-configs' }" id="configs"
                role="tabpanel" aria-labelledby="configs-tab">
                <configs :configs="service.configs" />
              </div>
            </div>
          </div>
        </section>
      </div>
  </div>
  <div class="card-group" v-if="$route.name == getRouteName('ldapServiceLog').value && service != null">
    <ldap-requests :service="service" />
  </div>
  <request v-if="$route.name == getRouteName('ldapRequest').value"></request>
</template>

<style scoped>
.tab-pane {
  padding: 0;
  padding-top: 1rem;
}
</style>

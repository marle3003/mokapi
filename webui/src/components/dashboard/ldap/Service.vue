<script setup lang="ts">
import { type Ref, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import Servers from './Server.vue'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import Searches from './Requests.vue'
import ConfigCard from '../ConfigCard.vue'
import Request from './Request.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';

const route = useRoute()
const serviceName = route.params.service?.toString()
let service: Ref<LdapService | null>
if (serviceName){
  const { dashboard } = useDashboard()
  const result = dashboard.value.getService(serviceName, 'ldap')
  service = result.service as Ref<LdapService | null>
  onUnmounted(() => {
      result.close()
  })
}
</script>

<template>
  <div v-if="$route.name == getRouteName('ldapService').value && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="LDAP" />
      </div>
      <div class="card-group">
        <servers :service="service" />
      </div>
      <div class="card-group">
          <config-card :configs="service.configs" />
      </div>
      <div class="card-group">
        <searches :service="service" />
      </div>
  </div>
  <request v-if="$route.name == getRouteName('ldapRequest').value"></request>
</template>
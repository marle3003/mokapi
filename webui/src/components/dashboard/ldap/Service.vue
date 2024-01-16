<script setup lang="ts">
import { type Ref, onUnmounted } from 'vue'
import { useService } from '@/composables/services'
import { useRoute } from 'vue-router'
import Servers from './Server.vue'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import Searches from './Searches.vue'
import Search from './Search.vue'
import ConfigCard from '../ConfigCard.vue'

const {fetchService} = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()
let service: Ref<LdapService | null>
if (serviceName){
    const result = <{service: Ref<LdapService | null>, close: () => void}>fetchService(serviceName, 'ldap')
    service = result.service
    onUnmounted(() => {
        result.close()
})
}
</script>

<template>
  <div v-if="$route.name == 'ldapService' && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="LDAP" />
      </div>
      <div class="card-group">
        <servers :service="service" />
      </div>
      <div class="card-group">
          <config-card :service="service" />
      </div>
      <div class="card-group">
        <searches :service="service" />
      </div>
  </div>
  <search v-if="$route.name == 'ldapSearch'"></search>
</template>
<script setup lang="ts">
import { useRoute } from 'vue-router'
import Servers from './Server.vue'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import Searches from './Requests.vue'
import ConfigCard from '../ConfigCard.vue'
import Request from './Request.vue'
import { getRouteName } from '@/composables/dashboard';
import '@/assets/ldap.css'
import { useService } from '@/composables/service.ts'
import { provideBreadcrumb } from '@/composables/breadcrumb.ts'
import { computed } from 'vue'
import Breadcrumb from '../Breadcrumb.vue'

const route = useRoute();
const { service } = useService<LdapService>('ldap')

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
<script setup lang="ts">
import { type Ref, onUnmounted } from 'vue'
import { useService } from '@/composables/services';
import { useRoute } from 'vue-router';
import Servers from './Servers.vue';
import ServiceInfoCard from '../ServiceInfoCard.vue';
import Mails from './Mails.vue';

const {fetchService} = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()
let service: Ref<SmtpService | null>
if (serviceName){
    const result = <{service: Ref<SmtpService | null>, close: () => void}>fetchService(serviceName, 'smtp')
    service = result.service
    onUnmounted(() => {
        result.close()
})
}
</script>

<template>
  <div v-if="$route.name == 'smtpService' && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="SMTP" />
      </div>
      <div class="card-group">
          <div class="card">
              <div class="card-body">
                  <div class="card-title text-center">Servers</div>
                  <servers :servers="[service.server]" />
              </div>
          </div>
      </div>
      <div class="card-group">
        <mails :service="service" />
      </div>
  </div>
</template>
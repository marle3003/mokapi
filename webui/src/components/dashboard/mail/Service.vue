<script setup lang="ts">
import { type Ref, onUnmounted } from 'vue'
import { useService } from '@/composables/services'
import { useRoute } from 'vue-router'
import Servers from './Servers.vue'
import Mailboxes from './Mailboxes.vue'
import Settings from './Settings.vue'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import Mails from './Mails.vue'
import Mail from './Mail.vue'
import Rules from './Rules.vue'
import ConfigCard from '../ConfigCard.vue'
import Mailbox from './Mailbox.vue'

const { fetchService } = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()
let service: Ref<MailService | null>
if (serviceName){
    const result = <{service: Ref<MailService | null>, close: () => void}>fetchService(serviceName, 'mail')
    service = result.service
    onUnmounted(() => {
        result.close()
})
}
</script>

<template>
  <div v-if="$route.name == 'mailService' && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="MAIL" />
      </div>

      <div class="card-group" v-if="$route.name === 'mailService'">
        <section class="card">
          <div class="card-body">
            <div class="nav card-tabs" id="myTab" role="tablist">
              <button class="active" id="servers-tab" data-bs-toggle="tab" data-bs-target="#servers" type="button" role="tab" aria-controls="servers" aria-selected="true">Servers</button>
              <button id="mailboxes-tab" data-bs-toggle="tab" data-bs-target="#mailboxes" type="button" role="tab" aria-controls="mailboxes" aria-selected="false">Mailboxes</button>
              <button v-if="service.rules && service.rules.length > 0" id="rules-tab" data-bs-toggle="tab" data-bs-target="#rules" type="button" role="tab" aria-controls="rules" aria-selected="false">Rules</button>
              <button id="settings-tab" data-bs-toggle="tab" data-bs-target="#settings" type="button" role="tab" aria-controls="settings" aria-selected="false">Settings</button>
              <button id="configs-tab" data-bs-toggle="tab" data-bs-target="#configs" type="button" role="tab" aria-controls="configs" aria-selected="false">Configs</button>
            </div>
            <div class="tab-content">
              <div class="tab-pane fade show active" id="servers" role="tabpanel" aria-labelledby="servers-tab">
                <servers :service="service" />
              </div>
            </div>
            <div class="tab-content">
              <div class="tab-pane fade" id="mailboxes" role="tabpanel" aria-labelledby="mailboxes-tab">
                <mailboxes :service="service" />
              </div>
            </div>
            <div class="tab-content" v-if="service.rules && service.rules.length > 0">
              <div class="tab-pane fade" id="rules" role="tabpanel" aria-labelledby="rules-tab">
                <rules :rules="service.rules" />
              </div>
            </div>
            <div class="tab-content">
              <div class="tab-pane fade" id="settings" role="tabpanel" aria-labelledby="settings-tab">
                <settings :settings="service.settings" />
              </div>
            </div>
            <div class="tab-content">
              <div class="tab-pane fade" id="configs" role="tabpanel" aria-labelledby="configs-tab">
                <config-card :configs="service.configs" :hide-title="true" />
              </div>
            </div>
          </div>
        </section>
      </div>

      <div class="card-group"  v-if="$route.name === 'mailService'">
        <mails :service="service" />
      </div>
  </div>
  <Mailbox v-if="$route.name === 'smtpMailbox' && service" :service="service" :mailbox-name="$route.params.name.toString()" />
  
  <mail v-if="$route.name == 'smtpMail'"></mail>
</template>

<style>
.dashboard .card-group .card.configs {
  margin: 0;
  border: none;
}
.dashboard .card.configs .card-body {
  padding: 0;
}
</style>
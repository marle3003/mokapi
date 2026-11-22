<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Servers from './Servers.vue'
import Mailboxes from './Mailboxes.vue'
import Settings from './Settings.vue'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import Mails from './Mails.vue'
import Mail from './Mail.vue'
import Rules from './Rules.vue'
import Configs from '../Configs.vue'
import Mailbox from './Mailbox.vue'
import { getRouteName } from '@/composables/dashboard';
import { useRouter } from '@/router'
import { withSelection } from '@/composables/log-selection'
import { useService } from '@/composables/service.ts'
import { provideBreadcrumb } from '@/composables/breadcrumb.ts'
import Breadcrumb from '../Breadcrumb.vue'

const route = useRoute()
const router = useRouter()
const { service } = useService<MailService>('mail')

const activeTab = ref('tab-mailboxes')

function setTab(tab: string) {
  router.replace({
    query: withSelection(route.query, 'message', undefined),
    hash: `#${tab}`
  })
}
watch(() => route.hash, (hash) => {
    activeTab.value = hash ? hash.slice(1) : 'tab-mailboxes'
  },
  { immediate: true }
)

const tail = provideBreadcrumb()
const breadcrumb = computed<BreadcrumbItem[]>(() => {
  const base: BreadcrumbItem[] = [
    { text: 'Mail', to: { name: getRouteName('mail').value } },
  ]
  if (!service.value) return base
  const isOverview = route.name === getRouteName('mailService').value
  base.push(isOverview
    ? { text: service.value.name }
    : { text: service.value.name, to: { name: getRouteName('mailService').value,
        params: { service: service.value.name } } })
  return [...base, ...tail.value]
})
</script>

<template>
  <Breadcrumb :items="breadcrumb" />

  <div v-if="$route.name == getRouteName('mailService').value && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="Mail" />
      </div>

      <div class="card-group">
        <section class="card" aria-label="Service Data">
          <div class="card-body">
            <div class="nav card-tabs" role="tablist">
              <button :class="{ active: activeTab === 'tab-mailboxes' }" id="mailboxes-tab" type="button" role="tab"
                aria-controls="mailboxes" @click="setTab('tab-mailboxes')">
                Mailboxes
              </button>
              <button :class="{ active: activeTab === 'tab-mails' }" id="mails-tab" type="button" role="tab"
                aria-controls="mails-pane" @click="setTab('tab-mails')">
                Mails
              </button>
              <button :class="{ active: activeTab === 'tab-servers' }" id="servers-tab" type="button" role="tab"
                aria-controls="servers" @click="setTab('tab-servers')">
                Servers
              </button>
              <button v-if="service.rules && service.rules.length > 0" :class="{ active: activeTab === 'tab-rules' }" id="rules-tab"
                type="button" role="tab" aria-controls="rules" @click="setTab('tab-rules')">
                Rules
              </button>
              <button :class="{ active: activeTab === 'tab-settings' }" id="settings-tab" type="button" role="tab"
                aria-controls="settings" @click="setTab('tab-settings')">
                Settings
              </button>
              <button :class="{ active: activeTab === 'tab-configs' }" id="configs-tab" type="button" role="tab"
                aria-controls="configs" @click="setTab('tab-configs')">
                Configs
              </button>
            </div>
            <div class="tab-content">
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-mailboxes' }" id="mailboxes" role="tabpanel" aria-labelledby="mailboxes-tab">
                <mailboxes :service="service" />
              </div>
              <!-- v-if: the log reloads while it is mounted, so a hidden one would poll -->
              <div v-if="activeTab === 'tab-mails'" class="tab-pane fade show active" id="mails-pane"
                role="tabpanel" aria-labelledby="mails-tab">
                <div class="d-flex justify-content-end mb-2">
                  <router-link :to="{ name: getRouteName('mailServiceLog').value, params: { service: service.name }, query: { refresh: route.query.refresh } }">
                    Open as tab
                  </router-link>
                </div>
                <div class="card-group">
                  <mails :service="service" />
                </div>
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-servers' }" id="servers" role="tabpanel" aria-labelledby="servers-tab">
                <servers :service="service" />
              </div>
              <div v-if="service.rules && service.rules.length > 0" class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-rules' }" id="rules" role="tabpanel" aria-labelledby="rules-tab">
                <rules :rules="service.rules" />
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-settings' }" id="settings" role="tabpanel" aria-labelledby="settings-tab">
                <settings :settings="service.settings" />
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-configs' }" id="configs" role="tabpanel" aria-labelledby="configs-tab">
                <configs :configs="service.configs" />
              </div>
            </div>
          </div>
        </section>
      </div>
  </div>
  <div class="card-group" v-if="$route.name == getRouteName('mailServiceLog').value && service != null">
    <mails :service="service" />
  </div>
  <Mailbox v-if="$route.name === getRouteName('smtpMailbox').value && service" :service="service" :mailbox-name="$route.params.name!.toString()" />
  
  <mail v-if="$route.name == getRouteName('smtpMail').value"></mail>
</template>
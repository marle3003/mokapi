<script setup lang="ts">
import { useRouter } from "vue-router";
import { useRoute } from "@/router";
import {  onUnmounted, type PropType } from "vue";
import Markdown from "vue3-markdown-it";
import { usePrettyDates } from "@/composables/usePrettyDate";
import Message from "@/components/Message.vue";
import { getRouteName, useDashboard } from "@/composables/dashboard";

const props = defineProps({
  service: { type: Object as PropType<MailService>, required: true },
  mailboxName: { type: String, required: true },
});

const { dashboard } = useDashboard()
const { mailbox, isLoading } = dashboard.value.getMailbox(props.service.name, props.mailboxName)
const { messages, close } = dashboard.value.getMailboxMessages(props.service.name, props.mailboxName)
const route = useRoute();
const router = useRouter();
const { format } = usePrettyDates();

onUnmounted(() => close());

function goToMail(msg: MessageInfo, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('smtpMail').value,
    params: { id: msg.messageId },
  };
  if (openInNewTab) {
    const routeData = router.resolve(to);
    window.open(routeData.href, '_blank')
  } else {
    router.push(to)
  }
}
</script>

<template>
  <div v-if="!mailbox && !isLoading">
        <message :message="`Mailbox ${mailboxName} not found`"></message>
  </div>
  <div v-else="!isLoading">
    <div class="card-group">
      <section class="card" aria-label="Mailbox">
        <div class="card-body" v-if="mailbox">
          <div class="row">
            <div class="col col-6 header mb-3">
              <label id="mailbox-name" class="label">Mailbox Name</label>
              <p aria-labelledby="mailbox-name">
                {{ mailbox.name }}
              </p>
            </div>
            <div class="col">
              <p id="service" class="label">Service</p>
              <p aria-labelledby="service">
                <router-link :to="route.service(service, 'mail')">
                  {{ service.name }}
                </router-link>
              </p>
            </div>
            <div class="col text-end">
              <span class="badge bg-secondary" aria-label="Service Type">MAIL</span>
            </div>
          </div>
          <div class="row mb-2">
            <div class="col-2">
              <p id="username" class="label">Username</p>
              <p aria-labelledby="username">{{ mailbox.username }}</p>
            </div>
            <div class="col-4">
              <p id="password" class="label">Password</p>
              <p aria-labelledby="password">{{ mailbox.password }}</p>
            </div>
          </div>
          <div class="row" v-if="mailbox.description">
            <div class="col">
              <p id="mailbox-description" class="label">Description</p>
              <markdown
                :source="mailbox.description"
                aria-labelledby="mailbox-description"
                :html="true"
              ></markdown>
            </div>
          </div>
        </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="folders-title">
        <div class="card-body">
          <h2 id="folders-title" class="card-title text-center">Folders</h2>
          <table class="table dataTable" aria-labelledby="folders-title">
            <thead>
              <tr>
                <th scope="col" class="text-left" style="width: 20%">Name</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="folder in mailbox?.folders" :key="folder">
                <td>{{ folder }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="mails-title">
        <div class="card-body">
          <h2 id="mails-title" class="card-title text-center">Mails</h2>
          <table class="table dataTable selectable" aria-labelledby="mails-title">
            <thead>
              <tr>
                <th scope="col" class="text-left" style="width: 20%">Subject</th>
                <th scope="col" class="text-left" style="width: 20%">From</th>
                <th scope="col" class="text-left" style="width: 20%">To</th>
                <th scope="col" class="text-center" style="width: 15%">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="message in messages"
                :key="message.messageId"
                @mouseup.left="goToMail(message)" @mousedown.middle="goToMail(message, true)"
              >
                <td>
                  <router-link @click.stop class="row-link" :to="{ name: getRouteName('smtpMail').value, params: { id: message.messageId } }">
                    {{ message.subject }}
                  </router-link>
                </td>
                <td class="address-list">
                  <ul class="list-unstyled">
                    <li v-for="addr of message.from">
                      <strong v-if="addr.name">{{ addr.name }}</strong>
                      <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                      <span v-else>{{ addr.address }}</span>
                    </li>
                  </ul>
                </td>
                <td class="address-list">
                  <ul class="list-unstyled">
                    <li v-for="addr of message.to">
                      <strong v-if="addr.name">{{ addr.name }}</strong>
                      <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                      <span v-else>{{ addr.address }}</span>
                    </li>
                  </ul>
                </td>
                <td class="text-center">{{ format(message.date) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.address-list ul {
    margin-bottom: 0;
}
</style>
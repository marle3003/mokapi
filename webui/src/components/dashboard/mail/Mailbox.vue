<script setup lang="ts">
import { transformPath, useFetch } from "@/composables/fetch";
import { useRouter } from "vue-router";
import { useRoute } from "@/router";
import { onMounted, onUnmounted, ref, watchEffect, type PropType } from "vue";
import Markdown from "vue3-markdown-it";
import { usePrettyDates } from "@/composables/usePrettyDate";

const props = defineProps({
  service: { type: Object as PropType<MailService>, required: true },
  mailboxName: { type: String, required: true },
});

const mailbox = ref<SmtpMailbox | null>();
const route = useRoute();
const router = useRouter();
const { format } = usePrettyDates();

onMounted(() => {
  const path = transformPath(
    `/api/services/mail/${props.service.name}/mailboxes/${props.mailboxName}`,
  );
  fetch(path)
    .then((res) => res.json())
    .then((data) => (mailbox.value = data))
    .catch((err) => console.log(err));
});

onUnmounted(() => res.close());

const res = useFetch(
  `/api/services/mail/${props.service.name}/mailboxes/${props.mailboxName}/messages`,
);
const messages = ref<Message[]>()

watchEffect(() => {
  if (!res.data) {
    messages.value = [];
    return;
  }
  messages.value = res.data;
});

function goToMail(msg: Message) {
  router.push({
    name: "smtpMail",
    params: { id: msg.data.messageId },
  });
}
</script>

<template>
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

  <div class="card-group" v-if="mailbox?.folders">
    <div class="card">
      <div class="card-body">
        <div class="card-title text-center">Folders</div>
        <table class="table dataTable">
          <caption class="visually-hidden">Folders</caption>
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
    </div>
  </div>

  <div class="card-group">
    <div class="card">
      <div class="card-body">
        <div class="card-title text-center">Mails</div>
        <table class="table dataTable selectable">
          <caption class="visually-hidden">
            Mails
          </caption>
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
              :key="message.data.messageId"
              @click="goToMail(message)"
            >
              <td>{{ message.data.subject }}</td>
              <td>
                <ul class="list-unstyled">
                  <li v-for="addr of message.data.from">
                    <strong v-if="addr.name">{{ addr.name }}</strong>
                    <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                    <span v-else>{{ addr.address }}</span>
                  </li>
                </ul>
              </td>
              <td>
                <ul class="list-unstyled">
                  <li v-for="addr of message.data.to">
                    <strong v-if="addr.name">{{ addr.name }}</strong>
                    <span v-if="addr.name"> &lt;{{ addr.address }}&gt;</span>
                    <span v-else>{{ addr.address }}</span>
                  </li>
                </ul>
              </td>
              <td class="text-center">{{ format(message.data.date) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

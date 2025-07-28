<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed, type PropType } from 'vue';
import Markdown from 'vue3-markdown-it'

const props = defineProps({
    service: { type: Object as PropType<MailService>, required: true },
})

const route = useRoute()
const router = useRouter()

const anyDescription = computed(() => {
  if (!props.service.mailboxes) {
    return false
  }

  for (const mb of props.service.mailboxes) {
    if (mb.description && mb.description !== '') {
      return true
    }
  }
  return false
})

function goToMailbox(mb: SmtpMailbox){
    if (getSelection()?.toString()) {
        return
    }
    
    router.push({
        name: 'smtpMailbox',
        params: {
          service: props.service.name,
          name: mb.name
        },
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <table class="table dataTable selectable">
        <caption class="visually-hidden">Mailboxes</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left">Mailbox</th>
                <th scope="col" class="text-left">Username</th>
                <th scope="col" class="text-left">Password</th>
                <th v-if="anyDescription" scope="col" class="text-left">Description</th>
                <th scope="col" class="text-center" style="width: 20%">Received Messages</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="mb in service.mailboxes" :key="mb.name" @click="goToMailbox(mb)">
                <td>{{ mb.name }}</td>
                <td>{{ mb.username }}</td>
                <td>{{ mb.password }}</td>
                <td v-if="anyDescription"><markdown :source="mb.description" class="description" :html="true"></markdown></td>
                <td class="text-center">{{ mb.numMessages }}</td>
            </tr>
        </tbody>
    </table>
</template>
<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { computed, type PropType } from 'vue';
import Markdown from 'vue3-markdown-it'
import { getRouteName } from '@/composables/dashboard';

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

const mailboxes = computed(() => {
  if (!props.service.mailboxes) {
    return undefined
  }
  return props.service.mailboxes.sort((x, y) => x.name.localeCompare(y.name))
});

function goToMailbox(mb: SmtpMailbox, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }
    
    const to = {
        name: getRouteName('smtpMailbox').value,
        params: {
          service: props.service.name,
          name: mb.name
        },
    }
    if (openInNewTab) {
      const routeData = router.resolve(to);
      window.open(routeData.href, '_blank')
    } else {
      router.push(to)
    }
}
</script>

<template>
  <div class="table-responsive-sm">
    <table class="table dataTable selectable" aria-label="Mailboxes">
        <thead>
            <tr>
                <th scope="col" class="text-left">Mailbox</th>
                <th scope="col" class="text-left">Username</th>
                <th scope="col" class="text-left">Password</th>
                <th v-if="anyDescription" scope="col" class="text-left">Description</th>
                <th scope="col" class="text-center col-1">Mails</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="mb in mailboxes" :key="mb.name" @mouseup.left="goToMailbox(mb)" @mousedown.middle="goToMailbox(mb, true)">
                <td>
                  <router-link @click.stop class="row-link" :to="{ name: getRouteName('smtpMailbox').value, params: { service: props.service.name, name: mb.name } }">
                    {{ mb.name }}
                  </router-link>
                </td>
                <td>{{ mb.username }}</td>
                <td>{{ mb.password }}</td>
                <td v-if="anyDescription"><markdown :source="mb.description" class="description" :html="true"></markdown></td>
                <td class="text-center">{{ mb.numMessages }}</td>
            </tr>
        </tbody>
    </table>
  </div>
</template>
<script setup lang="ts">
import { transformPath } from '@/composables/fetch';
import { Modal } from 'bootstrap';
import { computed, onMounted, ref, type PropType } from 'vue';
import Markdown from 'vue3-markdown-it'

const props = defineProps({
    service: { type: Object as PropType<MailService>, required: true },
})

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

let mailbox = ref<SmtpMailbox | null>(null)
const mailboxModal = ref<Element | null>(null)
let dialog:  Modal

onMounted(()=> {
  if (!mailboxModal.value) {
    return
  }
  dialog = new Modal(mailboxModal.value)
})

function showMailbox(mb: SmtpMailbox){
  if (getSelection()?.toString()) {
    return
  }

  const path = transformPath('/api/services/mail/{{mailServerName}}/mailboxes/'+mb.name)
  fetch(path).then(res => res.json()).then(data => mailbox.value = data).catch(err => console.log(err))

  if (dialog) {
    dialog.show()
  }
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
            <tr v-for="mb in service.mailboxes" :key="mb.name" @click="showMailbox(mb)">
                <td>{{ mb.name }}</td>
                <td>{{ mb.username }}</td>
                <td>{{ mb.password }}</td>
                <td v-if="anyDescription"><markdown :source="mb.description" class="description" :html="true"></markdown></td>
                <td class="text-center">{{ mb.numMessages }}</td>
            </tr>
        </tbody>
    </table>

  <!-- Modal -->
  <div class="modal fade" id="mailboxModal" ref="mailboxModal" tabindex="-1" aria-hidden="true" aria-labelledby="dialog-mailbox-title">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">

        <div class="modal-header">
              <h6 id="dialog-mailbox-title" class="modal-title">Mailbox Details</h6>
              <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>

        <div class="modal-body" v-if="mailbox">
          <div class="card">
            <div class="card-body">
              <div class="row mb-2">
                <div class="col">
                  <p id="dlg-mailbox-name" class="label">Name</p>
                  <p aria-labelledby="dlg-mailbox-name">{{ mailbox.name }}</p>
                </div>
              </div>
              <div class="row mb-2">
                <div class="col-6">
                  <p id="dlg-mailbox-username" class="label">Username</p>
                  <p aria-labelledby="dlg-mailbox-username">{{ mailbox.username }}</p>
                </div>
                <div class="col-6">
                  <p id="dlg-mailbox-password" class="label">Password</p>
                  <p aria-labelledby="dlg-mailbox-password">{{ mailbox.password }}</p>
                </div>
              </div>
              <div class="row mb-2" v-if="mailbox.description">
                <div class="col">
                  <p id="dlg-mailbox-description" class="label">Description</p>
                  <markdown aria-labelledby="dlg-mailbox-description" :source="mailbox.description" class="description" :html="true"></markdown>
                </div>
              </div>
              <div class="row mb-2">
                <div class="col-6">
                  <p id="dlg-mailbox-mails" class="label">Mails</p>
                  <p aria-labelledby="dlg-mailbox-mails">{{ mailbox.numMessages }}</p>
                </div>
                <div class="col-6">
                  <p id="dlg-mailbox-folders" class="label">Folders</p>
                  <p v-if="mailbox.folders" aria-labelledby="dlg-mailbox-folders">{{ mailbox.folders.length }}</p>
                  <p v-else aria-labelledby="dlg-mailbox-folders">0</p>
                </div>
              </div>
            </div>
          </div>
          <div v-if="mailbox.folders && mailbox.folders.length > 0" class="card">
            <div class="card-body">
              <div class="card-title text-center">Folders</div>
              <table class="table dataTable selectable">
                <caption class="visually-hidden">Mailbox folders</caption>
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Name</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="folder in mailbox.folders" :key="folder">
                        <td>{{ folder }}</td>
                    </tr>
                </tbody>
              </table>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>
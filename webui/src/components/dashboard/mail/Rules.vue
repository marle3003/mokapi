<script setup lang="ts">
import { Modal } from "bootstrap";
import { onMounted, ref, type PropType } from "vue";
import Markdown from 'vue3-markdown-it'

defineProps({
  rules: { type: Array as PropType<Array<SmtpRule>>, required: true },
});

let rule = ref<SmtpRule | null>(null);
const ruleModal = ref<Element | null>(null);
let dialog: Modal;

onMounted(() => {
  if (!ruleModal.value) {
    return;
  }
  dialog = new Modal(ruleModal.value);
});

function showRule(r: SmtpRule) {
  if (getSelection()?.toString()) {
    return;
  }

  rule.value = r;

  if (dialog) {
    dialog.show();
  }
}
</script>

<template>
  <table class="table dataTable selectable" data-testid="rules">
    <caption class="visually-hidden">
      Rules
    </caption>
    <thead>
      <tr>
        <th scope="col" class="text-left">Name</th>
        <th scope="col">Action</th>
        <th scope="col" class="text-left">Sender</th>
        <th scope="col" class="text-left">Recipient</th>
        <th scope="col" class="text-left">Subject</th>
        <th scope="col" class="text-left">Body</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="rule in rules" :key="rule.name" @click="showRule(rule)">
        <td>{{ rule.name }}</td>
        <td>{{ rule.action }}</td>
        <td>{{ rule.sender }}</td>
        <td>{{ rule.recipient }}</td>
        <td>{{ rule.subject }}</td>
        <td>{{ rule.body }}</td>
      </tr>
    </tbody>
  </table>

  <!-- Modal -->
  <div
    class="modal fade"
    id="ruleModal"
    ref="ruleModal"
    tabindex="-1"
    aria-hidden="true"
    aria-labelledby="dialog-rule-title"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h6 id="dialog-rule-title" class="modal-title">Rule Details</h6>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
            aria-label="Close"
          ></button>
        </div>

        <div class="modal-body" v-if="rule">
          <div class="card">
            <div class="card-body">
              <div class="row mb-2">
                <div class="col-8">
                  <p id="dlg-rule-name" class="label">Name</p>
                  <p aria-labelledby="dlg-rule-name">{{ rule.name }}</p>
                </div>
                <div class="col-4">
                  <p id="dlg-rule-action" class="label">Action</p>
                  <p aria-labelledby="dlg-rule-action">{{ rule.action }}</p>
                </div>
              </div>
              <div class="row mb-2" v-if="rule.description">
                <div class="col">
                  <p id="dlg-rule-description" class="label">Description</p>
                  <markdown
                    aria-labelledby="dlg-rule-description"
                    :source="rule.description"
                    class="description"
                    :html="true"
                  ></markdown>
                </div>
              </div>
              <div class="row mb-2" v-if="rule.sender">
                <div class="col">
                  <p id="dlg-rule-sender" class="label">Sender</p>
                  <p aria-labelledby="dlg-rule-sender">
                    {{ rule.sender }}
                  </p>
                </div>
              </div>
              <div class="row mb-2" v-if="rule.recipient">
                <div class="col">
                  <p id="dlg-rule-recipient" class="label">Recipient</p>
                  <p aria-labelledby="dlg-rule-recipient">
                    {{ rule.recipient }}
                  </p>
                </div>
              </div>
              <div class="row mb-2" v-if="rule.subject">
                <div class="col">
                  <p id="dlg-rule-subject" class="label">Subject</p>
                  <p aria-labelledby="dlg-rule-subject">
                    {{ rule.subject }}
                  </p>
                </div>
              </div>
              <div class="row mb-2" v-if="rule.body">
                <div class="col">
                  <p id="dlg-rule-body" class="label">Body</p>
                  <p aria-labelledby="dlg-rule-body">
                    {{ rule.body }}
                  </p>
                </div>
              </div>
              <div v-if="rule.rejectResponse">
                <div class="row mb-2">
                  <div class="col-6">
                    <p id="dlg-rule-status" class="label">
                      Reject Response Status Code
                    </p>
                    <p aria-labelledby="dlg-rule-status">
                      {{ rule.rejectResponse.statusCode }}
                    </p>
                  </div>
                  <div class="col-6">
                    <p id="dlg-rule-enhanced" class="label">
                      Reject Response Enhanced Code
                    </p>
                    <p aria-labelledby="dlg-rule-enhanced">
                      {{ rule.rejectResponse.enhancedStatusCode }}
                    </p>
                  </div>
                </div>
                <div class="row mb-2">
                  <div class="col">
                    <p id="dlg-rule-message" class="label">
                      Reject Response Message
                    </p>
                    <p aria-labelledby="dlg-rule-message">
                      {{ rule.rejectResponse.message }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

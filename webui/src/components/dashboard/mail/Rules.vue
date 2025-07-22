<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    rules: { type: Array as PropType<Array<SmtpRule>>, required: true },
})

</script>

<template>
    <table class="table dataTable" data-testid="rules">
        <thead>
            <tr>
              <th scope="col"></th>
              <th scope="col" class="text-left">Name</th>
              <th scope="col" class="text-left">Sender</th>
              <th scope="col" class="text-left">Recipient</th>
              <th scope="col" class="text-left">Subject</th>
              <th scope="col" class="text-left">Body</th>
              <th scope="col" class="text-left">Action</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="rule in rules">
              <td>
                <i class="bi bi-check-lg" v-if="rule.action == 'allow'"></i>
                <i class="bi bi-x-lg" v-if="rule.action == 'deny'"></i>
              </td>
              <td>{{ rule.name }}</td>
              <td>{{ rule.sender }}</td>
              <td>{{ rule.recipient }}</td>
              <td>{{ rule.subject }}</td>
              <td>{{ rule.body }}</td>
              <td>{{ rule.action }}</td>
            </tr>
        </tbody>
    </table>
</template>
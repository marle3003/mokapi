<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    rules: { type: Array as PropType<Array<SmtpRule>>, required: true },
})

</script>

<template>
    <table class="table dataTable" data-testid="rules">
        <caption class="visually-hidden">Rules</caption>
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
            <tr v-for="rule in rules">
              <td>{{ rule.name }}</td>
              <td>{{ rule.action }}</td>
              <td>{{ rule.sender }}</td>
              <td>{{ rule.recipient }}</td>
              <td>{{ rule.subject }}</td>
              <td>{{ rule.body }}</td>
            </tr>
        </tbody>
    </table>
</template>
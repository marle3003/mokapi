<script setup lang="ts">
defineProps<{
  response: MqttSubscribeResponse
}>();
</script>

<template>
  <div class="table-responsive-sm mt-4">
    <table class="table dataTable selectable" aria-label="Topics">
      <thead>
        <tr>
          <th scope="col" class="text-left col">Reason Codes</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(code, codeIndex) in response.reasonCodes" :key="codeIndex">
          <td>{{ code }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
<script setup lang="ts">
import { type PropType, ref, reactive } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { transformPath } from '@/composables/fetch'

defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
})

const { format } = usePrettyDates()

function showConfig(config: Config, newTab: boolean){
  if (newTab) {
    window.open(transformPath(`/api/configs/${config.id}`))
  } else {
    window.location.href = transformPath(`/api/configs/${config.id}`)
  }
}
</script>

<template>
  <div class="card">
      <div class="card-body">
          <div class="card-title text-center">Configs</div>
          <table class="table dataTable selectable">
              <thead>
                  <tr>
                      <th scope="col" class="text-left">URL</th>
                      <th scope="col" class="text-left">Provider</th>
                      <th scope="col" class="text-left">Last Update</th>
                  </tr>
              </thead>
              <tbody>
                  <tr v-for="config in service.configs" :key="config.url" @mousedown.left="showConfig(config, false)" @mousedown.middle="showConfig(config, true)">
                      <td>{{ config.url }}</td>
                      <td>{{ config.provider }}</td>
                      <td>{{ format(config.time) }}</td>
                  </tr>
              </tbody>
          </table>
      </div>
  </div>
</template>
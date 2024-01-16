<script setup lang="ts">
import { type PropType, computed } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { transformPath } from '@/composables/fetch'

const { format } = usePrettyDates()

const props = defineProps({
    service: { type: Object as PropType<Service>, required: true },
})

const configs = computed(() => {
    if (!props.service.configs) {
        return []
    }
    return props.service.configs.sort(compareConfig)
})

function compareConfig(c1: Config, c2: Config) {
    const url1 = c1.url.toLowerCase()
    const url2 = c2.url.toLowerCase()
    return url1.localeCompare(url2)
}

function showConfig(config: Config, newTab: boolean){
  const selection = getSelection()?.toString()
  if (selection) {
    return
  }

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
                  <tr v-for="config in configs" :key="config.url" @mouseup.left="showConfig(config, false)" @mousedown.middle="showConfig(config, true)">
                      <td>{{ config.url }}</td>
                      <td>{{ config.provider }}</td>
                      <td>{{ format(config.time) }}</td>
                  </tr>
              </tbody>
          </table>
      </div>
  </div>
</template>
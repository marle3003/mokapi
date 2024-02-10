<script setup lang="ts">
import { type PropType, computed } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { transformPath } from '@/composables/fetch'
import { useRouter } from '@/router';
import { useRoute } from 'vue-router';

const { format } = usePrettyDates()

const props = defineProps({
    configs: { type: Object as PropType<Config[] | ConfigRef[]>, required: true },
    title: { type: Object as PropType<string>, required: false },
})

const route = useRoute()

const configs = computed(() => {
    if (!props.configs) {
        return []
    }
    return props.configs.sort(compareConfig)
})

function compareConfig(c1: Config | ConfigRef, c2: Config | ConfigRef) {
    const url1 = c1.url.toLowerCase()
    const url2 = c2.url.toLowerCase()
    return url1.localeCompare(url2)
}

function showConfig(config: Config | ConfigRef, newTab: boolean){
  const selection = getSelection()?.toString()
  if (selection) {
    return
  }

  useRouter().push({
        name: 'config',
        params: { id: config.id },
        query: { refresh: route.query.refresh }
    })
    return

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
          <div class="card-title text-center">{{ title ? title : "Configs" }}</div>
          <table class="table dataTable selectable">
              <thead>
                  <tr>
                      <th scope="col" class="text-left w-100">URL</th>
                      <th scope="col" class="text-center" style="min-width: 100px;">Provider</th>
                      <th scope="col" class="text-center" style="min-width: 200px;">Last Update</th>
                  </tr>
              </thead>
              <tbody>
                  <tr v-for="config in configs" :key="config.url" @mouseup.left="showConfig(config, false)" @mousedown.middle="showConfig(config, true)">
                      <td>{{ config.url }}</td>
                      <td class="text-center">{{ config.provider }}</td>
                      <td class="text-center">{{ format(config.time) }}</td>
                  </tr>
              </tbody>
          </table>
      </div>
  </div>
</template>
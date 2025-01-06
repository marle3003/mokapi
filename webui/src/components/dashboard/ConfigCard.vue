<script setup lang="ts">
import { computed } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRouter } from '@/router'
import { useRoute } from 'vue-router'

const { format } = usePrettyDates()

const props = defineProps<{
    configs: Config[] | ConfigRef[] | undefined,
    title?: string,
}>()

const route = useRoute()

const configs = computed(() => {
    if (!props.configs) {
        return []
    }
    return props.configs.sort(compareConfig)
})

const title = computed(() => props.title ? props.title : "Configs")

function compareConfig(c1: Config | ConfigRef, c2: Config | ConfigRef) {
    const url1 = c1.url.toLowerCase()
    const url2 = c2.url.toLowerCase()
    return url1.localeCompare(url2)
}

function showConfig(config: Config | ConfigRef){
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
}
</script>

<template>
  <section class="card" aria-labelledby="configs">
      <div class="card-body">
          <div id="configs" class="card-title text-center">{{ title }}</div>
          <table class="table dataTable selectable" style="table-layout: fixed;">
            <caption class="visually-hidden">{{ title }}</caption>
              <thead>
                  <tr>
                      <th scope="col" class="text-left col-6 col-md-9">URL</th>
                      <th scope="col" class="text-center col-2">Provider</th>
                      <th scope="col" class="text-center col-2">Last Update</th>
                  </tr>
              </thead>
              <tbody>
                  <tr scope="row" v-for="config in configs" :key="config.url" @mouseup.left="showConfig(config)" @mousedown.middle="showConfig(config)">
                      <td>{{ config.url }}</td>
                      <td class="text-center">{{ config.provider }}</td>
                      <td class="text-center">{{ format(config.time) }}</td>
                  </tr>
              </tbody>
          </table>
      </div>
  </section>
</template>
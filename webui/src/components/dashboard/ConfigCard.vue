<script setup lang="ts">
import { computed, onUnmounted, type Ref } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRouter } from '@/router'
import { useRoute } from 'vue-router'
import { getRouteName, useDashboard } from '@/composables/dashboard'

const { format } = usePrettyDates()

const props = withDefaults(defineProps<{
    configs?: Config[] | ConfigRef[] | null,
    title?: string,
    hideTitle?: boolean
    useCard?: boolean
}>(), { useCard: true })

const route = useRoute()
let data: Ref<Config[] | null> | undefined

if (props.configs === undefined) {
    const result = useDashboard().dashboard.value.getConfigs()
    data = result.data
    onUnmounted(() => {
        result.close();
    })
}

const configs = computed(() => {
    if (!props.configs) {
        return data?.value ?? []
    }
    return props.configs.sort(compareConfig)
})

const title = computed(() => props.title ? props.title : "Configs")

function compareConfig(c1: Config | ConfigRef, c2: Config | ConfigRef) {
    const url1 = c1.url.toLowerCase()
    const url2 = c2.url.toLowerCase()
    return url1.localeCompare(url2)
}

function gotToConfig(config: Config | ConfigRef, openInNewTab = false){
  const selection = getSelection()?.toString()
  if (selection) {
    return
  }

  const router = useRouter();
  const to = {
    name: getRouteName('config').value,
    params: { id: config.id },
    query: { refresh: route.query.refresh }
  }

  if (openInNewTab) {
    const routeData = router.resolve(to);
    window.open(routeData.href, '_blank')
  } else {
    useRouter().push(to)
  }
}
</script>

<template>
  <section class="card" aria-labelledby="configs" v-if="useCard">
      <div class="card-body">
          <h2 v-if="!hideTitle" id="configs" class="card-title text-center">{{ title }}</h2>
          <table class="table dataTable selectable" style="table-layout: fixed;" aria-labelledby="configs">
            <caption class="visually-hidden">{{ title }}</caption>
              <thead>
                  <tr>
                      <th scope="col" class="text-left col-6 col-md-9">URL</th>
                      <th scope="col" class="text-center col-2">Provider</th>
                      <th scope="col" class="text-center col-2">Last Update</th>
                  </tr>
              </thead>
              <tbody>
                  <tr scope="row" v-for="config in configs" :key="config.url" @mouseup.left="gotToConfig(config)" @mousedown.middle="gotToConfig(config, true)">
                      <td>
                            <router-link @click.stop class="row-link" :to="{ name: getRouteName('config').value, params: { id: config.id } }">
                            {{ config.url }}
                            </router-link>
                      </td>
                      <td class="text-center">{{ config.provider }}</td>
                      <td class="text-center">{{ format(config.time) }}</td>
                  </tr>
              </tbody>
          </table>
      </div>
  </section>

  <table class="table dataTable selectable" style="table-layout: fixed;" aria-labelledby="configs" v-else>
    <caption class="visually-hidden">{{ title }}</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left col-6 col-md-9">URL</th>
                <th scope="col" class="text-center col-2">Provider</th>
                <th scope="col" class="text-center col-2">Last Update</th>
            </tr>
        </thead>
        <tbody>
            <tr scope="row" v-for="config in configs" :key="config.url" @mouseup.left="gotToConfig(config)" @mousedown.middle="gotToConfig(config, true)">
                <td>
                    <router-link @click.stop class="row-link" :to="{ name: getRouteName('config').value, params: { id: config.id } }">
                        {{ config.url }}
                    </router-link>
                </td>
                <td class="text-center">{{ config.provider }}</td>
                <td class="text-center">{{ format(config.time) }}</td>
            </tr>
        </tbody>
    </table>

</template>

<style scoped>
    .row-link {
        text-decoration: none;
        color: inherit;
    }

</style>
<script setup lang="ts">
import { computed, onUnmounted, type Ref } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRouter } from '@/router'
import { getRouteName, useDashboard } from '@/composables/dashboard'

const { format } = usePrettyDates()

const props = withDefaults(defineProps<{
    configs?: Config[] | ConfigRef[] | null,
    title?: string,
    hideTitle?: boolean
    useCard?: boolean
}>(), { useCard: true })

let data: Ref<Config[] | null> | undefined

if (props.configs === undefined) {
    const result = useDashboard().dashboard.value.getConfigs()
    data = result.data
    onUnmounted(() => {
        result.close();
    })
}

const configs = computed(() => {
    if (props.configs) {
        return props.configs.sort(compareConfig)
    }
    if (data && data.value) {
        console.log('sort')
        return data.value.sort(compareConfig)
    }
    return [];
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
    params: { id: config.id }
  }

  if (openInNewTab) {
    const routeData = router.resolve(to);
    window.open(routeData.href, '_blank')
  } else {
    router.push(to)
  }
}
function formatProvider(config: ConfigRef) {
    if (!config) {
        return '';
    }
    switch (config.provider.toLocaleLowerCase()) {
        case 'file': return 'File';
        case 'http': return 'HTTP';
        case 'git': return 'GIT';
        case 'npm': return 'NPM';
    }
    return '';
}
</script>

<template>
  <section class="card" aria-labelledby="configs" v-if="useCard">
      <div class="card-body">
          <h2 v-if="!hideTitle" id="configs" class="card-title text-center">{{ title }}</h2>
          <div class="table-responsive-sm">
            <table class="table dataTable selectable" aria-labelledby="configs">
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
                        <td class="text-center">{{ formatProvider(config) }}</td>
                        <td class="text-center">{{ format(config.time) }}</td>
                    </tr>
                </tbody>
            </table>
          </div>
      </div>
  </section>

  <div v-else class="table-responsive-sm">
    <table class="table dataTable selectable" aria-label="Configs">
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
                    <td class="text-center">{{ formatProvider(config) }}</td>
                    <td class="text-center">{{ format(config.time) }}</td>
                </tr>
            </tbody>
    </table>
  </div>

</template>
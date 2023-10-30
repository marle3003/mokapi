<script setup lang="ts">
import type { PropType } from 'vue';

const props = defineProps({
    level1: { type: String, required: true },
    level2: { type: String, required: true },
    level3: { type: String },
    config: { type: Object as PropType<DocConfig>, required: true },
})

function matchLevel2(label: any): boolean {
  return label.toString().toLowerCase() == props.level2.toLowerCase()
}
function matchLevel3(label: any): boolean {
  return label.toString().toLowerCase() == props.level3?.toLowerCase()
}
function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
function hasChildren(item: DocConfig | DocEntry | string) {
    if (typeof item === 'string') {
        return false
    }
    const entry = <DocEntry>item
    if ('file' in entry || 'component' in entry) {
        return false
    }
    return true
}
function showItem(name: string, item: DocConfig | DocEntry | string){
    if (typeof item == 'string' && props.level1 != name) {
        return true
    }
    const entry = <DocEntry>item
    if ('hideInNavigation' in entry) {
        return !entry.hideInNavigation
    }
    return true
}
</script>

<template>
    <ul class="nav nav-pills flex-column mb-auto pe-3">
        <li class="nav-item" v-for="(v, k) of config[level1]">
            <p v-if="hasChildren(v)" class="chapter-text">{{ k }}</p>
            <ul v-if="hasChildren(v)" class="nav nav-pills flex-column mb-auto pe-3 chapter">
                <li class="nav-item" v-for="(_, k2) of v">
                    <router-link v-if="k != k2" class="nav-link" :class="matchLevel2(k) && matchLevel3(k2) ? 'active': ''" :to="{ name: 'docs', params: {level2: formatParam(k), level3: formatParam(k2)} }" style="padding-left: 2rem">{{ k2 }}</router-link>
                </li>
            </ul>
            <router-link v-if="!hasChildren(v) && showItem(k, v)" class="nav-link chapter-text" :class="matchLevel2(k) ? 'active': ''" :to="{ name: 'docs', params: {level2: formatParam(k)} }">{{ k }}</router-link>
        </li>
    </ul>   
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { useFetch, type Response as FetchResponse } from '@/composables/fetch'
import { commaList, type FakerNode } from '@/composables/faker-tree'
import { generatorError } from '@/composables/faker-playground'
import FakerValue from './FakerValue.vue'

const props = defineProps<{ node: FakerNode }>()

const sample = ref<FetchResponse>()

// The node is addressed by its path, nodes of the same name in other branches stay apart.
function generate() {
    sample.value = useFetch(`/api/faker/node/fake?path=${encodeURIComponent(props.node.path)}`, undefined, false, false)
}
</script>

<template>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Name</p>
            <p>{{ node.name }}</p>
        </div>
        <div class="col">
            <p class="label">Path</p>
            <p>{{ node.path }}</p>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Attributes</p>
            <p>{{ commaList(node.attributes) }}</p>
        </div>
        <div v-if="node.weight" class="col">
            <p class="label">Weight</p>
            <p>{{ node.weight }}</p>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Depends on</p>
            <p>{{ commaList(node.dependsOn) }}</p>
        </div>
        <div class="col">
            <p class="label">Custom</p>
            <p>{{ node.custom ? 'Yes' : 'No' }}</p>
        </div>
    </div>
    <div v-if="node.hasFakeFunc">
        <button type="button" class="btn btn-primary btn-sm" @click="generate">Generate</button>
        <div v-if="sample" class="mt-2">
            <div v-if="sample.error" class="alert alert-danger" role="alert">{{ generatorError(sample.error) }}</div>
            <faker-value v-else-if="!sample.isLoading" :value="sample.data" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useFetch, type Response as FetchResponse } from '@/composables/fetch'
import { fakeRequest, fakeRequestInit, generatorError, isTreeNode, sourceNote, type FakeResult } from '@/composables/faker-playground'
import FakerValue from './FakerValue.vue'

defineEmits<{ selectNode: [path: string] }>()

const pathText = ref('')
const schemaText = ref('')
// invalid JSON in the schema field, found before a request is sent
const inputError = ref('')
const response = ref<FetchResponse>()

const result = computed<FakeResult | undefined>(() => response.value?.isLoading ? undefined : response.value?.data ?? undefined)
const generatorMessage = computed(() => response.value?.error ? generatorError(response.value.error) : '')

function generate() {
    const built = fakeRequest(pathText.value, schemaText.value)
    inputError.value = 'error' in built ? built.error : ''
    response.value = 'error' in built ? undefined : useFetch('/api/faker/fake', fakeRequestInit(built.request), false, false)
}
</script>

<template>
    <form class="faker-playground" @submit.prevent="generate">
        <div class="mb-2">
            <label for="faker-path" class="label">Property path</label>
            <input id="faker-path" v-model="pathText" type="text" class="form-control form-control-sm"
                placeholder="user.email" autocomplete="off" spellcheck="false" />
        </div>
        <div class="mb-2">
            <label for="faker-schema" class="label">Schema</label>
            <textarea id="faker-schema" v-model="schemaText" class="form-control form-control-sm font-monospace" rows="4"
                placeholder='{"type": "string", "format": "email"}' spellcheck="false"></textarea>
        </div>
        <button type="submit" class="btn btn-primary btn-sm">Generate</button>
    </form>

    <div v-if="inputError" class="alert alert-danger mt-2" role="alert">{{ inputError }}</div>
    <div v-else-if="generatorMessage" class="alert alert-danger mt-2" role="alert">{{ generatorMessage }}</div>
    <div v-else-if="result" class="faker-result mt-2">
        <faker-value :value="result.value" />
        <div class="table-responsive">
            <table class="table dataTable faker-matches" aria-label="Matches">
                <thead>
                    <tr>
                        <th>Property path</th>
                        <th>Node</th>
                        <th>Tokens</th>
                        <th>Value</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="match in result.matches ?? []" :key="match.propertyPath">
                        <td><code>{{ match.propertyPath }}</code></td>
                        <td>
                            <button v-if="isTreeNode(match)" type="button" class="btn btn-link btn-sm faker-node-link"
                                @click="$emit('selectNode', match.node)">{{ match.node }}</button>
                            <span v-else>{{ match.node }}</span>
                            <span v-if="sourceNote(match)" class="faker-source">{{ sourceNote(match) }}</span>
                        </td>
                        <td>{{ match.tokens.join(' ') || '-' }}</td>
                        <td><code>{{ JSON.stringify(match.value) }}</code></td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

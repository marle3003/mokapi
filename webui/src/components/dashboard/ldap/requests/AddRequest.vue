<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
    request: LdapAddRequest
}>()

const attributes = computed(() => [...props.request.attributes].sort((x, y) => x.type.localeCompare(y.type)))
</script>

<template>
    <table class="table dataTable" aria-label="Attributes">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 20%">Type</th>
                <th scope="col" class="text-left">Values</th>
            </tr>
        </thead>
        <tbody>
            <!-- an attribute type can be sent more than once -->
            <tr v-for="(attr, attrIndex) of attributes" :key="attrIndex">
                <td>{{ attr.type }}</td>
                <td>
                    <p v-for="(v, vIndex) of attr.values" :key="vIndex">{{ v }}</p>
                </td>
            </tr>
        </tbody>
    </table>
</template>

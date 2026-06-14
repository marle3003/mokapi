<script setup lang="ts">
import type { PropType } from 'vue'
import { usePrettyText } from '@/composables/usePrettyText'

defineProps({
    headers: { type: Object as PropType<KafkaHeader> },
})

const { fromBinary } = usePrettyText()
</script>

<template>
    <table class="table dataTable">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 15%">Name</th>
                <th scope="col" class="text-left">Value</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(header, name) in headers!" :key="name" data-bs-toggle="modal" :data-bs-target="'#modal-'+name">
                <td>
                    <span role="button" @click.stop data-bs-toggle="modal" :data-bs-target="'#modal-'+name" :title="'Read more about header '+name" tabindex="0">
                        {{ name }}
                    </span>
                </td>
                <td>{{ header.value || fromBinary(header.binary) }}</td>
            </tr>
        </tbody>
    </table>
</template>
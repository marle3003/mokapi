<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    topic: { type: Object as PropType<KafkaTopic>, required: true },
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Partitions</div>
            <table class="table dataTable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">ID</th>
                        <th scope="col" class="text-left">Leader</th>
                        <th scope="col" class="text-center">Start Offset</th>
                        <th scope="col" class="text-center">Offset</th>
                        <th scope="col" class="text-center">Segments</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="partition in topic.partitions" :key="partition.id">
                        <td>{{ partition.id }}</td>
                        <td>{{ partition.leader.name }} ({{ partition.leader.addr }})</td>
                        <td class="text-center">{{ partition.startOffset }}</td>
                        <td class="text-center">{{ partition.offset }}</td>
                        <td class="text-center">{{ partition.segments }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
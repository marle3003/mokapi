<script setup lang="ts">
import type { PropType } from 'vue';

defineProps({
    service: { type: Object as PropType<LdapService>, required: true },
})
</script>

<template>
    <section class="card" aria-labelledby="servers">
        <div class="card-body">
            <h2 id="servers" class="card-title text-center">Servers</h2>

            <table class="table dataTable" aria-labelledby="servers">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Address</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>{{ service.server }}</td>
                    </tr>
                </tbody>
            </table>

        </div>
    </section>
</template>
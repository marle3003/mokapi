<script setup lang="ts">
defineProps<{
    request: LdapModifyRequest
}>()
</script>

<template>
    <table class="table dataTable" aria-label="Modifications">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 20%">Modification</th>
                <th scope="col" class="text-left" style="width: 20%">Type</th>
                <th scope="col" class="text-left">Values</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="(item, itemIndex) of request.items" :key="itemIndex">
                <td>{{ item.modification }}</td>
                <td>{{ item.attribute.type }}</td>
                <td><p v-for="(v, vIndex) of item.attribute.values" :key="vIndex">{{ v }}</p></td>
            </tr>
        </tbody>
    </table>
</template>

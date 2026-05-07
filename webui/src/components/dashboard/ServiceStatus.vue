<script setup lang="ts">
const props = defineProps<{
    errors?: Error[]
}>()
</script>

<template>
    <section v-if="errors && errors.length > 0" class="card error" aria-labelledby="errors">
            <div class="card-body">
                <h2 id="errors" class="card-title text-center">Errors</h2>

                <div class="table-responsive-sm">
                    <table class="table dataTable selectable" data-testid="endpoints"  aria-labelledby="errors">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left">Message</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="err of errors">
                                <td>{{ err.message }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </section>
</template>

<style scoped>
.dashboard .card-group .card.error {
    border-color: var(--color-red-shadow) !important;
}
</style>
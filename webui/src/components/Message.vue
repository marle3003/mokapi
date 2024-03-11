<script setup lang="ts">
import { useRouter } from 'vue-router';

defineProps<{
    message: string
    showBackToHome?: boolean
}>()

const router = useRouter()
</script>

<template>
    <div class="position-absolute top-50 start-50 translate-middle">
        <div class="card-group">
            <div class="card"  style="width: 400px">
                <div class="card-body">
                    <div class="row">
                        <div class="col-sm-3">
                            <div class="d-flex flex-wrap align-items-center h-100">
                            <img src="/icon.svg" width="50"/>
                            </div>
                        </div>
                        <div class="col-sm-9 card-text">
                            <div class="d-flex flex-wrap align-items-center h-100 message">
                                {{ message }}
                            </div>
                        </div>
                    </div>
                    <div class="row justify-content-center pt-4" v-if="showBackToHome">
                        <div class="col-auto">
                            <button type="button" class="btn btn-primary btn-sm" @click="router.push({name: 'home'})">Back to Home</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.card {
    border-color: var(--color-border);
    background-color: var(--color-background-soft);
}
.message{
    font-size: 1rem;
    color: var(--color-text);
}
</style>
<script setup lang="ts">
defineProps({
    message: { type: String, required: true },
})
</script>

<template>
    <div class="position-absolute top-50 start-50 translate-middle">
        <div class="card-group">
            <div class="card"  style="width: 400px">
                <div class="card-body">
                    <div class="row">
                        <div class="col-sm-3">
                            <div class="d-flex flex-wrap align-items-center h-100">
                            <img src="/public/icon.svg" width="50"/>
                            </div>
                        </div>
                        <div class="col-sm-9">
                            <div class="d-flex flex-wrap align-items-center h-100 message">
                                {{ message }}
                            </div>
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
}
</style>
<script setup lang="ts">
// The full view of a protocol request: a head card, then the cards Request, Event Handlers and Response.
// The protocol fills the slots; a card whose slot is left out (a request type without response, an
// unknown type, a protocol without handlers) is not shown.
</script>

<template>
    <div class="card-group">
        <div class="card">
            <div class="card-body">
                <slot name="head" />
            </div>
        </div>
    </div>
    <div v-if="$slots.request" class="card-group">
        <section class="card" aria-labelledby="request">
            <div class="card-body">
                <h2 id="request" class="card-title text-center">Request</h2>
                <slot name="request" />
            </div>
        </section>
    </div>
    <div v-if="$slots.handlers" class="card-group">
        <section class="card" aria-labelledby="actions">
            <div class="card-body">
                <h2 id="actions" class="card-title text-center">Event Handlers</h2>
                <slot name="handlers" />
            </div>
        </section>
    </div>
    <div v-if="$slots.response" class="card-group">
        <section class="card" aria-labelledby="response">
            <div class="card-body">
                <h2 id="response" class="card-title text-center">Response</h2>
                <slot name="response" />
            </div>
        </section>
    </div>
</template>

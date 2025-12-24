<script setup lang="ts">
import { computed } from 'vue';


const props = defineProps<{
    settings: MailSettings,
}>()

const maxRecipients = computed(() => {
    if (props.settings.maxRecipients === 0) {
        return 'unlimited'
    }
    return props.settings.maxRecipients
})
</script>

<template>
    <div class="row">
        <div class="col col-2">
            <label id="maxRecipients" class="label">Max Recipients</label>
            <p aria-labelledby="maxRecipients">{{ maxRecipients }}</p>
        </div>
        <div class="col col-2">
            <label id="autoCreateMailbox" class="label">Auto Create Mailbox</label>
            <p aria-labelledby="autoCreateMailbox">{{ settings.autoCreateMailbox }}</p>
        </div>
    </div>
</template>
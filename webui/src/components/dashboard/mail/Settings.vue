<script setup lang="ts">

defineProps<{
    settings: MailSettings,
}>()
</script>

<template>
    <section class="card" data-testid="service-info" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col col-2">
                        <label id="maxRecipients" class="label">Max Recipients</label>
                        <p aria-labelledby="maxRecipients">{{ settings.maxRecipients }}</p>
                    </div>
                    <div class="col col-2">
                        <label id="autoCreateMailbox" class="label">Auto Create Mailbox</label>
                        <p aria-labelledby="autoCreateMailbox">{{ settings.autoCreateMailbox }}</p>
                    </div>
                </div>
            </div>
        </section>
</template>
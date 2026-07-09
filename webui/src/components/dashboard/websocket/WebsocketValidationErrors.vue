<script setup lang="ts">
import { useMarkdown } from '@/composables/markdown';
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { computed, ref } from 'vue';
import SourceView from '../SourceView.vue'
import SchemaExpand from '../SchemaExpand.vue'
import SchemaValidate from '../SchemaValidate.vue'

const props = defineProps<{
    channel?: WebsocketChannel
    validationErrors: Record<string, string>
    value?: string
}>()

const { formatSchema } = usePrettyLanguage()
const messageConfig = ref<WebsocketMessage | null>(null)

const pattern = 'Validation error count ([0-9]+):'

const errors = computed(() => {
    const result = []
    for (const [name, detail] of Object.entries(props.validationErrors)) {
        const m = detail.match(pattern)
        if (!m) {
            continue
        }

        const list = detail.replace(m[0], '').split('\n\t-').filter(x => x !== '')
        result.push({ name, summary: m[0].slice(0, -1), list})
    }
    return result
})

// prevent selection text closes the accordeon
addEventListener('click', function(event: any) {
    const trigger = event.target.closest('[data-bs-toggle="collapse"]');
    if (trigger) {
        // Check if the user has selected any text
        const hasSelection = getSelection()?.toString();
        
        if (hasSelection) {
        // Stop Bootstrap from seeing the click and toggling the collapse
            event.stopPropagation();
            event.preventDefault();
        }
    }
}, true)
</script>

<template>
  <div class="p-0 m-0" style="width: 100%">
  <!-- Table Header -->
  <div class="row header fw-bold border-bottom m-0" style="width: 100%">
    <div class="col-4">Message Id</div>
    <div class="col-8">Error</div>
  </div>

    <!-- Clickable Header Row -->
    <div v-for="err of errors" class="table-row">
        <div
         class="row data align-items-center border-bottom m-0"
         :class="{ 'bg-hover': err.list }"
         :data-bs-toggle="err.list ? 'collapse' : ''" 
         :data-bs-target="'#'+err.name"
         >
        <div class="col-4">{{ err.name }}</div>
        <div class="col-8">{{ err.summary }}</div>
        </div> 
    
        <!-- Collapsible Detail Row -->
        <div v-if="err.list" :id="err.name" class="collapse detail border-bottom" >
            <div class="row p-3 pt-1">
                <div class="col-12">
                    <ul class="mb-0 ps-0">
                        <li v-for="s in err.list"><span>{{ s }}</span></li>
                    </ul>
                </div>
            </div>
            <div v-if="channel" :set="messageConfig = channel?.messages[err.name]" class="mb-2">
                <div v-if="messageConfig">
                    <source-view :source="{ preview: {content: formatSchema(messageConfig.payload), contentType: 'application/json' } }" :hide-content-type="true" :height="500" class="mb-2" :filename="channel.name+'-message.json'" />
                    <div class="d-flex gap-2" role="group">
                        <schema-expand :schema="messageConfig.payload" :title="'Value - '+channel.name" :source="{filename: channel.name+'-message.json'}" />
                        <schema-validate :title="'Value Validator - '+channel.name" :schema="messageConfig.payload" :source="{preview: { content: value ?? '', contentType: messageConfig.contentType, contentTypeTitle: messageConfig.contentType, description: '' }, }" :example-enabled="false"/>
                    </div>
                </div>
            </div>
        </div>
    </div>

</div>

</template>

<style scoped>
.header, .data {
        border-bottom-width: 2px !important;
    }
    .header > div, .data > div, .detail {
        padding: 3px 0 3px 12px;
    }
    .bg-hover:hover {
        background-color: var(--datatable-background-active);
        cursor: pointer;
    }
    .data[aria-expanded="true"] {
        border-bottom-style: none !important;
    }
    .detail {
        border-bottom-width: 2px !important;

        ul li span{
            color: var(--color-red);
        }
    }
    .table-row:last-child .detail {
        border-bottom-style: none !important;
    }
</style>
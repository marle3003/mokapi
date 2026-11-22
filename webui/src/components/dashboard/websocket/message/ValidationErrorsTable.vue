<script setup lang="ts">
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { computed, onMounted, onUnmounted } from 'vue';
import SourceView from '../../SourceView.vue'
import SchemaExpand from '../../SchemaExpand.vue'
import SchemaValidate from '../../SchemaValidate.vue'

const props = defineProps<{
    channel?: WebsocketChannel
    validationErrors: Record<string, string>
    value?: string
}>()

const { formatSchema } = usePrettyLanguage()

const pattern = 'Validation error count ([0-9]+):'

const errors = computed(() => {
    const result = []
    for (const [name, detail] of Object.entries(props.validationErrors)) {
        const m = detail.match(pattern)
        if (!m) {
            continue
        }

        const list = detail.replace(m[0], '').split('\n\t-').filter(x => x !== '')
        const safeId = 'err-' + name.replace(/[^a-zA-Z0-9-_]/g, '-')
        const messageConfig = props.channel?.messages[name]

        result.push({ name, safeId, summary: m[0].slice(0, -1), messageConfig: messageConfig, list })
    }
    return result
})

// prevent selection text closes the accordeon
function keepAccordionOpenWhileSelectingText(event: MouseEvent) {
    const trigger = (event.target as Element).closest('[data-bs-toggle="collapse"]');
    if (trigger) {
        const hasSelection = getSelection()?.toString();

        if (hasSelection) {
            event.stopPropagation();
            event.preventDefault();
        }
    }
}
// the log panel mounts the table for every selected message, so the listener must not outlive it
onMounted(() => addEventListener('click', keepAccordionOpenWhileSelectingText, true))
onUnmounted(() => removeEventListener('click', keepAccordionOpenWhileSelectingText, true))
</script>

<template>
    <div class="p-0 m-0" role="table" aria-label="Validation Errors">
        <!-- Table Header -->
        <div class="rowgroup">
            <div class="row header fw-bold border-bottom m-0" role="row">
                <div class="col-4" role="columnheader">Message Id</div>
                <div class="col-8" role="columnheader">Error</div>
                <div class="col-2" role="columnheader"></div>
            </div>
        </div>

        <!-- Clickable summary Row -->
        <div v-for="err of errors" :key="err.safeId" class="table-row" role="rowgroup">
            <div class="row data align-items-center border-bottom m-0" :class="{ 'bg-hover': err.list }" role="row"
                :data-bs-toggle="err.list ? 'collapse' : ''" :data-bs-target="'#' + err.safeId" :aria-expanded="false"
                :aria-controls="err.safeId">
                <div class="col-5" role="cell">{{ err.name }}</div>
                <div class="col-5" role="cell">{{ err.summary }}</div>
                 <div class="col-2 text-end pe-3" role="cell" aria-hidden="true">
                    <span class="bi bi-chevron-down chevron"></span>
                </div>
            </div>

            <!-- Collapsible Detail Row -->
            <div v-if="err.list" :id="err.safeId" class="collapse detail border-bottom">
                <div class="row p-3 pt-1">
                    <div class="col-12">
                        <!-- Error list -->
                        <ul class="error-list mb-0 ps-0" :aria-label="`Validation errors for ${err.name}`">
                            <li v-for="(s, i) in err.list" :key="i" class="error-item">
                                <span>{{ s }}</span>
                            </li>
                        </ul>
                    </div>
                </div>
                <div v-if="channel" class="mb-2">
                    <div v-if="err.messageConfig">
                        <div class="schema-label mb-2">
                            Expected Schema
                        </div>
                        <source-view
                            :source="{ preview: { content: formatSchema(err.messageConfig.payload), contentType: 'application/json' } }"
                            :hide-content-type="true" :height="500" class="mb-2"
                            :filename="channel.name + '-message.json'" />
                        <div class="d-flex gap-2" role="group">
                            <schema-expand :schema="err.messageConfig.payload" :title="'Value - ' + channel.name"
                                :source="{ filename: channel.name + '-message.json' }" />
                            <schema-validate :title="'Value Validator - ' + channel.name"
                                :schema="err.messageConfig.payload"
                                :source="{ preview: { content: value ?? '', contentType: err.messageConfig.contentType, contentTypeTitle: err.messageConfig.contentType, description: '' }, }"
                                :example-enabled="false" />
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

.header>div, .data>div, .detail {
    padding: 3px 0 3px 12px;
}

.bg-hover:hover {
    cursor: pointer;
    transition: background-color 0.15s ease;
}

.bg-hover:hover {
    background-color: var(--datatable-background-active);
}
.chevron {
    display: inline-block;
    transition: transform 0.2s ease !important;
    color: var(--color-text-light);
    font-size: 0.75rem;
}
.data[aria-expanded="true"] .chevron {
    transform: rotate(180deg) !important;
}

.data[aria-expanded="true"] {
    border-bottom-style: none !important;
}

.detail {
    border-bottom-width: 2px !important;

    ul li {
        color: var(--color-red);
    }
    ul li span {
        color: var(--color-red);
    }
}

.table-row:last-child .detail {
    border-bottom-style: none !important;
}


</style>
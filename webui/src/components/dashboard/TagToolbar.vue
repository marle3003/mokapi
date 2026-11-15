<script setup lang="ts">
import { toggleTag, type TagGrouping, type TagInfo } from '@/composables/tag-group'
import TagGroupSwitch from '@/components/dashboard/TagGroupSwitch.vue'

const props = defineProps<{
    tags: TagInfo[],
    noun: 'paths' | 'topics' | 'channels',
}>()

const selected = defineModel<string[]>('selected', { required: true })
const grouping = defineModel<TagGrouping>('grouping', { required: true })

function toggle(name: string) {
    selected.value = toggleTag(selected.value, name, props.tags.map(t => t.name))
}
</script>

<template>
    <div v-if="tags.length > 0" class="d-flex flex-wrap align-items-start justify-content-between gap-2 mt-3 mb-2">
        <fieldset v-if="tags.length > 1" aria-describedby="tag-help">

            <legend class="visually-hidden">
                Filter {{ noun }} by tags
            </legend>

            <p id="tag-help" class="visually-hidden">
                Select one or more tags to filter the {{ noun }}. Selecting "All" enables all tags.
            </p>

            <div class="form-check form-check-inline">
                <input class="form-check-input tag-checkbox" type="checkbox" id="all" value="all"
                    @change="toggle('__all')" :checked="selected.includes('__all')" aria-controls="tag-list">
                <label class="form-check-label" for="all">All</label>
            </div>

            <div id="tag-list" style="display: inline-block">
                <div class="form-check form-check-inline" v-for="tag in tags" :key="tag.name" :title="tag.description">
                    <input class="form-check-input tag-checkbox" type="checkbox" :id="tag.name" :value="tag.name"
                        @change="toggle(tag.name)" :checked="selected.includes(tag.name) || selected.includes('__all')"
                        :aria-describedby="tag.description ? `desc-${tag.name}` : undefined">
                    <label class="form-check-label" :for="tag.name">{{ tag.name }}</label>

                    <span v-if="tag.description" :id="`desc-${tag.name}`" class="visually-hidden">
                        {{ tag.description }}
                    </span>
                </div>
            </div>

        </fieldset>
        <tag-group-switch v-model="grouping" class="ms-auto flex-shrink-0" />
    </div>
</template>

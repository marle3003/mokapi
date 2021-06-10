<template>
  <b-tabs content-class="mt-3 ml-2" class="responses" align="left">
    <b-tab v-for="response in sorted" :key="response.status">
      <template v-slot:title>
        <span v-if="response.status >= 200 && response.status < 300" class="success">
          <b-icon icon="circle-fill" class="icon mr-1" ></b-icon>
          {{ response.status }}
        </span>
        <span v-if="response.status >= 300 && response.status < 400" class="warning">
          <b-icon icon="circle-fill" class="icon mr-1" ></b-icon>
          {{ response.status }}
        </span>
        <span v-if="response.status >= 400 && response.status < 500" class="client-error">
          <b-icon icon="circle-fill" class="icon mr-1" ></b-icon>
          {{ response.status }}
        </span>
        <span v-if="response.status >= 500 && response.status < 600" class="danger">
          <b-icon icon="circle-fill" class="icon mr-1" ></b-icon>
          {{ response.status }}
        </span>
      </template>
      <p class="label">Description</p>
      <p>{{ response.description }}</p>
      <p class="label">Content Type</p>
      <div v-if="response.contentTypes != null && response.contentTypes.length === 1">
          <div v-for="content in response.contentTypes" :key="content.type">
            <p>{{ content.type }}</p>
            <p v-if="content.schema != null" class="label">Schema</p>
            <schema v-if="content.schema != null" v-bind:schema="content.schema"></schema>
          </div>
      </div>
      <div v-if="response.contentTypes != null && response.contentTypes.length > 1">
        <p>
          <b-form-select v-model="selected[response.status]">
            <b-form-select-option
              v-for="content in response.contentTypes" :key="content.type" :value="content.type">
              {{ content.type }}
            </b-form-select-option>
            </b-form-select>
        </p>
        <div v-for="content in response.contentTypes" :key="content.type">
          <div v-if="selected[response.status] === content.type">
            <p class="label">Schema</p>
            <schema v-bind:schema="content.schema"></schema>
          </div>
        </div>
      </div>
    </b-tab>
  </b-tabs>
</template>

<script>
import Schema from '@/components/Schema'

export default {
  name: 'response',
  components: {'schema': Schema},
  props: ['responses'],
  data () {
    return {
      selected: {}
    }
  },
  created () {
    for (let r of this.responses) {
      if (r.contentTypes !== null) {
        this.selected[r.status] = r.contentTypes[0].type
      }
    }
  },
  computed: {
    sorted: function () {
      if (!this.responses) {
        return []
      }

      function compare (a, b) {
        return a.status - b.status
      }

      // eslint-disable-next-line vue/no-side-effects-in-computed-properties
      return this.responses.sort(compare)
    }
  }
}
</script>

<style scoped>
  .responses .icon{
      vertical-align: middle;
      font-size: 0.5rem;
  }
  .custom-select {
    background: var(--var-bg-color-primary) url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' width='4' height='5' viewBox='0 0 4 5'%3e%3cpath fill='%23d3d4d5' d='M2 0L0 2h4zm0 5L0 3h4z'/%3e%3c/svg%3e") no-repeat right .75rem center/8px 10px !important;
    color: var(--var-color-primary);
    border-color: var(--var-border-color);
    padding: 0;
    height: 1.5rem;
  }
  .custom-select:hover{
    background: var(--var-bg-color-secondary) url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' width='4' height='5' viewBox='0 0 4 5'%3e%3cpath fill='%23d3d4d5' d='M2 0L0 2h4zm0 5L0 3h4z'/%3e%3c/svg%3e") no-repeat right .75rem center/8px 10px !important;
  }
</style>

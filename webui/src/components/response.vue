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
      <div v-if="response.contentTypes != null && response.contentTypes.length === 1">
          <div v-for="content in response.contentTypes" :key="content.type">
            <p class="label">Content Type</p>
            <p>{{ content.type }}</p>
            <p v-if="content.schema != null" class="label">Schema</p>
            <schema v-if="content.schema != null" v-bind:schema="content.schema"></schema>
          </div>
      </div>
      <b-card v-if="response.contentTypes != null && response.contentTypes.length > 1">
        <template #header>
          <b-dropdown>
            <b-dropdown-item  v-for="content in response.contentTypes" :key="content.type">
              {{ content.type }}
            </b-dropdown-item>
          </b-dropdown>
        </template>
        <b-card-text v-for="content in response.contentTypes" :key="content.type">
          <p class="label">Schema</p>
          <schema v-bind:schema="content.schema"></schema>
        </b-card-text>
      </b-card>
    </b-tab>
  </b-tabs>
</template>

<script>
import Schema from '@/components/Schema'

export default {
  name: 'response',
  components: {'schema': Schema,},
  props: ['responses'],
  computed: {
    sorted: function() {
      if (!this.responses) {
        return []
      }
      
      function compare(a, b) {
        return a.status - b.status
      }

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
</style>

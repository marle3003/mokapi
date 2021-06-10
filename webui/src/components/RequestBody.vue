<template>
  <div v-if="current !== null">
    <hr />
    <h5>Request Body</h5>
    <p class="label">Description</p>
    <p>{{ current.description }}</p>
    <p class="label">Content Type</p>
    <p v-if="options.length <= 1"> {{selected}} </p>
    <b-col v-else xs="auto" sm="8" md="6" lg="3" xl="3" class="pl-0">
    <p><b-form-select v-model="selected" :options="options" size="sm"></b-form-select></p>
    </b-col>
    <p class="label">Schema</p>
    <schema v-bind:schema="current.schema"></schema>
  </div>
</template>

<script>
import Schema from '@/components/Schema'

export default {
  name: 'parameters',
  components: {'schema': Schema},
  props: ['operation'],
  data () {
    return {
      selected: ''
    }
  },
  computed: {
    current: function () {
      if (this.operation.requestBodies && this.operation.requestBodies.length > 0) {
        return this.operation.requestBodies[0]
      }
      return null
    },
    options: function () {
      let options = []
      if (this.operation.requestBodies) {
        this.operation.requestBodies.forEach(x => options.push(x.contentType))
      }
      return options
    }
  },
  methods: {
  },
  mounted: function () {
    if (this.options.length > 0) {
      this.selected = this.options[0]
    }
  }
}
</script>

<style scoped>
  .label{
    color: #a0a1a7;
    margin-bottom: 0
  }
</style>

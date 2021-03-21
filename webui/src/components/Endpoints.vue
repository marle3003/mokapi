<template>
    <b-table hover :items="endpoints" :fields="fields" table-class="dataTable endpoints" @row-clicked="routerLinkToEndpoint">
      <template v-slot:cell(summary)="data">
        <vue-simple-markdown :source="data.value" />
      </template>
      <template v-slot:cell(operations)="data">
        <span v-for="( operation, index ) in data.value" :key="index" class="mr-1">
          <b-badge pill class="operation" :class="operation.method" >{{ operation.method }}</b-badge>
        </span>
      </template>
    </b-table>
</template>

<script>

export default {
  name: 'endpoints',
  props: ['service'],
  data() {
    return {
      fields: ['path', 'summary', 'operations'],
    }
  },
  computed: {
    endpoints: function () {
      if (this.service === null){
        return []
      }

      function compare(a, b) {
        if (a.path < b.path) {
          return -1
        }
        if (a.path > b.path) {
          return 1
        }
        return 0
      }

      return this.service.endpoints.sort(compare)
    }
  },
  methods: {
    routerLinkToEndpoint (item, index, event) {
      this.$router.push({ name: 'endpoint', params: {path: item.path} })
    }
  }
}
</script>

<style>
  .endpoints tbody{
    cursor: pointer;
  }
  .endpoints a:hover{
    color: var(--var-color-primary)
  }
</style>

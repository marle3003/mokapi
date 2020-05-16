<template>
  <b-card title="Endpoints" class="w-100">
    <b-table striped hover :items="items" :fields="fields" @row-clicked="routerLinkToEndpoint">
        <template v-slot:cell(operations)="data">
            <span v-for="( operation, index ) in data.value" :key="index" class="mr-1">
            <b-badge pill :variant="operation.variant" >{{ operation.label }}</b-badge>
            </span>
        </template>
    </b-table>
  </b-card>
</template>

<script>
export default {
    name: "endpoints",
    components: {},
    data() {
      return {
        fields: ['path', 'operations'],
        items: [
          { path: "/users", operations: [{label: "GET", variant: "primary"}, {label: "POST", variant: "success"}] },
          { path: "/users/{id}", operations:  [{label: "GET", variant: "primary"}, {label: "DELETE", variant: "danger"}] },
        ]
      }
    },
    methods:{
      routerLinkToEndpoint(item, index, event){
        this.$router.push({ name: 'endpoint', params: {path: item.path} });
      }
    }
}
</script>

<style scoped>
  .card-title{
      font-size: 1.2rem;
  }
  .operation{
      font-size: 1rem;
  }
</style>

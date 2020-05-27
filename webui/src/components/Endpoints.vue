<template>
  <b-card class="w-100">
    <b-card-title>
      <img src="@/assets/endpoint.png" width="24" height="24" /> Endpoints
    </b-card-title>
    <b-table hover :items="endpoints" :fields="fields" tbody-class="operations" @row-clicked="routerLinkToEndpoint">
        <template v-slot:cell(operations)="data">
          <span v-for="( operation, index ) in data.value" :key="index" class="mr-1 operation">
            <b-badge pill :class="operation.method" >{{ operation.method }}</b-badge>
          </span>
        </template>
    </b-table>
  </b-card>
</template>

<script>
export default {
    name: "endpoints",
    props: ["service"],
    components: {},
    data() {
      return {
        fields: ['path', 'operations'],
      }
    },
    computed: {
      endpoints: function () {
        if (this.service == null){
          return [];
        }
        
        function compare(a, b) {
          if (a.path < b.path)
            return -1;
          if (a.path > b.path)
            return 1;
          return 0;
        }

        return  this.service.endpoints.sort(compare);
      }
    },
    methods:{
      routerLinkToEndpoint(item, index, event){
        this.$router.push({ name: 'endpoint', params: {path: item.path} });
      }
    }
}
</script>

<style>
  .operations{
    cursor: pointer;
  }
</style>

<style scoped>
  .card-title{
      font-size: 1.2rem;
  }
  .operation{
      font-size: 1rem;
  }
</style>

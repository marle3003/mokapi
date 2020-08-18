<template>
  <b-card class="w-100">
    <b-card-title>
      <img src="@/assets/models.png" width="24" height="24" /> Models
    </b-card-title>
    <b-table hover :items="models" :fields="fields" tbody-class="models" @row-clicked="routerLinkToEndpoint">
        <template v-slot:cell(operations)="data">
          <span v-for="( operation, index ) in data.value" :key="index" class="mr-1 model">
            <b-badge pill :class="operation.method" >{{ operation.method }}</b-badge>
          </span>
        </template>
    </b-table>
  </b-card>
</template>

<script>
export default {
    name: "models",
    props: ["service"],
    components: {},
    data() {
      return {
        fields: ['path', 'models'],
      }
    },
    computed: {
      models: function () {
        if (this.service == null || this.service.models == undefined){
          return [];
        }
        
        function compare(a, b) {
          if (a.path < b.path)
            return -1;
          if (a.path > b.path)
            return 1;
          return 0;
        }

        return  this.service.models.sort(compare);
      }
    },
    methods:{
      routerLinkToEndpoint(item, index, event){
        this.$router.push({ name: 'model', params: {path: item.path} });
      }
    }
}
</script>

<style>
  .models{
    cursor: pointer;
  }
</style>

<style scoped>
  .card-title{
      font-size: 1.2rem;
  }
  .model{
      font-size: 1rem;
  }
</style>

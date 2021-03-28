<template>
  <b-table hover :items="models" :fields="fields" class="dataTable" tbody-class="models">
    <template v-slot:cell(properties)="data">
        <div v-for="( property, index ) in data.value" :key="index" class="mr-1">
          {{property.name}}: <span class="type">{{property.type}}</span>
        </div>
      </template>
  </b-table>
</template>

<script>
export default {
    name: "models",
    props: ['service'],
    components: {},
    data() {
      return {
        fields: ['name', 'type', 'properties'],
      }
    },
    computed: {
      models: function () {
        if (this.service == null || this.service.models == undefined){
          return [];
        }
        
        function compare(a, b) {
          return (''+a.name).localeCompare(''+b.name)
        }

        this.service.models.forEach(x => {
          if (x.properties != null) { 
            x.properties.sort(compare)
          }
        })

        return  this.service.models.sort(compare);
      }
    }
}
</script>

<style>
  .models td{
    padding-top: 0.2rem
  }
</style>

<style scoped>
  .model{
      font-size: 1rem;
  }
  .type{
    color: var(--var-code-orange);
  }
</style>

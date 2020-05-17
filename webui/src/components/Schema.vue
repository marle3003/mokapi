<template>
  <b-table ref="table" small :items="itemsProvider" :fields="fields" class="schema" thead-class="hidden_header">
    <!-- Use it for editing <template v-slot:cell(add)="data">
      <div v-b-hover="handleHover" v-if="data.item.type === 'object'" v-on:click="addProperty">
        <b-icon icon="plus" v-if="isHovered" class="hover"></b-icon>
        <b-icon icon="plus" v-else></b-icon>
      </div>
    </template> -->
    <template v-slot:cell(type)="data">
      <span v-if="data.item.rowType === 'root'">{{ data.item.type }}</span>
      <span v-else v-bind:style="{ paddingLeft: data.item.level * 18 + 'px'}">{{ data.item.name}}: {{data.item.type}}</span>
    </template>
  </b-table>
</template>

<script>
export default {
    name: 'schema',
    props: ['schema'],
    data() {
      return {
        fields: [ 
          //{key: 'add', tdClass: "schema_add"},
          {key: 'type' }],
        isHovered: false,
      }
    },
    methods:{
      itemsProvider(){
        return this.getItems(this.schema, 0)
      },
      addProperty(){
        this.items.push({type: 'string'})
        this.$refs.table.refresh();
      },
      handleHover(hovered) {
        this.isHovered = hovered
      },
      getItems(schema, level){
        let items = []

        if (schema.name == undefined && level === 0) {
          items.push({type: schema.type, name: '', rowType: 'root'})
        }else{
          items.push({type: schema.type, name: schema.name, rowType: 'property', level: level})
        }
        
        if (schema.type === 'object' && level < 10 && schema.properties != undefined){
          for (let i = 0; i < schema.properties.length; i++){
            let p = schema.properties[i]

            items = items.concat(this.getItems(p, level + 1))
          }
        }

        return items;
      }
    }
}
</script>

<style>
.schema .hidden_header{
    display: none;
}
.schema .schema_add{
  width: 20px;
}
</style>
<style scoped>
.hover{
background-color: rgba(0, 0, 0, 0.1);
}
</style>
<template>
  <b-table ref="table" small :items="itemsProvider" :fields="fields" class="schema" thead-class="hidden_header">
    <!-- Use it for editing <template v-slot:cell(add)="data">
      <div v-b-hover="handleHover" v-if="data.item.type === 'object'" v-on:click="addProperty">
        <b-icon icon="plus" v-if="isHovered" class="hover"></b-icon>
        <b-icon icon="plus" v-else></b-icon>
      </div>
    </template> -->
    <template v-slot:cell(type)="data">
      <span v-bind:style="{ paddingLeft: data.item.level * 18 + 'px'}">
        {{ data.item.text}}
      </span>
      <span v-if="data.item.ref != undefined && data.item.ref !== ''">
        ({{ data.item.ref }})
      </span>
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

        var item = {}

        if (level === 0) {
          item = {type: schema.type, level: 0, text: schema.type, ref: schema.ref}
        }else{
          var text = schema.type
          if (schema.name != undefined && schema.name !== ''){
            text = schema.name + ': ' + schema.type
          }
          item = {type: schema.type, level: level, text: text, ref: schema.ref}
        }          

        items.push(item)

        if (schema.type === 'array'){
          var itemType = schema.items.ref !== '' ? schema.items.ref : schema.items.type;
          item['text'] = 'array['+itemType+']'
          item['refText'] = schema.items.ref

          // get all properties but not type => not incrementing level because we remove first level (shift)
          var arrayItems = this.getItems(schema.items, level)
          arrayItems.shift() // remove object type
          items = items.concat(arrayItems)
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
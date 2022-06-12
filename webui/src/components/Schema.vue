<template>
<div>
  <div v-for="(row, index) in rows" :key="index" >
    <b-icon :style="{visibility: row.type === 'prop' ? 'visible' : 'hidden'}" :icon="isOpened(index) ? 'dash-square' : 'plus-square'" font-scale="0.7" @click="toggleDetails(index)" />
    <pre style="display:inline"><code><span v-html="row.code" /></code></pre>
    <b-card v-show="isOpened(index)" class="w-100">
          <b-row class="">
            <b-col>
              <p class="label">Required</p>
              <p>{{ row.required }}</p>
               <p class="label">Format</p>
              <p>{{ row.format }}</p>
            </b-col>
            <b-col>
               <p class="label">Nullable</p>
              <p>{{ row.nullable }}</p>
            </b-col>
          </b-row>
          <b-row>
             <b-col>
              <div v-if="row.description != null">
                <p class="label">Description</p>
                <vue-simple-markdown :source="row.description" />
              </div>
            </b-col>
          </b-row>
        </b-card>
    </div>
</div>
</template>

<script>
export default {
  name: 'schema',
  props: ['schema'],
  data () {
    return {
      fields: [
        // {key: 'add', tdClass: "schema_add"},
        { key: 'type' }],
      isHovered: false,
      opened: []
    }
  },
  computed: {
    rows: function () {
      return this.getRows(this.schema, 0, [])
    }
  },
  methods: {
    isOpened (index) {
      return this.opened.indexOf(index) >= 0
    },
    toggleDetails (index) {
      const i = this.opened.indexOf(index)
      if (i >= 0) {
        this.opened.splice(i, 1)
      } else {
        this.opened.push(index)
      }
    },
    itemsProvider () {
      return this.getItems(this.schema, 0)
    },
    addProperty () {
      this.items.push({type: 'string'})
      this.$refs.table.refresh()
    },
    handleHover (hovered) {
      this.isHovered = hovered
    },
    getRefOrEmpty (ref) {
      if (!ref) {
        return ''
      }
      return '<span class="ref-type">[' + ref + ']</span>'
    },
    truncate (str, n) {
      if (!str) return ''
      return (str.length > n) ? str.substr(0, n - 1) + '&hellip;' : str
    },
    getRows (schema, level, paths) {
      let rows = []
      if (!schema) {
        return rows
      }

      let row = {
        format: schema.format,
        description: schema.description,
        required: false,
        nullable: schema.nullable
      }
      row.format = schema.format

      if (schema.ref) {
        if (paths.indexOf(schema.ref) >= 0) {
          return rows
        }
        paths.push(schema.ref)
      }

      let ident = ' '.repeat(2 * level)
      if (schema.type === 'object') {
        row.code = ident
        if (schema.name) {
          row.code += schema.name + ': '
        }
        row.code += this.getRefOrEmpty(schema.ref) + '{'
        rows.push(row)

        for (let i = 0; i < schema.properties.length; i++) {
          let propRows = this.getRows(schema.properties[i], level + 1, paths)
          if (propRows.length > 0 && schema.required) {
            propRows[0].required = schema.required.indexOf(schema.properties[i].name)
          }
          console.log(propRows)
          rows = rows.concat(propRows)
        }
        rows.push({code: ident + '}'})
      } else if (schema.type === 'array') {
        let s = ident
        if (schema.name) {
          s += schema.name + ': '
        }
        if (schema.items !== undefined && schema.items !== null) {
          s += this.getRefOrEmpty(schema.items.ref)
        }
        s += ' ['
        // get all properties but not type => not incrementing level because we remove first level (shift)
        var arrayRows = this.getRows(schema.items, level, paths)
        arrayRows.shift() // remove object type
        arrayRows.pop()
        if (arrayRows.length === 0) {
          rows.push({code: s + '...]'})
        } else {
          rows.push({code: s})
          rows = rows.concat(arrayRows)
          rows.push({code: ident + ']'})
        }
      } else {
        let s = '<span class="prop-type">' + schema.type + '</span>'
        if (schema.name) {
          s = schema.name + ': ' + s
        }
        if (schema.description) {
          s += '<span class="comment"> (' + this.truncate(schema.description, 50) + ')</span>'
        }
        row.code = ident + s
        row.type = 'prop'
        rows.push(row)
      }

      return rows
    },
    getItems (schema, level) {
      let items = []

      try {
        let item = {}

        if (level === 0) {
          item = {type: schema.type, level: 0, text: schema.type, ref: schema.ref}
        } else {
          var text = schema.type
          if (schema.name !== undefined && schema.name !== '') {
            text = schema.name + ': ' + schema.type
          }
          item = {type: schema.type, level: level, text: text, ref: schema.ref}
        }

        items.push(item)

        if (schema.type === 'array') {
          if (typeof schema.items === 'undefined') {
            console.error('field items is undefined in schema type array')
            return items
          }

          let itemType = schema.items.ref !== undefined ? schema.items.ref : schema.items.type
          item['text'] = schema.name + ': array[' + itemType + ']'
          item['refText'] = schema.items.ref

          // get all properties but not type => not incrementing level because we remove first level (shift)
          var arrayItems = this.getItems(schema.items, level)
          arrayItems.shift() // remove object type
          items = items.concat(arrayItems)
        }

        if (schema.type === 'object' && level < 2 && schema.properties !== undefined) {
          for (let i = 0; i < schema.properties.length; i++) {
            let prop = this.getItems(schema.properties[i], level + 1)
            items = items.concat(prop)
          }
        }
      } catch (e) {
        console.log(e)
      }
      return items
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
  .prop-type {
    color: var(--var-code-orange);
  }
  .ref-type{
    color: var(--purple);
  }
  .comment {
    color: var(--var-code-green);
  }
</style>
<style scoped>
  .hover{
    background-color: rgba(0, 0, 0, 0.1);
  }
</style>

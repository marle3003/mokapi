<template>
  <div class="doc">
    <div class="page-body">
      <b-card>
        <b-row>
          <b-col cols="2">
            <b-list-group>
              <b-list-group-item class="menu-item title border-0">Mokapi</b-list-group-item>
              <div v-for="item in navigation" :key="item.label">
                <b-list-group-item
                  :to="item.url != undefined ? { path: item.url} : null"
                  class="menu-item"
                  v-b-toggle="item.label.toLowerCase()">
                  {{item.label}}
                  </b-list-group-item>
                <b-collapse v-if="item.children != undefined" style="padding:0.4rem" :id="item.label.toLowerCase()" v-model="item.isOpen">
                  <b-list-group-item
                    v-for="child in item.children"
                    :to="{ path: child.url}"
                    class="menu-item"
                    :key="child.label">{{child.label}}</b-list-group-item>
                </b-collapse>
              </div>
            </b-list-group>
          </b-col>
          <b-col>
            <component v-if="doc != null" v-bind:is="doc.obj.vue.component" class="content"></component>
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
import config from '@/assets/docs/config.yml'

export default {

  data () {
    return {
      files: [],
      navigation: [],
      doc: null
    }
  },
  watch: {
    $route (to, from) {
      this.init()
    }
  },
  mounted () {
    this.importAll(require.context('@/assets/docs/', true, /\.md$/))
    this.navigation = this.parseNavigation(config.nav)
    this.init()
  },
  methods: {
    parseNavigation (list) {
      var nav = []
      for (var index in list) {
        var item = list[index]
        var properties = Object.keys(item)
        var value = item[properties[0]]
        var navItem = {label: properties[0], isOpen: false}
        if (Array.isArray(value)) {
          navItem['children'] = this.parseNavigation(value)
        } else {
          navItem['path'] = value
          navItem['url'] = '/docs/' + value.substring(0, value.length - 3)
        }
        nav.push(navItem)
      }
      return nav
    },
    setOpen (path) {
      var navItem = this.navigation.find(x => x.children !== undefined && this.hasChild(path, x.children))
      if (navItem !== undefined) {
        navItem.isOpen = true
      }
    },
    hasChild (path, list) {
      var navItem = list.find(x => x.path === path)
      return navItem !== undefined
    },
    init () {
      this.resetNavigation()

      let path = this.$route.params.topic
      let subject = this.$route.params.subject
      if (subject !== undefined) {
        path += '/' + subject
      }
      path = path.toLowerCase() + '.md'

      this.doc = this.files.find(x => x.path.toLowerCase() === path)
      this.setOpen(path)
    },
    resetNavigation () {
      for (var index in this.navigation) {
        this.navigation[index].isOpen = false
      }
    },
    importAll (r) {
      r.keys().forEach(key => {
        let v = r(key)
        // key ./index.md
        this.files.push({path: key.substring(2), obj: v})
      })
    }
  }
}
</script>

<style scoped>
.doc {
  width: 90%;
  margin: auto;
  margin-top: 42px;
}
.card {
  margin: 15px;
}
.card p {
  margin-bottom: 0;
}
.menu-item {
  border: 0;
  padding: 0.2rem;
}
.menu-item:focus {
  outline: none;
  box-shadow: none;
}
.menu-item.title {
  font-weight: 500;
}
.content {
  line-height: 1.6;

}
.router-link-active {
  color: #007bff;
  background-color: #D0E6FF;
  font-weight: 500;
  border-radius: 8px;
}
a:hover{
  background-color: #D0E6FF;
  border-radius: 8px;
}

/* markdown deep selector */
.content >>> h1 {
  font-size: 1.5rem;
  font-weight: bold;
  margin: 0 0 2rem;
}
.content >>> h2 {
  font-size: 1.25rem;
  margin: 1.5rem 0 0.5rem;
}
</style>

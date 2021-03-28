<template>
  <div class="doc">
    <div class="page-body">

        <b-row>
          <b-col cols="2">
            <b-navbar sticky>
            <b-list-group class="w-100">
              <div v-for="item in navigation" :key="item.label" class="w-100">
                <b-list-group-item
                  :to="item.url != undefined ? { path: item.url} : null"
                  class="menu-item"
                  v-b-toggle="item.label.toLowerCase()">
                  <span style="display: flex; flex-grow: 1">{{item.label}}</span>
                  <span v-if="item.children != undefined">
                    <span class="icon">
                      <b-icon icon="chevron-right"></b-icon>
                    </span>
                  </span>
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
            </b-navbar>
          </b-col>
          <b-col>
            <component v-if="doc != null" v-bind:is="doc.obj.vue.component" class="content"></component>
          </b-col>
        </b-row>

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
.col{
  padding-left: 0;
}
.col-2{
  padding-left: 0;
}
.navbar {
  padding: 0;
}
.menu-item {
  border: 0;
  padding: 0.2rem;
  padding-right: 1.5rem;
  display: flex;
}
.menu-item svg {
  -moz-transition: all .3s linear;
  -webkit-transition: all .3s linear;
  transition: all .3s linear;
}
.not-collapsed svg {
  -moz-transform:rotate(90deg);
  -webkit-transform:rotate(90deg);
  transform:rotate(90eg);
}
.content {
  line-height: 1.6;
  font-size: 0.8rem;
}
.router-link-active {
  font-weight: 500;
}
/* markdown deep selector */
.content >>> h1 {
  font-size: 1.75rem;
  font-weight: 600;
  margin: 0 0 1.3rem;
}
.content >>> h2 {
  font-size: 1.25rem;
  font-weight: 500;
  margin: 1.5rem 0 0.5rem;
}
.content >>> h3 {
  font-size: 1.1rem;
  font-weight: 500;
  margin: 1.5rem 0 0.5rem;
}
</style>

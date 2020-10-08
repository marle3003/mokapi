<template>
  <div class="doc">
    <div class="page-body">
      <b-card>
        <b-row>
          <b-col cols="2">
            <b-list-group>
              <b-list-group-item class="menu-item title">Mokapi</b-list-group-item>
              <div v-for="(item, index) in navigation" :key="item.text">
                <b-list-group-item 
                  :href="item.href"
                  class="menu-item"

                  v-b-toggle="item.text.toLowerCase()">
                  {{item.text}}
                  </b-list-group-item>
                <b-collapse style="padding:0.4rem" :id="item.text.toLowerCase()" v-model="topicState[getNavigationIndex(index)]">
                <b-list-group-item
                  v-for="child in item.children"
                  :to="{ path: child.href}"
                  class="menu-item"
                  :key="child.text">{{child.text}}</b-list-group-item>
                </b-collapse>
              </div>
            </b-list-group>
          </b-col>
          <b-col>
            <component v-if="current != null" v-bind:is="current.obj.vue.component" class="content"></component>
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
  import prism from '@/assets/prism.js'
  import styles from '@/assets/prism.css'

  export default {

    data () {
      return {
        files: [],
        navigation: {},
        current: null,
        topicState: [],
        active: ""
      }
    },
    watch: {
      $route(to, from) {
        this.init();
      }
    },
    mounted() {
      this.importAll(require.context('@/assets/docs/', true, /\.md$/));
      this.init();
    },
    methods: {
      isActive(url){
        this.$route.fullPath == url;
      },
      getNavigationIndex(key){
        let index = 0;
        for(var navItem in this.navigation){
          if (navItem == key){
            return index;
          }
          index++
        }
        return -1;
      },
      init() {
        this.resetTopicState();

        let topic = this.$route.params.topic;
        let subject = this.$route.params.subject;
        let navigation = topic;
        if (subject != undefined){
          navigation += '/' + subject;
        }
        this.active = navigation;
        this.current = this.files.find(x => x.navigation.toLowerCase() == navigation.toLowerCase());
         
        let index = Object.values(this.navigation).map(function(x) { return x.text; }).indexOf(topic);
        this.topicState[index] = true;
      },
      resetTopicState(){
        this.topicState.fill(false);
      },
      importAll(r) {
        r.keys().forEach(key => {
          let v = r(key)
          if (v.attributes.navigation != undefined) {
            let navigation = v.attributes.navigation;

            let pieces = navigation.split('/');
            let navi = this.navigation;
            pieces.forEach((value, index, array) => {
              if (navi[value] == undefined){
                navi[value] = {text: value, children:{}}
                if (index == array.length -1){
                  navi[value]["href"] = "/docs/" + navigation
                }else{
                  this.topicState.push(false);
                }
              }
              navi = navi[value].children
            })

            this.files.push({navigation: navigation, obj: v});
          }
        });
      }
    }
  }
</script>

<style scoped>
.doc{
    width: 90%;
    margin: auto;
    margin-top: 42px;
}
 .card{
    margin: 15px;
  }
.card p{
    margin-bottom: 0;
}
.menu-item{
border: 0;
padding: 0.2rem;

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
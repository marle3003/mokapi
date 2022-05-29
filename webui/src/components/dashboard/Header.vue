<template>
  <div class="page-header w-100">
    <b-row>
      <b-col class="mr-auto" cols="auto">
        <b-navbar class="p-0">
          <b-navbar-nav>
            <b-nav-item :to="{ name: 'dashboard', query: {refresh: this.$route.query.refresh} }">Overview</b-nav-item>
            <b-nav-item
              :to="{ name: 'http', query: {refresh: this.$route.query.refresh} }"
              v-if="httpEnabled"
            >HTTP</b-nav-item>
            <b-nav-item
              :to="{ name: 'kafka', query: {refresh: this.$route.query.refresh} }"
              v-if="kafkaEnabled"
            >Kafka</b-nav-item>
            <b-nav-item
              :to="{ name: 'smtp', query: {refresh: this.$route.query.refresh} }"
              v-if="smtpEnabled"
            >SMTP</b-nav-item>
          </b-navbar-nav>
        </b-navbar>
      </b-col>
      <b-col cols="auto" class="pr-0">
        <div v-if="displayBack"
            class="close"
            @click="$router.go(-1)"
          >
            <b-icon
              icon="x"
              class="border rounded p-1"
            ></b-icon>
          </div>
      </b-col>
    </b-row>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import Refresh from '@/mixins/Refresh'

export default {
  mixins: [Api, Refresh],
  data () {
    return {
      httpEnabled: false,
      kafkaEnabled: false,
      smtpEnabled: false
    }
  },
  computed: {
    displayBack: function() {
      return this.$route.meta.showBack
    }
  },
  methods: {
    async getData () {
      this.info().then(i => {
        this.httpEnabled = i.activeServices.includes('http')
        this.kafkaEnabled = i.activeServices.includes('kafka')
        this.smtpEnabled = i.activeServices.includes('smtp')
      })
    }
  }
}
</script>

<style scoped>
.page-header {
  margin-left: -8px;
}
.page-header .nav-link {
  color: var(--var-color-primary);
  position: relative;
  border-radius: 6px;
  margin-right: 5px;
}
.page-header .nav-link:hover {
  color: var(--var-color-primary);
  text-decoration: none;
  background-color: var(--var-bg-color-secondary);
  opacity: 0.8;
}
.page-header .nav-link.router-link-exact-active {
  color: var(--var-color-primary);
  text-decoration: none;
  background-color: var(--var-bg-color-secondary);
  opacity: 0.8;
}
</style>

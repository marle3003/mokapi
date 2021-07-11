<template>
  <div class="service-list">
      <h1>Services</h1>
      <div class="page-header">
        <b-navbar class="p-0">
          <b-navbar-nav>
            <b-nav-item to="/services">All</b-nav-item>
            <b-nav-item :to="{ name: 'httpServices' }" v-if="hasHttp">HTTP</b-nav-item>
            <b-nav-item :to="{ name: 'kafkaServices' }" v-if="hasKafka">Kafka</b-nav-item>
            <b-nav-item :to="{ name: 'ldapServices' }" v-if="hasLdap">LDAP</b-nav-item>
            <b-nav-item :to="{ name: 'smtpServices' }" v-if="hasSmtp">SMTP</b-nav-item>
          </b-navbar-nav>
        </b-navbar>
      </div>
      <div class="page-body">
        <b-link :to="{ name: service.type.toLowerCase()+'Service', params: {name: service.name } }" v-show="$route.name === 'serviceList' || $route.name === service.type.toLowerCase() + 'Services'"
          router-tag="div" v-for="service in services" :key="service.name" class="service"
          :class="[service.name === undefined ? 'disabled' : '']" :event="[service.name === undefined ? '' : 'click']">
          <async-service-info :service="service" v-if="service.type === 'kafka'"></async-service-info>
          <smtp-service-info :service="service" v-if="service.type === 'smtp'"></smtp-service-info>
          <ldap-service-info :service="service" v-if="service.type === 'ldap'"></ldap-service-info>
          <service-info :service="service" v-else></service-info>
        </b-link>
      </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import ServiceInfo from '@/components/ServiceInfo'
import AsyncServiceInfo from '@/components/asyncapi/ServiceInfo'
import SmtpServiceInfo from '@/components/smtp/ServiceInfo'
import LdapServiceInfo from '@/components/ldap/ServiceInfo'

export default {
  components: {
    'service-info': ServiceInfo,
    'async-service-info': AsyncServiceInfo,
    'smtp-service-info': SmtpServiceInfo,
    'ldap-service-info': LdapServiceInfo
  },
  mixins: [Api],
  data () {
    return {
      services: [],
      timer: null,
      loaded: false
    }
  },
  computed: {
    hasHttp () {
      for (let s of this.services) {
        if (s.type.toLowerCase() === 'http') {
          return true
        }
      }
      return false
    },
    hasKafka () {
      for (let s of this.services) {
        if (s.type.toLowerCase() === 'kafka') {
          return true
        }
      }
      return false
    },
    hasLdap () {
      for (let s of this.services) {
        if (s.type.toLowerCase() === 'ldap') {
          return true
        }
      }
      return false
    },
    hasSmtp () {
      for (let s of this.services) {
        if (s.type.toLowerCase() === 'smtp') {
          return true
        }
      }
      return false
    }
  },
  created () {
    this.getData()
    this.timer = setInterval(this.getData, 20000)
  },
  methods: {
    async getData () {
      function compare (s1, s2) {
        if (s1.name === s2.name) {
          return 0
        }
        if (s1.name === undefined) {
          return 1
        } else if (s2.name === undefined) {
          return -1
        }
        const a = s1.name.toLowerCase()
        const b = s2.name.toLowerCase()
        if (a < b) {
          return -1
        }
        if (a > b) {
          return 1
        }
        return 0
      }

      let services = await this.getServices()
      this.services = services.sort(compare)
      this.loaded = true
    }
  },
  beforeDestroy () {
    clearInterval(this.timer)
  }
}
</script>

<style scoped>
  .page-header .nav-link{
    color: var(--var-color-primary);
    position:relative;
    border-radius: 6px;
    margin-right: 5px;
  }
  .page-header .nav-link:hover{
    color: var(--var-color-primary);
    text-decoration: none;
    background-color: var(--var-bg-color-secondary);
    opacity: 0.8;
  }
  .page-header .nav-link.router-link-exact-active{
    color: var(--var-color-primary);
    text-decoration: none;
    background-color: var(--var-bg-color-secondary);
    opacity: 0.8;
  }
  .service-list{
      width: 90%;
    margin: 42px auto auto;
  }
  .service-list .router-link-active{
    text-decoration: none;
  }
  .service > .card{
    margin: 15px 15px 15px 0;
    cursor: pointer;
  }
  .service > .card:hover {
    background-color: var(--var-bg-color-secondary);
    opacity: 0.8;
  }
  .disabled .card{
    cursor: not-allowed;
  }
  .card p{
      margin-bottom: 0;
  }
  .name{
    font-size: 1.25rem;
    font-weight: 500;
    padding-bottom: 0.5rem;
  }
</style>

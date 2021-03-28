<template>
  <div>
    <b-navbar toggleable="lg" type="light"  class="topHeader">
      <b-navbar-brand href="/" >{{title}}</b-navbar-brand>

      <b-collapse id="nav-collapse" is-nav>
        <b-navbar-nav>
        <b-nav-item to="/dashboard">Dashboard</b-nav-item>
        <b-nav-item :to="{ name: 'serviceList' }">Services</b-nav-item>
        <b-nav-item :to="{ name: 'docsStart' }">Docs</b-nav-item>
        </b-navbar-nav>
      </b-collapse>

      <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>
    </b-navbar>
  </div>
</template>

<script>
export default {
  name: 'topHeader',
  components: {},
  data () {
    return {
      title: process.env.VUE_APP_TITLE
    }
  }
}
</script>

<style scoped>
.topHeader.navbar {
    padding-bottom: 0;
    padding-top: 0;
    margin-top: 8px;
}
.topHeader .navbar-brand{
  color: var(--var-color-primary);
  font-size: 1.0rem;
  font-weight: 500;
}
.topHeader .navbar-nav::before{
  content: '';
  display: inline-block;
  border-left: 2px solid var(--var-color-primary);
  margin-left: 0.5rem;
  margin-right: 1rem;
}
#nav-collapse .nav-link{
  color: var(--var-color-primary);
  font-weight: 500;
  font-size: 1.0rem;
}

#nav-collapse .nav-link:hover{
  border-bottom: 4px solid var(--var-color-primary);
  margin-bottom: -4px;
   text-decoration: none;
}

#nav-collapse .nav-link.router-link-active{
  border-bottom: 4px solid var(--var-color-primary);
  margin-bottom: -4px;
}
#nav-collapse .nav-link.router-link-exact-active{
  text-decoration: none;
}
</style>

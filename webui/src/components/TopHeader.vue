<template>
<div>
  <b-navbar toggleable="lg" type="light" class="header">
  <b-navbar-brand href="/" >{{title}}</b-navbar-brand>

  <b-collapse id="nav-collapse" is-nav>
    <b-navbar-nav>
      <b-nav-item to="/dashboard">Dashboard</b-nav-item>
      <b-nav-item :to="{ name: 'serviceList' }">Services</b-nav-item>
    </b-navbar-nav>
  </b-collapse>

  <b-navbar-toggle target="nav-collapse">

  </b-navbar-toggle>

  </b-navbar>
  </div>
</template>

<script>
export default {
    name: "topHeader",
    components: {},
    data() {
      return {
        title: process.env.VUE_APP_TITLE
      }
    }
}
</script>

<style scoped>
.header{
  background-color: white;
  box-shadow: 0 3px 8px 0;
}
#nav-collapse .nav-link{
  color: #212529;
  font-weight: 600;
}

#nav-collapse .nav-link:hover{
  color: #007bff;
  border-bottom: 4px solid #007bff;
  margin-bottom: -4px
}

#nav-collapse .nav-link.router-link-active{
  color: #007bff;
  border-bottom: 4px solid #007bff;
  margin-bottom: -4px
}
</style>
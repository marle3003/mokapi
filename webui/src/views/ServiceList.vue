<template>
  <div class="service-list">
      <div class="page-header">
          <h2>Services</h2>
      </div>
      <div class="page-body">
        <b-link :to="{ name: 'service', params: {name: service.name} }" router-tag="div" v-for="service in services" :key="service.name">
          <b-card>
            <b-row>
              <b-col class="col-auto">
                <img src="@/assets/service.png" />
              </b-col>
              <b-col>
                <p class="name">{{ service.name }}</p>
                <p class="description">{{ service.description }}</p>
              </b-col>
              <b-col>
              <p>{{ service.version }}</p>
              </b-col>
            </b-row>
          </b-card>
        </b-link>
      </div>
  </div>
</template>

<script>
export default {
    components: {},
    data(){
      return {
        services: []
      }
    },
    mounted(){
      this.$http.get(
      'http://localhost:8081/api/services')
        .then(response => (this.services = response.data))
    }
}
</script>

<style scoped>
.service-list{
    width: 90%;
    margin: auto;
    margin-top: 42px;
}
.page-header h2{
    font-weight: 700;
}
  .card{
    margin: 15px;
    cursor: pointer;
  }
.card p{
    margin-bottom: 0;
}
  .description{
    color: #a0a1a7;
  }
  .name{
    font-size: 1.5rem;
    font-weight: 600;
  }
</style>
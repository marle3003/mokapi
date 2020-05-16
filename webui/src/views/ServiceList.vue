<template>
  <div class="service-list">
      <div class="page-header">
          <h2>Services</h2>
      </div>
      <div class="page-body">
          <b-link :to="{ name: 'service', params: {name: 'demo'} }" router-tag="div">
           <b-card>
    <b-row>
        <b-col class="col-auto">
            <img src="@/assets/service.png" />
        </b-col>
      <b-col>
        <p class="name">Demo</p>
        <p class="description">A demo service</p>
      </b-col>
      <b-col>
        <p>1.0.0</p>
      </b-col>
    </b-row>
  </b-card>
  </b-link>
      </div>
  </div>
</template>

<script>
export default {
    components: {}
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
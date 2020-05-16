<template>
  <b-card class="w-100">
    <b-row>
      <b-col class="col-auto">
            <img src="@/assets/service.png" />
        </b-col>
      <b-col>
        <p class="label">Service Name</p>
        <p class="title">Demo</p>
        <p class="label">Description</p>
        <p>A demo service</p>
      </b-col>
      <b-col>
        <p class="label">Version</p>
        <p>1.0.0</p>
        <p class="label">Base URL</p>
        <ul class="list-unstyled">
            <li>http://demo</li>
            <li>http://staging-api.example.com:7001</li>
        </ul>
      </b-col>
    </b-row>
  </b-card>
</template>

<script>
export default {
    name: "serviceInfo",
    components: {}
}
</script>

<style scoped>

  .label{
    color: #a0a1a7;
    margin-bottom: 0
  }
  .title{
    font-size: 1.5rem;
    font-weight: 600;
  }
</style>

<template>
      <b-card class="w-100">
      <b-row v-if="service !== null">
        <b-col class="col-image">
              <img src="@/assets/service.png" />
          </b-col>
        <b-col>
          <p class="label">Service Name</p>
          <p class="title">{{ service.name }}</p>
          <p class="label">Version</p>
          <p>{{ service.version }}</p>
        </b-col>
        <b-col>
          <p class="label">Mokapifile</p>
          <p>{{ service.mokapifile }}</p>
          <p class="label">Base URL</p>
          <ul class="list-unstyled">
              <li v-for="url in service.baseUrls" :key="url.url" v-b-tooltip.hover.left :title="url.description">{{ url.url }}</li>
          </ul>
        </b-col>
      </b-row>
      <b-row v-if="service !== null">
        <b-col class="col-image">
        </b-col>
        <b-col>
          <div v-if="service.description != null">
            <p class="label">Description</p>
            <vue-simple-markdown :source="service.description" />
          </div>
        </b-col>
      </b-row>
    </b-card>
</template>

<script>
export default {
  name: 'serviceInfo',
  props: ['service'],
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
  .col-image{
    width: 100px;
    flex: 0 0 auto
  }
</style>

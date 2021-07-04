<template>
  <div class="dashboard" v-if="mail !== null">
    <div class="page-header">
      <h2>Smtp Mail</h2>
    </div>
    <div class="page-body">
      <b-card class="w-100">
        <b-row>
          <b-col>
            <p class="label">From</p>
            <p>{{ mail.from}}</p>
            <p class="label">To</p>
            <p>{{ mail.to}}</p>
            <p class="label">Data</p>
            <p v-html="this.$options.filters.body(mail.data)" class="mailBody"></p>
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'

export default {
  name: 'SmtpMail',
  mixins: [Api],
  data () {
    return {
      mail: null
    }
  },
  created () {
    this.getData()
  },
  methods: {
    async getData () {
      let id = this.$route.params.id
      this.mail = await this.getSmtpMail(id)
    }
  },
  filters: {
    body: function (body) {
      return '<p class="mailBody">' + body.replaceAll('\r\n', '&nbsp;</p><p class="mailBody">') + '</p>'
    }
  }
}
</script>

<style scoped>
  .dashboard{
    width: 90%;
    margin: 42px auto auto;
  }
  .page-header h2{
    font-weight: 400;
  }
</style>
<style>
.mailBody{
  margin-bottom: 5px;
}
</style>

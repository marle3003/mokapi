<template>
  <div class="dashboard" v-if="mail !== null">
    <div class="page-header">
      <h2>Smtp Mail</h2>
    </div>
    <div class="page-body">
      <b-card class="w-100 mb-4">
        <b-row>
          <b-col>
            <div v-if="mail.sender !== null">
              <div class="label">Sender</div>
              <div class="value">
                <span v-if="mail.sender.Name !== ''">{{ mail.sender.Name }} &lt;</span><span>{{ mail.sender.Address }}</span><span v-if="mail.sender.Name !== ''">&gt;</span>
              </div>
            </div>
            <div v-if="mail.from !== null && mail.from.length > 0">
              <div class="label">From</div>
              <div class="value">
                <div v-for="from in mail.from" :key="from.Address">
                  <span v-if="from.Name !== ''">{{ from.Name }} &lt;</span><span>{{ from.Address }}</span><span v-if="from.Name !== ''">&gt;</span>
                </div>
              </div>
            </div>
            <div v-if="mail.replyTo !== null && mail.replyTo.length > 0">
              <div class="label">Reply to</div>
              <div class="value">
                <div v-for="replyTo in mail.replyTo" :key="replyTo.Address">
                  <span v-if="replyTo.Name !== ''">{{ replyTo.Name }} &lt;</span><span>{{ replyTo.Address }}</span><span v-if="replyTo.Name !== ''">&gt;</span>
                </div>
              </div>
            </div>
            <div v-if="mail.contentType !== ''">
              <div class="label">Content type</div>
              <div class="value">{{ mail.contentType}}</div>
            </div>
            <div v-if="mail.encoding !== ''">
              <div class="label">Encoding</div>
              <div class="value">{{ mail.encoding}}</div>
            </div>
          </b-col>
          <b-col>
            <div v-if="mail.to !== null && mail.to.length > 0">
              <div class="label">To</div>
              <div class="value">
                <div v-for="to in mail.to" :key="to.Address">
                  <span v-if="to.Name !== ''">{{ to.Name }} &lt;</span><span>{{ to.Address }}</span><span v-if="to.Name !== ''">&gt;</span>
                </div>
              </div>
            </div>
            <div v-if="mail.cc !== null && mail.cc.length > 0">
              <div class="label">CC</div>
              <div class="value">
                <div v-for="cc in mail.cc" :key="cc.Address">
                  <span v-if="cc.Name !== ''">{{ cc.Name }} &lt;</span><span>{{ cc.Address }}</span><span v-if="cc.Name !== ''">&gt;</span>
                </div>
              </div>
            </div>
            <div v-if="mail.bcc !== null && mail.bcc.length > 0">
              <div class="label">BCC</div>
              <div class="value">
                <div v-for="bcc in mail.bcc" :key="bcc.Address">
                  <span v-if="bcc.Name !== ''">{{ bcc.Name }} &lt;</span><span>{{ bcc.Address }}</span><span v-if="bcc.Name !== ''">&gt;</span>
                </div>
              </div>
            </div>
            <div class="label">Time</div>
            <div>{{ mail.time | moment }}</div>
          </b-col>
        </b-row>
      </b-card>
      <b-card class="w-100" >
        <b-row>
          <b-col>
            <p class="label" v-if="mail.subject !== ''">Subject</p>
            <p>{{ mail.subject }}</p>
            <p class="label">Body</p>
            <p v-html="this.$options.filters.body(mail.htmlBody)" v-if="mail.htmlBody !== ''"></p>
            <p v-if="mail.textBody !== ''">{{ mail.textBody}}</p>
          </b-col>
        </b-row>
      </b-card>
    </div>
  </div>
</template>

<script>
import Api from '@/mixins/Api'
import moment from 'moment'

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
    },
    moment: function (date) {
      return moment(date).local().format('YYYY-MM-DD HH:mm:ss')
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
  .value{
    padding-bottom: 1rem;
  }
</style>
<style>
.mailBody{
  margin-bottom: 5px;
}
</style>

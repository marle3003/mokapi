'use strict'
const merge = require('webpack-merge')
const prodEnv = require('./prod.env')

module.exports = merge(prodEnv, {
  NODE_ENV: '"development"',
  VUE_APP_TITLE: '"MokApi (DEV)"',
  VUE_APP_ApiBaseUrl: '"http://localhost:8082"'
})

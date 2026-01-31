import { every } from 'mokapi';
import { produce } from 'mokapi/kafka'

export default () => {
    every('10ms', () => {
        produce({ topic: 'user_signedup' })
    }, { times: 10 })
}
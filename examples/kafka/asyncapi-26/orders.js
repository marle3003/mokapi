import { every } from 'mokapi';
import { produce } from 'mokapi/kafka'

export default () => {
    every('10ms', () => {
        produce({ topic: 'orders' })
    }, { times: 10 })
}
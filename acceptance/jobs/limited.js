import { every } from 'mokapi'

export default function() {
    every('200ms', () => {
        console.log('limited run')
    }, { times: 2 })
}

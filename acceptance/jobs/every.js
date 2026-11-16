import { every } from 'mokapi'

export default function() {
    every('5m', () => {
        console.log('tick')
    })
}

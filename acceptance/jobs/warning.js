import { every } from 'mokapi'

export default function() {
    every('1h', () => {
        console.warn('careful')
    })
}

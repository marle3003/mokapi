import { every } from 'mokapi'

export default function() {
    every('1h', () => {
        throw new Error('job failed')
    })
}

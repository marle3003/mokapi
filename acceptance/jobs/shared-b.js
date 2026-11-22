import { every } from 'mokapi'

export default function() {
    every('1h', () => {
        console.log('shared run')
    }, { tags: { name: 'shared' } })
}

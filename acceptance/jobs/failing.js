import { every } from 'mokapi'

export default function() {
    every('1h', () => {
        console.error('something went wrong')
    })
}

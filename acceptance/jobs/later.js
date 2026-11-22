import { every } from 'mokapi'

export default function() {
    every('1h', () => {
        console.log('later run')
    }, { skipImmediateFirstRun: true, tags: { name: 'later' } })
}

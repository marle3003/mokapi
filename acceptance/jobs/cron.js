import { cron } from 'mokapi'

export default function() {
    cron('* * * * *', () => {
        console.log('cron run')
    })

    cron('* * * * *', () => {
        console.log('cron run with times')
    }, { times: 1, skipImmediateFirstRun: false })
}

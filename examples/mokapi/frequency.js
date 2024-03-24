import { findByName} from 'mokapi/faker'

export default function() {
    const faker = findByName('Faker')
    const frequency = ['daily', 'weekly', 'monthly', 'yearly']
    faker.insert(0, {
        name: 'Frequency',
        test: (r) => {
            return r.lastName() === 'frequency'
        },
        fake: (r) => {
            return frequency[Math.floor(Math.random()*frequency.length)]
        }
    })
}
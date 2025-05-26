import { findByName, ROOT_NAME } from 'mokapi/faker'

export default function() {
    const root = findByName(ROOT_NAME)
    const frequency = ['daily', 'weekly', 'monthly', 'yearly']
    root.children.unshift({
        name: 'Frequency',
        attributes: ['frequency'],
        fake: (r) => {
            return frequency[Math.floor(Math.random()*frequency.length)]
        }
    })
}
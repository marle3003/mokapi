import { findByName, ROOT_NAME } from 'mokapi/faker'

export default function () {
    // the path id stays in context.values inside friends[]
    findByName(ROOT_NAME).children.push({
        name: 'ownerId',
        attributes: ['ownerId'],
        fake: (r) => r.context.values.id
    })
}

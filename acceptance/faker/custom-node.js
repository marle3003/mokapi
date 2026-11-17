import { findByName, ROOT_NAME } from 'mokapi/faker'

export default function () {
    const root = findByName(ROOT_NAME)

    root.children.push({
        name: 'frequency',
        attributes: ['frequency'],
        fake: () => 'weekly'
    })

    // same node name in two branches
    root.children.push({
        name: 'billing',
        children: [{ name: 'code', attributes: ['code'], fake: () => 'BILL-1' }]
    })
    root.children.push({
        name: 'shipping',
        children: [{ name: 'code', attributes: ['code'], fake: () => 'SHIP-1' }]
    })
}

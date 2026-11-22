import { fake, findByName, ROOT_NAME } from 'mokapi/faker'

// A custom node for the shop story: SKUs of the pet shop, so that a property named sku
// doesn't get a random word.
export default function () {
    const root = findByName(ROOT_NAME)

    root.children.push({
        name: 'sku',
        attributes: ['sku'],
        fake: () => fake({ type: 'string', pattern: '^PET-[0-9]{4}$' })
    })
}

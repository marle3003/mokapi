import { produce } from 'mokapi/kafka'

export default function() {
    produce({ topic: 'orders' })
    produce({ topic: 'orders', value: {orderId: 1, customer: 'Alice', items: [{itemId: 200, quantity: 3}]} })
}
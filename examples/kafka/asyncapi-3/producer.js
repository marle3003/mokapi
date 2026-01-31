import { produce } from 'mokapi/kafka';

export default function() {
    produce({ topic: 'users.signedup' });
}
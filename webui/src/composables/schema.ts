export function useSchema() {
    function printType(schema: Schema): string {
        if (!schema) {
            return ''
        }

        let types = ''
        if (Array.isArray(schema.type)) {
            for (const type of schema.type) {
                if (types.length > 0) {
                    types += ', '
                }
                if (type === 'array'){
                    types += `${schema.items.type}[]`
                } else {
                    types += type
                }
            }
            return types
        } else {
            if (schema.type === 'array'){
                return `${schema.items.type}[]`
            }
            return schema.type
        }
    }

    return { printType }
}
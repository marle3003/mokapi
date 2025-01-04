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
                if (type === 'array' && schema.items){
                    types += `${schema.items.type}[]`
                } else {
                    types += type
                }
            }
            return types
        } else {
            if (schema.type === 'array' && schema.items){
                return `${schema.items.type}[]`
            }
            return schema.type
        }
    }

    return { printType }
}
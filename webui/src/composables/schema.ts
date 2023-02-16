export function useSchema() {
    function printType(schema: Schema): string {
        if (schema.type == 'array'){
            return `${schema.items.type}[]`
        }
        return schema.type
    }

    return {printType}
}
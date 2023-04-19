declare module 'mokapi/yaml' {
    function parse(s: string): any
    function stringify(value: any): string
}
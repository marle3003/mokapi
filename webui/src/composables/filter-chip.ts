// A chip of a log toolbar stands for one active filter field.
export interface FilterChip<Field extends string = string> {
    field: Field
    label: string
    // the full text, the label is cut when it is long
    title: string
}

const maxLabelLength = 40

export function shorten(text: string): string {
    return text.length <= maxLabelLength ? text : text.slice(0, maxLabelLength - 1) + '…'
}

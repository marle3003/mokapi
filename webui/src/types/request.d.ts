declare interface Response {
    data: any
    isLoading: boolean
    error: string
    close: () => void
}
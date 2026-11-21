// Stands in for the 'mokapi' module, which only exists inside the Mokapi runtime. The mock
// scripts in e2e/mocks/ import `app` from it, so unit tests can load their data. Mapped in
// `resolve.alias` of vite.config.ts when VITEST is set and in `paths` of tsconfig.json.
// The runtime types these values by the OpenAPI operation; handlers compare them with numbers.
/* eslint-disable @typescript-eslint/no-explicit-any */
interface HttpRequest {
  path: Record<string, any>
  query: Record<string, any>
  header: Record<string, any>
}
/* eslint-enable @typescript-eslint/no-explicit-any */

interface HttpResponse {
  data: unknown
  headers: Record<string, string>
  rebuild(statusCode: number): void
}

type HttpHandler = (request: HttpRequest, response: HttpResponse) => void

interface HttpApi {
  operation(operationId: string, handler: HttpHandler): HttpApi
}

const noop: HttpApi = { operation: () => noop }

export const app = { http: (): HttpApi => noop }

// Stands in for 'mokapi/encoding', which only exists inside the Mokapi runtime. e2e/mocks/kafka.ts
// calls `base64.encode` while the module loads (event `avro`, `message.binary`), so it has to
// return a string. Mapped in `resolve.alias` of vite.config.ts when VITEST is set and in `paths`
// of tsconfig.json.
export const base64 = {
  encode: (value: string): string => btoa(value),
}

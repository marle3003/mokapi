export function isValidImage(target: EventTarget | null): [true, HTMLImageElement] | [false, undefined] {
  if (hasTouchSupport() || !target || !(target instanceof HTMLImageElement) || target.classList.contains('no-dialog')) {
    return [false, undefined]
  }
  return [true, target as HTMLImageElement]
}
export function hasTouchSupport() {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
}
import type { Modal } from 'bootstrap'

// Bootstrap's JavaScript touches `document` when it is imported, so it can't be imported at module
// level in code that is also rendered on the server. Dialogs are only opened in the browser.
export async function loadModal(): Promise<typeof Modal> {
    return (await import('bootstrap')).Modal
}

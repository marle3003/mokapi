import { test as base } from '@playwright/test'
import { HomeModel } from './home'

type WebsiteFixture = {
    home: HomeModel
}

export { expect } from '@playwright/test'

export const test = base.extend<WebsiteFixture>({
    context: async ({ context }, use) => {
        await context.addInitScript(() => {
            const addNoMotionClass = () => {
                document.body.classList.add('no-motion');
            };
            document.addEventListener('DOMContentLoaded', addNoMotionClass);
        });
        await use(context);
    },

    home: async ({page}, use) => {
        const home = new HomeModel(page)
        await use(home)
    },
})
import { test as base } from '@playwright/test'
import { HomeModel } from './home'

type WebsiteFixture = {
    home: HomeModel
}

export { expect } from '@playwright/test'

export const test = base.extend<WebsiteFixture>({
    home: async ({page}, use) => {
        const home = new HomeModel(page)
        await use(home)
    },
})
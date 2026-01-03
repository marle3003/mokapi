<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useMetrics } from '@/composables/metrics';
import { useRoute } from 'vue-router';
import type { AppInfo } from '@/types/dashboard';
import { computed, onBeforeUnmount, onMounted, nextTick, useTemplateRef, type PropType } from 'vue';
import { useRouter } from '@/router';

const props = defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
});

const { dashboard, getMode } = useDashboard();
const { sum } = useMetrics();
const tabsContainer = useTemplateRef<HTMLElement>('tabsContainer');
const dropdownMenu = useTemplateRef<HTMLElement>('dropdownMenu');
const moreTabs = useTemplateRef<HTMLElement>('moreTabs');
const tabItems = computed(() => [
  { text: 'Overview', isVisible: true, to: { name: getRouteName('dashboard').value }, cssClass: 'overview' },
  { text: 'HTTP', isVisible: isServiceAvailable('http'), to: { name: getRouteName('http').value } },
  { text: 'Kafka', isVisible: isServiceAvailable('kafka'), to: { name: getRouteName('kafka').value }  },
  { text: 'Mail', isVisible: isServiceAvailable('mail'), to: { name: getRouteName('mail').value }  },
  { text: 'LDAP', isVisible: isServiceAvailable('ldap'), to: { name: getRouteName('ldap').value }  },
  { text: 'Jobs', isVisible: hasJobs.value, to: { name: getRouteName('jobs').value }  },
  { text: 'Configs', isVisible: true, to: { name: getRouteName('configs').value }  },
  { text: 'Faker', isVisible: getMode() === 'live', to: { name: getRouteName('tree').value }  },
  { text: 'Search', isVisible: props.appInfo.search.enabled, to: { name: getRouteName('search').value }  },
]);
const route = useRoute();
const router = useRouter();
const tabActive = computed(() => tabItems.value.find(x => x.text !== 'Overview' && route.matched.some(r => r.name === x.to.name)) ?? tabItems.value[0])

const response = dashboard.value.getMetrics('app')
const hasJobs = computed(() => {
  return sum(response.data, 'app_job_run_total') > 0
})

const handleResize = () => updateTabsResponsiveness();

onMounted(async () => {
  // Wait for Vue to finish rendering the DOM
  await nextTick();
  window.addEventListener("resize", handleResize);
  updateTabsResponsiveness();
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
});

function isServiceAvailable(service: string): boolean{
    if (!props.appInfo.activeServices){
        return false
    }
    return props.appInfo.activeServices.includes(service)
}

function updateTabsResponsiveness() {
  const container = tabsContainer.value;
  const dropdown = dropdownMenu.value;
  const more = moreTabs.value

  if (!container || !dropdown || !more) {
    return
  }

  // Reset everything
  const allTabs = [...container.querySelectorAll("li.nav-item:not(#moreTabs)")];
  for (const tab of allTabs) {
    tab.classList.remove('d-none');
  }
  for (const tab of [...dropdown.querySelectorAll("li")]) {
    tab.classList.add('d-none');
  }
  // Make sure it's measurable
  more.classList.remove("d-none");  
  const moreWidth = more.offsetWidth;
  more.classList.add("d-none");

  const documentWidth = document.body.clientWidth * 0.9;
  let usedWidth = moreWidth;
  let hidden = 0;

  for (const [index, item] of allTabs.entries()) {
    const tab = item as HTMLElement

    // do not show the more dropdown only for one item
    if (index === (allTabs.length-1) && hidden === 0) {
      break;
    }

    if (usedWidth + tab.offsetWidth > documentWidth) {
      tab.classList.add("d-none");
      const dropItem = Array.from(dropdown.querySelectorAll("li"))
        .find(el => el.textContent.trim() === tab.innerText);
      if (dropItem) {
        dropItem.classList.remove("d-none");
      }
      more.classList.remove("d-none");
      hidden++;
    } else {
      usedWidth += tab.offsetWidth;
    }
  }

  const activeTab = allTabs.find(tab => 
    tab.querySelector('.nav-link.router-link-exact-active')
  );
  const moreLink = more.querySelector('.nav-link')!;
  // If the active tab is hidden => it's in the dropdown
  if (moreLink && activeTab && activeTab.classList.contains('d-none')) {
    // Get the label text of the active tab
    const activeLabel = activeTab.querySelector('.nav-link')!.textContent.trim();
    // Replace "More" text with active tab label
    moreLink.textContent = activeLabel;
    moreLink.classList.add('router-link-exact-active')

  } else {
    // Restore default
    moreLink.textContent = 'More';
    moreLink.classList.remove('router-link-exact-active')
  }
};
</script>

<template>
  <nav class="navbar navbar-expand pb-1" aria-label="Services">
    <div>
      <ul class="navbar-nav me-auto mb-0" ref="tabsContainer">
        <template v-for="tabItem in tabItems" :key="tabItem.text">
          <li class="nav-item" :class="tabItem.cssClass ? tabItem.cssClass : ''" v-if="tabItem.isVisible">
              <router-link class="nav-link" :to="tabItem.to">{{ tabItem.text }}</router-link>
          </li>
        </template>
        <li id="moreTabs" class="nav-item dropdown d-none" ref="moreTabs">
            <a class="nav-link dropdown-toggle" data-bs-toggle="dropdown" href="#" role="button" aria-expanded="false">{{ tabActive?.text }}</a>
            <ul class="dropdown-menu" ref="dropdownMenu">
              <template v-for="tabItem in tabItems" :key="tabItem.text">
                <li>
                  <a :href="router.resolve(tabItem.to).href" class="dropdown-item">
                    {{ tabItem.text }}
                  </a>
                </li>
              </template>
            </ul>
        </li>
      </ul>
    </div>
  </nav>
</template>
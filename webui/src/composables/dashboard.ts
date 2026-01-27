import type { Dashboard } from "@/types/dashboard";
import { computed, ref } from "vue";
import * as live from "./dashboard.backend";
import * as demo from "./dashboard.demo";

const current = ref<Dashboard>(live.dashboard)

type Mode = 'live' | 'demo'

const mode = ref<Mode>('live')

const dashboard = computed<Dashboard>(() => {
    console.log(mode.value)
    if (mode.value === 'live') {
        return live.dashboard
    }
    return demo.useDemoDashboard()
})

export function useDashboard() {
    
    function setMode(m: Mode) {
        mode.value = m
    }

    function getMode(): Mode {
        return mode.value
    }

    return { dashboard, setMode, getMode }
}

export const getRouteName = (name: string) => {
    return computed<string>(x => {
        return mode.value === 'live' ? name : name + '-demo'
    })
    
}
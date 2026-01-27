import { ref, watch, type Ref } from 'vue';

const store = new Map<string, any>();

export function useLocalStorage<T>(key: string, defaultValue: T): Ref<T> {
    if (store.has(key)) {
        return store.get(key);
    }

    const stored = localStorage.getItem(key);
    const state = ref<T>(stored ? JSON.parse(stored) : defaultValue) as Ref<T>;

    watch(
        state,
        value => {
            localStorage.setItem(key, JSON.stringify(value));
        },
        { deep: true }
    );

    // Cross-tab sync
    window.addEventListener('storage', e => {
        if (e.key === key && e.newValue) {
            state.value = JSON.parse(e.newValue) as T;
        }
    });

    store.set(key, state);
    return state;
}
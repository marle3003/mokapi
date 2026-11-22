import { effectScope, ref, watch, type Ref } from 'vue';

const store = new Map<string, Ref<unknown>>();

// A value that was not written by this function (e.g. edited by hand) is ignored.
function parse<T>(json: string | null, fallback: T): T {
    if (!json) {
        return fallback;
    }
    try {
        return JSON.parse(json) as T;
    } catch {
        return fallback;
    }
}

// With `sync: false` the value is still saved on every change, but a change made in another
// browser window is not taken over (e.g. the open tabs, which belong to one window).
// The value is shared per key: `defaultValue` and `sync` only apply to the first caller of a key.
export function useLocalStorage<T>(key: string, defaultValue: T, { sync = true } = {}): Ref<T> {
    if (store.has(key)) {
        return store.get(key) as Ref<T>;
    }

    const state = ref<T>(parse(localStorage.getItem(key), defaultValue)) as Ref<T>;

    // The state is shared through the store, so saving must not stop with the component that created it.
    effectScope(true).run(() => watch(
        state,
        value => {
            localStorage.setItem(key, JSON.stringify(value));
        },
        { deep: true }
    ));

    if (sync) {
        window.addEventListener('storage', e => {
            if (e.key === key && e.newValue) {
                state.value = parse(e.newValue, state.value);
            }
        });
    }

    store.set(key, state);
    return state;
}
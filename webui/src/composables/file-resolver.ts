import type { RouteLocationNormalizedLoaded } from "vue-router"

const MAX_LEVEL = 4

export function useFileResolver() {
    
    function resolve(config: DocConfig, route: RouteLocationNormalizedLoaded) {
        let level1 = <string>route.params.level1
        if (!level1) {
            return { file: null, levels: [] }
        }
        let file: DocEntry | string
        ({ name: level1, file} = select(config, level1))
        if (!file) {
            return { file, levels: [] }
        }

        const levels = [ level1 ]
        let isIndex = false
        for (let index = 2; index <= MAX_LEVEL; index++) {
            let level = <string>route.params[`level${index}`];
            if (typeof file === 'string') {
                break
            }
            ({ level: level, file, isIndex } = getLevel(level, file))
            if (!level) {
                break
            }
            levels.push(level)
        }

        return { file, levels, isIndex }
    }

    function getLevel(level: string, file: DocEntry | string) {
        let isIndex = false
        if (typeof file !== 'string' && file.items) {
            if (!level) {
                if (file.index) {
                    isIndex = true
                    file = file.index
                } else {
                    // get first element as 'index' file
                    level = Object.keys(file.items)[0];
                    ({ file } = find(level, <DocEntry>file))
                }
            } else{
                ({ name: level, file } = find(level,  <DocEntry>file))
            }
        }
        return { level, file, isIndex }
    }

    function find(name: string, config: DocEntry) {
        name = getField(config.items, name)
        return { name: name, file: config.items![name] }
    }

    function select(obj: DocConfig, name: string) {
        name = getField(obj, name)
        return { name: name, file: obj[name] }
    }

    function getField(obj: any, name: string) {
        const searchFor = name.toLowerCase().replaceAll(/[-]/g, ' ')
        return Object.keys(obj).find(
            key => key.toLowerCase().replaceAll(/[\/]/g, ' ') === searchFor)!
    }

    function isKnown(config: DocConfig, level: string): boolean {
        if (getField(config, level)) {
            return true
        }
        return false
    }

    return { resolve, isKnown }
}
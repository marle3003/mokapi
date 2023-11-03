import type { RouteLocationNormalizedLoaded } from "vue-router"

export function useFileResolver() {
    
    function resolve(config: DocConfig, route: RouteLocationNormalizedLoaded) {
        let level1 = <string>route.params.level1
        let file: DocConfig | DocEntry | string
        ({ name: level1, file } = find(level1, config))

        let level2 = <string>route.params.level2
        if (!level2 && typeof file !== 'string' && !isEntry(file)) {
            // get first element as 'index' file
            level2 = Object.keys(file)[0];
            ({ file } = find(level2, <DocConfig>file))
        } else {
            ({ name: level2, file } = find(level2, <DocConfig>file))
        }

        let level3 = <string>route.params.level3
        if (level3 || typeof file !== 'string' && !isEntry(file)) {
            if (!level3) {
                // get first element as 'index' file
                level3 = Object.keys(file)[0];
                ({ file } = find(level3, <DocConfig>file))
            } else{
                ({ name: level3, file } = find(level3, <DocConfig>file))
            }
        }

        return { level1, level2, level3, file }
    }

    function find(name: string, config: DocConfig) {
        const searchFor = name.toLowerCase().replaceAll(/[-]/g, ' ')
        name = Object.keys(config).find(
            key => key.toLowerCase().replaceAll(/[\/]/g, ' ') === searchFor)!
        return { name: name, file: config[name] }
    }

    function isEntry(obj: any) {
        return 'file' in obj || 'component' in obj
    }

    return { resolve }
}
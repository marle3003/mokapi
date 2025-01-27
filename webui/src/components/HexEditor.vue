<script setup lang="ts">
import { computed, onBeforeUpdate, ref } from 'vue';


const props = defineProps<{
    data: string
}>()
const emit = defineEmits(['update'])

const text = ref<HTMLElement>()
const hex = ref<HTMLElement>()

const result = computed(() => {
    var result = '';
    for (var i=0; i<props.data.length; i++) {
        const hex = props.data.charCodeAt(i).toString(16).padStart(2, '0')
        result += hex.toUpperCase();
    }
    return result
})


function rows() {
    if (result.value.length === 0) {
        return 1
    }
    return result.value.match(/.{1,32}/g)!.length
}
function data() {
    if (result.value.length === 0) {
        return []
    }
    return result.value.match(/[0-9a-fA-F]{2}/g)!
}
function string() {
    let result = ''
    for (var i=0; i<props.data.length; i++) {
        const ch = props.data.charCodeAt(i)
        result += toText(ch)
    }
    return result
}
function mouseover(e: MouseEvent) {
    if (!e.target) {
        return
    }
    const target = <HTMLElement>e.target
    const offset = target.getAttribute('data-offset')
    if (!offset) {
        return
    }


    target.classList.add('hover');

    let other: HTMLElement
    if (target.parentNode === hex.value) {
        other = <HTMLElement>text.value?.querySelector(`[data-offset="${offset}"]`)
    } else {
        other = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
    }
    if (other) {
        other.classList.add('hover')
    }
}
function mouseleave(e: MouseEvent) {
    const target = <HTMLElement>e.target
    const offset = target.getAttribute('data-offset')
    if (!offset) {
        return
    }

    target.classList.remove('hover')

    let other: HTMLElement
    if (target.parentNode === hex.value) {
        other = <HTMLElement>text.value?.querySelector(`[data-offset="${offset}"]`)
    } else {
        other = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
    }
    if (other) {
        other.classList.remove('hover')
    }
}

let selection: {
    start?: HTMLElement,
    end?: HTMLElement,
} = {}
function mousedown(e: MouseEvent) {
    let target = <HTMLElement>e.target
    if (target.parentNode === text.value) {
        const offset = target.getAttribute('data-offset')
        target = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
    }

    selection = {
        start: target,
        end: undefined
    }
}

function mouseup(e: MouseEvent) {
    let target = <HTMLElement>e.target
    if (target.parentNode === text.value) {
        const offset = target.getAttribute('data-offset')
        target = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
    }

    if (selection.start && selection.start != target) {
        selection.end = target
        return
    }
    selection.start = undefined

    clearSelected()

    target.classList.add('selected')
    const offset = target.getAttribute('data-offset')
    let other: HTMLElement
    if (target.parentNode === hex.value) {
        other = <HTMLElement>text.value?.querySelector(`[data-offset="${offset}"]`)
    } else {
        other = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
    }
    if (other) {
        other.classList.add('selected')
    }
}
let input = ''
function keydown(e: KeyboardEvent) {
    const ch = e.key.toUpperCase()

    if ((e.ctrlKey || e.metaKey) && ch === 'C') {
        navigator.clipboard.writeText(getSelection());
        return
    }
    
    let selected = <HTMLElement>document.querySelector('.hexedit .selected')
    // only 0-9 and A-F
    if (ch.match(/^[0-9a-fA-F]+$/)) {
        if (selected.classList.contains('last')) {
            selected = createElement(selected)
        }
        if (input.length === 0) {
            input = ch
            selected.innerText = input + ' '
        } else {
            const value = input + ch
            selected.innerText = value
            input = ''

            clearSelected()
            select(selected.nextElementSibling)

            const offset = parseInt(selected.getAttribute('data-offset')!)
            const spanText = <HTMLElement>text.value?.querySelector(`[data-offset="${offset}"]`)
            spanText.innerText = toText(Number('0x' + value))

            const data = Array.from(hex.value!.querySelectorAll('span:not(.last)')).map(x => Number('0x'+x.innerHTML)) 
            emit('update', String.fromCharCode(...data))
        }
    } else {
        switch (e.code) {
            case 'Backspace':
                input = ''
                selected.innerText = '00'
                break
            case 'ArrowLeft':
                const previous = selected.previousElementSibling
                if (previous) {
                    clearSelected()
                    select(previous)
                }
                break
            case 'ArrowRight':
                const next = selected.nextElementSibling
                if (next) {
                    clearSelected()
                    select(next)
                }
                break
            case 'ArrowDown': {
                    const offset = parseInt(selected.getAttribute('data-offset')!) + 16
                    const el = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
                    if (el) {
                        clearSelected()
                        select(el)
                    }
                }
                break
            case 'ArrowUp': {
                    const offset = parseInt(selected.getAttribute('data-offset')!) - 16
                    const el = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
                    if (el) {
                        clearSelected()
                        select(el)
                    }
                }
                break
            case 'Escape':
                selected.classList.remove('selected')
        }
    }
    e.preventDefault()
    e.stopPropagation()
}

function mousemove(e: MouseEvent) {
    if (selection.start && !selection.end) {
        const parent = selection.start.parentNode!
        let target = <HTMLElement>e.target
        const offset = target.getAttribute('data-offset')
        if (!offset) {
            return
        }
        if (target.parentNode === text.value) {
            target = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
        }

        if (target.parentNode != parent ||selection.start == target) {
            return
        }

        clearSelection()
        
        const index1 = Array.prototype.indexOf.call(parent.children, selection.start)
        const index2 = Array.prototype.indexOf.call(parent.children, target)
        let startIndex = index1
        let endIndex = index2
        if (index2 < index1) {
            startIndex = index2
            endIndex = index1
        }

        for (let i = startIndex; i <= endIndex; i++) {
            const el = parent.children[i]
            el.classList.add('selection')
            const childOffset = el.getAttribute('data-offset')
            const spanText = <HTMLElement>text.value?.querySelector(`[data-offset="${childOffset}"]`)
            spanText.classList.add('selection')
        }
    }
}
function getSelection(): string {
    if (!selection.start) {
        return ''
    }
    const parent = selection.start.parentNode!
    const index1 = Array.prototype.indexOf.call(parent.children, selection.start)
    const index2 = Array.prototype.indexOf.call(parent.children, selection.end)

    let startIndex = index1
    let endIndex = index2
    if (index2 < index1) {
        startIndex = index2
        endIndex = index1
    }

    let result = ''
    for (let i = startIndex; i <= endIndex; i++) {
        const el = <HTMLElement>parent.children[i]
        result += el.innerText
    }
    return result
}
function createElement(from: HTMLElement) {
    const span = document.createElement('span')
    const offset = parseInt(from.getAttribute('data-offset')!)
    from.setAttribute('data-offset', (offset+1).toString())
    
    span.setAttribute('data-offset', offset.toString())
    span.classList.add('selected', 'virtual')
    
    from.parentNode!.insertBefore(span, from)
    from.classList.remove('selected')

    const spanText = document.createElement('span')
    spanText.classList.add('virtual')
    spanText.setAttribute('data-offset', offset.toString())
    text.value?.appendChild(spanText)

    return span
}
function toText(n: number): string {
    if (n < 32 || n > 127) {
        return '.'
    } else {
        return String.fromCharCode(n)
    }
}
function clearSelected() {
    let selected = hex.value?.querySelector('.selected');
    selected?.classList.remove('selected');
    selected = text.value?.querySelector('.selected');
    selected?.classList.remove('selected');
}
function clearSelection() {
    let selected = hex.value?.querySelector('.selection');
    selected?.classList.remove('selection');
    selected = text.value?.querySelector('.selection');
    selected?.classList.remove('selection');
}
function select(target: Element | null) {
    if (!target) {
        return
    }
    target.classList.add('selected')
    const offset = target.getAttribute('data-offset')
    let other: HTMLElement
    if (target.parentNode === hex.value) {
        other = <HTMLElement>text.value?.querySelector(`[data-offset="${offset}"]`)
    } else {
        other = <HTMLElement>hex.value?.querySelector(`[data-offset="${offset}"]`)
    }
    if (other) {
        other.classList.add('selected')
    }
}
onBeforeUpdate(() => {
    hex.value?.querySelectorAll('span.virtual').forEach(x => x.remove())
    text.value?.querySelectorAll('span.virtual').forEach(x => x.remove())
})
</script>

<template>
    <div class="hexedit" @mouseover="mouseover" @mouseout="mouseleave" @mouseup="mouseup" @mousedown="mousedown" @keydown="keydown" tabindex="0" @mousemove="mousemove">
        <div class="offset">
            <div v-for="r in rows()">{{ (r * 16).toString(16).padStart(4, '0') }}</div>
        </div>
        <div class="hex" ref="hex">
            <span v-for="(hex, index) in data()" :data-offset="index">{{ hex }}</span>
            <span class="last" :data-offset="data().length">+</span>
        </div>
        <div class="text" ref="text">
            <span v-for="(s, index) in string()" :data-offset="index">{{ s }}</span>
        </div>
    </div>
</template>

<style>
.hexedit:focus-visible {
    outline: none;
}
.hexedit {
    caret-color: transparent;
    -webkit-user-select: none;
    user-select: none; 
}
.hexedit > div {
    display: inline-block;
    padding-left: 7px;
    padding-right: 7px;
    vertical-align: top;
    font-family: monospace;
}
.hexedit .hex {
    min-width: 405px;
}
.hexedit .hex span {
    padding: 3px 4px 3px 4px;
    white-space: pre;
    cursor: pointer;
}
.hexedit .hex span.hover, .text span.hover {
    background-color: #a5899fff;
    text-shadow: 1px 1px 1px black;
}
.hexedit .hex span.selected, .text span.selected {
    background-color: #eabaabff;
    text-shadow: 1px 1px 1px black;
    border-bottom: solid 2px transparent;
    padding-bottom: 1px;
}
.hexedit .hex span.selected {
    animation: blink 1.2s infinite linear;
}
.hexedit .hex span.selection, .text span.selection {
    background-color: #eabaabff;
    text-shadow: 1px 1px 1px black;
}
.hexedit .hex span:nth-child(16)::after {
    content: "\a";
}
.hexedit .text {
    padding-left: 10px;
}
.hexedit .text span {
    white-space: pre;
}
.hexedit .text span:nth-child(16)::after {
    content: "\a";
}
@keyframes blink {
    0% { border-color:white;}
    60% { border-color:white;}
}
</style>
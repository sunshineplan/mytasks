import { writable } from 'svelte/store'

export interface List {
    id: number
    list: string
    incomplete: number
    completed: number
}

export interface Task {
    id: number
    task: string
    created: string
}

export const username = writable('')
export const component = writable('show')
export const current = writable({ id: 0 } as List)
export const lists = writable([] as List[])
export const tasks = writable({} as { [ListName: string]: { incomplete: Task[], completed: Task[] } })
export const showSidebar = writable(false)
export const loading = writable(0)

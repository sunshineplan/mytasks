import { writable } from 'svelte/store'

export interface List {
    list: string
    incomplete: number
    completed: number
}

export interface Task {
    id: string
    task: string
    created: string
}

export const username = writable('')
export const component = writable('show')
export const current = writable({} as List)
export const lists = writable([] as List[])
export const tasks = writable({} as { [ListName: string]: { incomplete: Task[], completed: Task[] } })
export const showSidebar = writable(false)
export const loading = writable(0)

export const reset = () => {
    username.set('')
    current.set({} as List)
    lists.set([])
    tasks.set({})
}

export const pubkey = "@pubkey@"

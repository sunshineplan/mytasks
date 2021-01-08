import { writable } from 'svelte/store'

export interface List {
    id: number
    list: string
    count: number
    seq: number
}

export interface Task {
    id: number
    task: string
    seq: number
}

export const username = writable('')
export const component = writable('tasks')
export const current = writable({ id: 0 } as List)
export const last = writable({} as List)
export const tasks = writable({} as { [ListName: string]: Task[] })
export const loading = writable(0)

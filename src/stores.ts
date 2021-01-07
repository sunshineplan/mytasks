import { writable } from 'svelte/store'

export interface List {
    id: number
    list: string
    count: number
    seq: number
}

interface Task {
    id: number
    task: string
    seq: number
}

interface Tasks {
    [ListName: string]: Task[]
}

export const username = writable('')
export const component = writable('tasks')
export const current = writable({} as List)
export const tasks = writable({} as Tasks)
export const loading = writable(0)

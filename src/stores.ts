import { writable } from 'svelte/store'

export const username = writable('')
export const component = writable('show')
export const current = writable(<List>{})
export const lists = writable(<List[]>[])
export const tasks = writable(<{ [ListName: string]: { incomplete: Task[], completed: Task[] } }>{})
export const showSidebar = writable(false)
export const loading = writable(0)

export const reset = () => {
  username.set('')
  current.set(<List>{})
  lists.set([])
  tasks.set({})
}

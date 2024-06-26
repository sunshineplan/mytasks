import { writable } from 'svelte/store'

export const component = writable('show')
export const showSidebar = writable(false)

const createLoading = () => {
  const { subscribe, update } = writable(0)
  return {
    subscribe,
    start: () => update(n => n + 1),
    end: () => update(n => n - 1)
  }
}
export const loading = createLoading()

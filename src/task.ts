import { writable } from 'svelte/store'

export const username = writable('')
export const current = writable(<List>{})
export const lists = writable(<List[]>[])
export const tasks = writable(<{ [ListName: string]: { incomplete: Task[], completed: Task[] } }>{})

export const init = async () => {
  const resp = await fetch('/info')
  if (resp.ok) {
    const info = await resp.json()
    if (Object.keys(info).length) {
      username.set(info.username)
      lists.set(info.lists)
    } else reset()
  } else if (resp.status == 409)
    await init()
  else reset()
}

const reset = () => {
  username.set('')
  current.set(<List>{})
  lists.set([])
  tasks.set({})
}

export const poll = async (signal: AbortSignal) => {
  let resp: Response
  try {
    resp = await fetch('/poll', { signal })
  } catch (e) {
    let message = ''
    if (typeof e === 'string') {
      message = e
    } else if (e instanceof Error) {
      message = e.message
    }
    resp = new Response(message, { status: 500 })
  }
  return resp
}

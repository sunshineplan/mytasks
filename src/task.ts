import { writable, get } from 'svelte/store'
import { Dexie } from 'dexie'
import { fire, post } from './misc'

const db = new Dexie('task')
db.version(1).stores({
  lists: 'list',
  tasks: 'list'
})

export const list = writable(<List>{})

const createLists = () => {
  const { subscribe, set } = writable(<List[]>[])
  return {
    subscribe,
    set,
    clear: async () => await db.table('lists').clear(),
    get: async () => {
      await db.table<List>('lists').filter(i => !i.completed && !i.incomplete).delete()
      const array = await db.table<List>('lists').toArray()
      if (array.length) lists.set(array)
      else await lists.fetch()
    },
    fetch: async () => {
      const resp = await post('/list/get')
      if (resp.ok) {
        const res = await resp.json()
        lists.set(res)
        await db.table<List>('lists').bulkAdd(res)
      } else await fire('Fatal', await resp.text(), 'error')
    },
    add: async (list: List) => {
      await db.table<List>('lists').add(list)
      const array = get(lists)
      if (array.slice(-1)[0].list == '') {
        array.splice(array.length - 1, 0, list)
        lists.set(array)
      } else lists.set([...array, list])
    },
    edit: async (name: string) => {
      const current = get(list)
      const resp = await post('/list/edit', { old: current.list, new: name })
      let msg = ''
      if (resp.ok) {
        const res = await resp.json()
        if (res.status) {
          await db.table<List>('lists').update(current.list, { list: name })
          await db.table<Tasks>('tasks').where('list').equals(current.list).modify({ list: name })
          lists.set(await db.table<List>('lists').toArray())
          current.list = name
          list.set(current)
          return 0
        } else msg = res.message
      } else msg = await resp.text()
      await fire('Fatal', msg, 'error')
      return 1
    },
    delete: async () => {
      const current = get(list)
      const resp = await post('/list/delete', { list: current.list })
      if (resp.ok) {
        await db.table<List>('lists').where('list').equals(current.list).delete()
        await db.table<Tasks>('tasks').where('list').equals(current.list).delete()
        lists.set(await db.table<List>('lists').toArray())
      } else await fire('Fatal', await resp.text(), 'error')
    }
  }
}
export const lists = createLists()

const createTasks = () => {
  const { subscribe, set } = writable(<Tasks>{})
  return {
    subscribe,
    set,
    clear: async () => await db.table('tasks').clear(),
    load: async () => {
      let current = get(list)
      if (!current.list) {
        const array = get(lists)
        if (array.length) {
          current = array[0]
          list.set(current)
        } else {
          current = { list: 'New List', incomplete: 0, completed: 0 }
          list.set(current)
          lists.set([current])
        }
      }
      const res = await db.table<Tasks>('tasks').where('list').equals(current.list).first()
      if (res) tasks.set(res)
      else tasks.set({ list: current.list, incomplete: [], completed: [] })
    },
    get: async (more?: number, goal?: number) => { // need check
      await tasks.load()
      const current = get(list)
      if (current.list) {
        const res = get(tasks)
        const total = current.completed
        const len = res.completed.length
        if (!goal)
          if (more) goal = Math.min(len + more, total)
          else goal = Math.min(10, total)
        if (len >= goal) return
        if (more) await tasks.moreCompleted(len)
      }
      if (!more) await tasks.fetch()
      await tasks.get(more, goal)
    },
    fetch: async () => {
      let current = get(list)
      const resp = await post('/task/get', { list: current.list })
      if (resp.ok) {
        const res = await resp.json()
        await db.table<Tasks>('tasks').add({ list: current.list, incomplete: res.incomplete, completed: res.completed })
      } else await fire('Fatal', await resp.text(), 'error')
    },
    moreCompleted: async (start: number) => {
      const current = get(list)
      const data = get(tasks)
      const resp = await post('/completed/more', { list: current.list, start })
      if (resp.ok) {
        const more = await resp.json()
        data.completed = data.completed.concat(more)
        await db.table<Tasks>('tasks').where('list').equals(current.list).modify({
          completed: data.completed
        })
      } else await fire('Fatal', await resp.text(), 'error')
    },
    save: async (task: Task) => {
      const current = get(list)
      let url = '/task/add'
      if (task.id) url = '/task/edit/' + task.id, task
      else task.list = current.list
      const resp = await post(url, task)
      if (resp.ok) {
        const res = await resp.json()
        if (res.status == 1) {
          const data = get(tasks)
          if (task.id) data.incomplete = data.incomplete.map(i => {
            if (i.id === task.id) i.task = task.task
            return i
          })
          else {
            data.incomplete = [
              {
                id: res.id,
                list: task.list,
                task: task.task,
                created: new Date().toLocaleString(),
                seq: res.seq
              },
              ...data.incomplete,
            ]
            await db.table<List>('lists').where('list').equals(current.list).modify(i => { i.incomplete++ })
          }
          await db.table<Tasks>('tasks').where('list').equals(current.list).modify({
            incomplete: data.incomplete
          })
          lists.set(await db.table<List>('lists').toArray())
          await tasks.get()
        } else {
          await fire('Error', res.message, 'error')
          return <number>res.error
        }
      } else await fire('Fatal', await resp.text(), 'error')
      return 0
    },
    complete: async (task: Task) => {
      const resp = await post('/task/complete/' + task.id)
      if (resp.ok) {
        const res = await resp.json()
        if (res.status && res.id) {
          const current = get(list)
          const data = get(tasks)
          const index = data.incomplete.findIndex(i => i.id == task.id)
          data.completed = [
            {
              id: res.id,
              list: current.list,
              task: data.incomplete[index].task,
              created: new Date().toLocaleString()
            },
            ...data.completed,
          ]
          data.incomplete.splice(index, 1)
          await db.table<List>('lists').where('list').equals(current.list).modify(i => {
            i.incomplete--
            i.completed++
          })
          await db.table<Tasks>('tasks').where('list').equals(current.list).modify({
            incomplete: data.incomplete,
            completed: data.completed
          })
          lists.set(await db.table<List>('lists').toArray())
          await tasks.get()
        } else await fire('Error', 'Error', 'error')
      } else await fire('Fatal', await resp.text(), 'error')
    },
    revert: async (task: Task) => {
      const resp = await post('/completed/revert/' + task.id)
      if (resp.ok) {
        const res = await resp.json()
        if (res.status && res.id) {
          const current = get(list)
          const data = get(tasks)
          const index = data.completed.findIndex(i => i.id == task.id)
          data.incomplete = [
            {
              id: res.id,
              list: current.list,
              task: data.completed[index].task,
              created: new Date().toLocaleString(),
              seq: res.seq
            },
            ...data.incomplete,
          ]
          data.completed.splice(index, 1)
          current.incomplete++
          current.completed--
          await db.table<Tasks>('tasks').where('list').equals(current.list).modify({
            incomplete: data.incomplete,
            completed: data.completed
          })
          await db.table<List>('lists').where('list').equals(current.list).modify(current)
          list.set(current)
          lists.set(await db.table<List>('lists').toArray())
          await tasks.get()
        } else await fire('Error', 'Error', 'error')
      } else await fire('Fatal', await resp.text(), 'error')
    },
    delete: async (task: Task, done?: boolean) => {
      let url = '/task/delete/'
      if (done) url = '/completed/delete/'
      const resp = await post(url + task.id)
      if (resp.ok) {
        const current = get(list)
        const data = get(tasks)
        if (done) {
          data.completed = data.completed.filter(i => i.id != task.id)
          current.completed--
        } else {
          data.incomplete = data.incomplete.filter(i => i.id != task.id)
          current.incomplete--
        }
        await db.table<Tasks>('tasks').where('list').equals(current.list).modify({
          incomplete: data.incomplete,
          completed: data.completed
        })
        await db.table<List>('lists').where('list').equals(current.list).modify(current)
        list.set(current)
        lists.set(await db.table<List>('lists').toArray())
        await tasks.get()
      } else await fire('Fatal', await resp.text(), 'error')
    },
    swap: async (a: Task, b: Task) => {
      const current = get(list)
      const resp = await post('/task/reorder', { list: current.list, orig: a.id, dest: b.id })
      if (resp.ok) {
        if ((await resp.text()) == '1') {
          const data = get(tasks)
          const seq = b.seq
          if (a.seq! > b.seq!) data.incomplete.forEach(i => { if (i.seq! >= b.seq! && i.seq! < a.seq!) i.seq!++ })
          else data.incomplete.forEach(i => { if (i.seq! > a.seq! && i.seq! <= b.seq!) i.seq!-- })
          data.incomplete.forEach(i => { if (i.id === a.id) i.seq = seq })
          data.incomplete.sort((a, b) => b.seq! - a.seq!)
          await db.table<Tasks>('tasks').where('list').equals(current.list).modify({
            incomplete: data.incomplete
          })
        } else await fire('Fatal', 'Failed to reorder.', 'error')
      } else await fire('Fatal', await resp.text(), 'error')
    },
    empty: async () => {
      const current = get(list)
      const resp = await post('/completed/empty', { list: current.list })
      if (resp.ok) {
        await db.table<Tasks>('tasks').where('list').equals(current.list).modify({ completed: [] })
        current.completed = 0
        await db.table<List>('lists').where('list').equals(current.list).modify(current)
        list.set(current)
        lists.set(await db.table<List>('lists').toArray())
        await tasks.get()
        return 0
      } else await fire('Fatal', await resp.text(), 'error')
      return 1
    }
  }
}
export const tasks = createTasks()

export const init = async (): Promise<string> => {
  const resp = await fetch('/info')
  if (resp.ok) {
    const username = await resp.text()
    if (username) {
      await lists.get()
      await tasks.get()
      return username
    } else await reset()
  } else if (resp.status == 409) {
    await lists.clear()
    await tasks.clear()
    return await init()
  } else await reset()
  return ''
}

const reset = async () => {
  list.set(<List>{})
  lists.set([])
  tasks.set(<Tasks>{})
  await lists.clear()
  await tasks.clear()
}

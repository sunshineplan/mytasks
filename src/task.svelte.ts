import { Dexie } from 'dexie'
import { getCookie } from 'typescript-cookie'
import { fire, loading, post } from './misc.svelte'

const db = new Dexie('task')
db.version(1).stores({
  lists: 'list',
  tasks: 'list'
})
const listTable = db.table<List>('lists')
const taskTable = db.table<Tasks>('tasks')

class MyTasks {
  username = $state('')
  component = $state('show')
  list = $state<List>({} as List)
  lists = $state<List[]>([])
  tasks = $state<Tasks>({} as Tasks)
  controller = $state(new AbortController())
  async clear() {
    await listTable.clear()
    await taskTable.clear()
  }
  async reset() {
    this.list = {} as List
    this.lists = []
    this.tasks = {} as Tasks
    await this.clear()
  }
  async init() {
    loading.start()
    let resp: Response
    try {
      resp = await fetch('/info')
    } catch (e) {
      console.error(e)
      resp = new Response(null, { "status": 500 })
    }
    loading.end()
    if (resp.ok) {
      const username = await resp.text()
      if (username) {
        await this.#getLists()
        await this.getTasks()
        this.username = username
      } else await this.reset()
    } else if (resp.status == 409) {
      await this.clear()
      await this.init()
    } else await this.reset()
  }
  async #getLists() {
    await listTable.filter(i => !i.completed && !i.incomplete).delete()
    const array = await listTable.toArray()
    if (array.length) this.lists = array
    else await this.#fetchLists()
  }
  async #fetchLists() {
    const resp = await post('/list/get')
    if (resp.ok) {
      const res = await resp.json()
      this.lists = res
      await listTable.bulkAdd(res)
    } else await fire('Fatal', await resp.text(), 'error')
  }
  async addList(list: List) {
    await listTable.add(list)
    if (this.lists.slice(-1)[0].list == '')
      this.lists.splice(this.lists.length - 1, 0, list)
    else this.lists = [...this.lists, list]
  }
  async editList(name: string) {
    this.controller.abort()
    const resp = await post('/list/edit', { old: this.list.list, new: name })
    let msg = ''
    if (resp.ok) {
      const res = await resp.json()
      if (res.status) {
        await listTable.update(this.list.list, { list: name })
        await taskTable.where('list').equals(this.list.list).modify({ list: name })
        this.lists = await listTable.toArray()
        this.list.list = name
        this.subscribe(true)
        return 0
      } else msg = res.message
    } else msg = await resp.text()
    await fire('Fatal', msg, 'error')
    this.subscribe(true)
    return 1
  }
  async deleteList() {
    this.controller.abort()
    const resp = await post('/list/delete', { list: this.list.list })
    if (resp.ok) {
      await listTable.where('list').equals(this.list.list).delete()
      await taskTable.where('list').equals(this.list.list).delete()
      this.lists = await listTable.toArray()
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async #loadTasks() {
    if (!this.list.list) {
      if (this.lists.length)
        this.list = this.lists[0]
      else {
        this.list = { list: 'New List', incomplete: 0, completed: 0 }
        this.lists = [this.list]
      }
    }
    const res = await taskTable.where('list').equals(this.list.list).first()
    if (res) this.tasks = res
    else this.tasks = { list: this.list.list, incomplete: [], completed: [] }
  }
  async getTasks(more?: number, goal?: number) { // need check
    await this.#loadTasks()
    if (this.list.list) {
      const total = this.list.completed
      const len = this.tasks.completed.length
      if (!goal)
        if (more) goal = Math.min(len + more, total)
        else goal = Math.min(10, total)
      if (len >= goal) return
      if (more) await this.moreCompleted(len)
    }
    if (!more) await this.#fetchTasks()
    await this.getTasks(more, goal)
  }
  async #fetchTasks() {
    const resp = await post('/task/get', { list: this.list.list })
    if (resp.ok) {
      const res = await resp.json()
      await taskTable.add({ list: this.list.list, incomplete: res.incomplete, completed: res.completed })
    } else await fire('Fatal', await resp.text(), 'error')
  }
  async moreCompleted(start: number) {
    const resp = await post('/completed/more', { list: this.list.list, start })
    if (resp.ok) {
      const more = await resp.json()
      this.tasks.completed = this.tasks.completed.concat(more)
      await taskTable.where('list').equals(this.list.list).modify({
        completed: this.tasks.completed
      })
    } else await fire('Fatal', await resp.text(), 'error')
  }
  async saveTask(task: Task) {
    let url = '/task/add'
    if (task.id) {
      const old = this.tasks.incomplete.find(i => i.id == task.id)
      if (old!.task == task.task) {
        return 0
      }
      url = '/task/edit/' + task.id
    }
    else task.list = this.list.list
    this.controller.abort()
    const resp = await post(url, task)
    if (resp.ok) {
      const res = await resp.json()
      if (res.status == 1) {
        if (task.id) this.tasks.incomplete = this.tasks.incomplete.map(i => {
          if (i.id === task.id) i.task = task.task
          return i
        })
        else {
          this.tasks.incomplete = [
            {
              id: res.id,
              list: task.list,
              task: task.task,
              created: new Date().toLocaleString(),
              seq: res.seq
            },
            ...this.tasks.incomplete,
          ]
          await listTable.where('list').equals(this.list.list).modify(i => { i.incomplete++ })
        }
        await taskTable.where('list').equals(this.list.list).modify({
          incomplete: $state.snapshot(this.tasks.incomplete)
        })
        this.lists = await listTable.toArray()
        await this.getTasks()
      } else {
        await fire('Error', res.message, 'error')
        this.subscribe(true)
        return <number>res.error
      }
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
    return 0
  }
  async completeTask(task: Task) {
    this.controller.abort()
    const resp = await post('/task/complete/' + task.id)
    if (resp.ok) {
      const res = await resp.json()
      if (res.status && res.id) {
        const index = this.tasks.incomplete.findIndex(i => i.id == task.id)
        this.tasks.completed = [
          {
            id: res.id,
            list: this.list.list,
            task: this.tasks.incomplete[index].task,
            created: new Date().toLocaleString()
          },
          ...this.tasks.completed,
        ]
        this.tasks.incomplete.splice(index, 1)
        await listTable.where('list').equals(this.list.list).modify(i => {
          i.incomplete--
          i.completed++
        })
        await taskTable.where('list').equals(this.list.list).modify({
          incomplete: $state.snapshot(this.tasks.incomplete),
          completed: $state.snapshot(this.tasks.completed)
        })
        this.lists = await listTable.toArray()
        await this.getTasks()
      } else await fire('Error', 'Error', 'error')
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async revertTask(task: Task) {
    this.controller.abort()
    const resp = await post('/completed/revert/' + task.id)
    if (resp.ok) {
      const res = await resp.json()
      if (res.status && res.id) {
        const index = this.tasks.completed.findIndex(i => i.id == task.id)
        this.tasks.incomplete = [
          {
            id: res.id,
            list: this.list.list,
            task: this.tasks.completed[index].task,
            created: new Date().toLocaleString(),
            seq: res.seq
          },
          ...this.tasks.incomplete,
        ]
        this.tasks.completed.splice(index, 1)
        this.list.incomplete++
        this.list.completed--
        await taskTable.where('list').equals(this.list.list).modify({
          incomplete: $state.snapshot(this.tasks.incomplete),
          completed: $state.snapshot(this.tasks.completed)
        })
        await listTable.where('list').equals(this.list.list).modify(this.list)
        this.lists = await listTable.toArray()
        await this.getTasks()
      } else await fire('Error', 'Error', 'error')
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async deleteTask(task: Task, done?: boolean) {
    let url = '/task/delete/'
    if (done) url = '/completed/delete/'
    this.controller.abort()
    const resp = await post(url + task.id)
    if (resp.ok) {
      if (done) {
        this.tasks.completed = this.tasks.completed.filter(i => i.id != task.id)
        this.list.completed--
      } else {
        this.tasks.incomplete = this.tasks.incomplete.filter(i => i.id != task.id)
        this.list.incomplete--
      }
      await taskTable.where('list').equals(this.list.list).modify({
        incomplete: $state.snapshot(this.tasks.incomplete),
        completed: $state.snapshot(this.tasks.completed)
      })
      await listTable.where('list').equals(this.list.list).modify(this.list)
      this.lists = await listTable.toArray()
      await this.getTasks()
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async swapTask(a: Task, b: Task) {
    this.controller.abort()
    const resp = await post('/task/reorder', { list: this.list.list, orig: a.id, dest: b.id })
    if (resp.ok) {
      if ((await resp.text()) == '1') {
        const seq = b.seq
        if (a.seq! > b.seq!) this.tasks.incomplete.forEach(i => { if (i.seq! >= b.seq! && i.seq! < a.seq!) i.seq!++ })
        else this.tasks.incomplete.forEach(i => { if (i.seq! > a.seq! && i.seq! <= b.seq!) i.seq!-- })
        this.tasks.incomplete.forEach(i => { if (i.id === a.id) i.seq = seq })
        this.tasks.incomplete.sort((a, b) => b.seq! - a.seq!)
        await taskTable.where('list').equals(this.list.list).modify({
          incomplete: $state.snapshot(this.tasks.incomplete)
        })
      } else await fire('Fatal', 'Failed to reorder.', 'error')
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async empty() {
    this.controller.abort()
    const resp = await post('/completed/empty', { list: this.list.list })
    if (resp.ok) {
      await taskTable.where('list').equals(this.list.list).modify({ completed: [] })
      this.list.completed = 0
      await listTable.where('list').equals(this.list.list).modify(this.list)
      this.lists = await listTable.toArray()
      await this.getTasks()
      this.subscribe(true)
      return 0
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
    return 1
  }
  async subscribe(reset?: boolean) {
    if (reset)
      this.controller = new AbortController()
    let resp: Response
    try {
      resp = await fetch('/poll', { signal: this.controller.signal })
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      console.error(e)
      resp = new Response(null, { status: 500 })
    }
    if (resp.ok) {
      const last = await resp.text()
      if (last && getCookie('last') != last) {
        await this.init()
      }
      await this.subscribe()
    } else if (resp.status == 401) {
      await this.init()
    } else {
      await new Promise((sleep) => setTimeout(sleep, 30000))
      await this.subscribe()
    }
  };
}
export const mytasks = new MyTasks
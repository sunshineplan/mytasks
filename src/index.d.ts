declare interface Window {
  universal: string
  pubkey: string
}

declare interface List {
  list: string
  incomplete: number
  completed: number
}

declare interface Task {
  id: string
  list: string
  task: string
  created: string
  seq?: number
}

declare interface Tasks {
  list: string
  incomplete: Task[]
  completed: Task[]
}

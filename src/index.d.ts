interface Window {
  universal: string
  pubkey: string
}

interface List {
  list: string
  incomplete: number
  completed: number
}

interface Task {
  id: string
  list: string
  task: string
  created: string
  seq?: number
}

interface Tasks {
  list: string
  incomplete: Task[]
  completed: Task[]
}

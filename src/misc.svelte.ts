import JSEncrypt from 'jsencrypt'
import Swal from 'sweetalert2'

class Toggler {
  status = $state(false)
  toggle() { this.status = !this.status }
  close() { this.status = false }
}
export const showSidebar = new Toggler

class Loading {
  #n = $state(0)
  show = $derived(this.#n > 0)
  start() { this.#n += 1 }
  end() { this.#n -= 1 }
}
export const loading = new Loading

export const encrypt = (pubkey: string, password: string) => {
  const encrypt = new JSEncrypt()
  encrypt.setPublicKey(pubkey)
  const s = encrypt.encrypt(password)
  if (s === false) return password
  return s
}

export const fire = (
  title?: string | undefined,
  html?: string | undefined,
  icon?: 'success' | 'error' | 'warning' | 'info' | 'question' | undefined
) => {
  const swal = Swal.mixin({
    customClass: { confirmButton: 'swal btn btn-primary' },
    buttonsStyling: false
  })
  return swal.fire(title, html, icon)
}

export const confirm = async (content: string) => {
  const confirm = await Swal.fire({
    title: 'Are you sure?',
    text: content + ' will be deleted permanently.',
    icon: 'warning',
    confirmButtonText: 'Delete',
    showCancelButton: true,
    focusCancel: true,
    customClass: {
      confirmButton: 'swal btn btn-danger',
      cancelButton: 'swal btn btn-primary'
    },
    buttonsStyling: false
  })
  return confirm.isConfirmed
}

export const valid = () => {
  let result = true
  Array.from(document.querySelectorAll('input'))
    .forEach(i => { if (!i.checkValidity()) result = false })
  return result
}

export const post = async (url: string, data?: object, universal?: boolean) => {
  let resp: Response
  const init: RequestInit = {
    method: 'post',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  }
  if (universal) init.credentials = 'include'
  loading.start()
  try {
    resp = await fetch(url, init)
  } catch (e) {
    let message = ''
    if (typeof e === 'string') {
      message = e
    } else if (e instanceof Error) {
      message = e.message
    }
    resp = new Response(message, { status: 500 })
  }
  loading.end()
  if (resp.status == 401) {
    await fire('Error', 'Login status has changed. Please Re-login!', 'error')
    window.location.href = '/'
  } else if (resp.status == 409) {
    await fire('Error', 'Data has changed.', 'error')
    window.location.href = '/'
  }
  return resp
}

export const pasteText = (event: ClipboardEvent) => {
  if (document.execCommand) {
    event.preventDefault()
    if (event.clipboardData) {
      document.execCommand('insertHTML', false, event.clipboardData.getData('text/plain'))
    }
  }
}

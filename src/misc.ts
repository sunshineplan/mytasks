import Swal from 'sweetalert2'

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

export const valid = () => {
  let result = true
  Array.from(document.querySelectorAll('input'))
    .forEach(i => { if (!i.checkValidity()) result = false })
  return result
}

export const post = async (url: string, data?: object) => {
  let resp: Response
  try {
    resp = await fetch(url, {
      method: 'post',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
  } catch (e) {
    return Promise.reject(await fire('Error', e, 'error'))
  }
  if (resp.status == 401) {
    await fire('Error', 'Login status has changed. Please Re-login!', 'error')
    window.location.href = '/'
  }
  return resp
}


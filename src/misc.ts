import Swal from 'sweetalert2'

export const BootstrapButtons = Swal.mixin({
  customClass: { confirmButton: 'swal btn btn-primary' },
  buttonsStyling: false
})

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
    return Promise.reject(await BootstrapButtons.fire('Error', e, 'error'))
  }
  if (resp.status == 401) {
    await BootstrapButtons.fire('Error', 'Login status has changed. Please Re-login!', 'error')
    window.location.href = '/'
  }
  return resp
}


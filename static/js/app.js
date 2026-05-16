document.addEventListener('click', (event) => {
  const target = event.target.closest('[data-confirm]')
  if (!target) return
  if (!window.confirm(target.dataset.confirm)) {
    event.preventDefault()
  }
})

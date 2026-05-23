document.addEventListener('click', (event) => {
  const target = event.target.closest('[data-confirm]')
  if (!target) return
  if (!window.confirm(target.dataset.confirm)) {
    event.preventDefault()
  }
})

const drawerRoot = document.querySelector('[data-drawer]')
const drawerToggle = document.querySelector('[data-drawer-toggle]')

const setDrawerOpen = (open) => {
  if (!drawerRoot) return
  drawerRoot.dataset.open = open ? 'true' : 'false'
  drawerRoot.setAttribute('aria-hidden', open ? 'false' : 'true')
  document.body.style.overflow = open ? 'hidden' : ''
  if (drawerToggle) drawerToggle.setAttribute('aria-expanded', open ? 'true' : 'false')
}

drawerToggle?.addEventListener('click', () => {
  setDrawerOpen(drawerRoot?.dataset.open !== 'true')
})

document.querySelectorAll('[data-drawer-close]').forEach((node) => {
  node.addEventListener('click', () => setDrawerOpen(false))
})

document.addEventListener('keydown', (event) => {
  if (event.key === 'Escape' && drawerRoot?.dataset.open === 'true') {
    setDrawerOpen(false)
  }
})

document.querySelectorAll('.admin-drawer-panel a').forEach((node) => {
  node.addEventListener('click', () => setDrawerOpen(false))
})

document.addEventListener('click', (event) => {
  const target = event.target.closest('[data-toggle-panel]')
  if (!target) return
  const panel = document.getElementById(target.dataset.togglePanel)
  if (panel) panel.classList.toggle('hidden')
})

const roomViewTimers = new WeakMap()

const roomViewRoots = (scope) => {
  const current = scope?.matches?.('[data-room-view-root]') ? [scope] : []
  const nested = scope?.querySelectorAll ? Array.from(scope.querySelectorAll('[data-room-view-root]')) : []
  return [...current, ...nested]
}

const roomURLWithView = (href, view) => {
  try {
    const url = new URL(href, window.location.origin)
    if (url.pathname !== '/admin/rooms') return href
    url.searchParams.set('view', view)
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return href
  }
}

const syncRoomFilterLinks = (root, view) => {
  root.querySelectorAll('a[href^="/admin/rooms"]').forEach((link) => {
    if (link.matches('[data-room-view-link]')) return
    const nextHref = roomURLWithView(link.getAttribute('href'), view)
    link.setAttribute('href', nextHref)
    if (link.hasAttribute('hx-get')) link.setAttribute('hx-get', nextHref)
  })
}

const setRoomView = (root, view, replaceURL = false) => {
  if (!root || !view) return
  root.dataset.roomView = view
  root.querySelectorAll('[data-room-view-link]').forEach((link) => {
    const active = link.dataset.roomViewValue === view
    link.classList.toggle('public-room-view-active', active)
    if (active) {
      link.setAttribute('aria-current', 'page')
    } else {
      link.removeAttribute('aria-current')
    }
  })
  const viewInput = root.querySelector('input[name="view"]')
  if (viewInput) viewInput.value = view
  syncRoomFilterLinks(root, view)
  const activeLink = root.querySelector(`[data-room-view-link][data-room-view-value="${view}"]`)
  if (replaceURL && activeLink?.href && window.history?.replaceState) {
    window.history.replaceState(null, '', activeLink.href)
  }
}

const scrollToRoomView = (root, view, behavior = 'auto') => {
  const swipe = root?.querySelector('[data-room-view-swipe]')
  const pane = root?.querySelector(`[data-room-view-pane="${view}"]`)
  if (!swipe || !pane) return
  swipe.scrollTo({ left: pane.offsetLeft, behavior })
}

const initRoomViews = (scope = document) => {
  roomViewRoots(scope).forEach((root) => {
    if (root.dataset.roomViewReady === 'true') return
    root.dataset.roomViewReady = 'true'
    const initialView = root.dataset.roomView || 'list'
    window.requestAnimationFrame(() => scrollToRoomView(root, initialView))
    const swipe = root.querySelector('[data-room-view-swipe]')
    swipe?.addEventListener('scroll', () => {
      window.clearTimeout(roomViewTimers.get(root))
      const timer = window.setTimeout(() => {
        const panes = Array.from(root.querySelectorAll('[data-room-view-pane]'))
        const nearestPane = panes.reduce((nearest, pane) => {
          if (!nearest) return pane
          const currentDistance = Math.abs(pane.offsetLeft - swipe.scrollLeft)
          const nearestDistance = Math.abs(nearest.offsetLeft - swipe.scrollLeft)
          return currentDistance < nearestDistance ? pane : nearest
        }, null)
        setRoomView(root, nearestPane?.dataset.roomViewPane || 'list', true)
      }, 120)
      roomViewTimers.set(root, timer)
    })
  })
}

document.addEventListener('click', (event) => {
  const link = event.target.closest('[data-room-view-link]')
  if (!link) return
  const root = link.closest('[data-room-view-root]')
  if (!root) return
  event.preventDefault()
  const nextView = link.dataset.roomViewValue || 'list'
  setRoomView(root, nextView, true)
  scrollToRoomView(root, nextView, 'smooth')
})

const tenantViewTimers = new WeakMap()

const tenantViewRoots = (scope) => {
  const current = scope?.matches?.('[data-tenant-view-root]') ? [scope] : []
  const nested = scope?.querySelectorAll ? Array.from(scope.querySelectorAll('[data-tenant-view-root]')) : []
  return [...current, ...nested]
}

const tenantURLWithView = (href, view) => {
  try {
    const url = new URL(href, window.location.origin)
    if (url.pathname !== '/admin/tenants') return href
    url.searchParams.set('view', view)
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return href
  }
}

const syncTenantFilterLinks = (root, view) => {
  root.querySelectorAll('a[href^="/admin/tenants"]').forEach((link) => {
    if (link.matches('[data-tenant-view-link]')) return
    const nextHref = tenantURLWithView(link.getAttribute('href'), view)
    link.setAttribute('href', nextHref)
    if (link.hasAttribute('hx-get')) link.setAttribute('hx-get', nextHref)
  })
}

const setTenantView = (root, view, replaceURL = false) => {
  if (!root || !view) return
  root.dataset.tenantView = view
  root.querySelectorAll('[data-tenant-view-link]').forEach((link) => {
    const active = link.dataset.tenantViewValue === view
    link.classList.toggle('public-room-view-active', active)
    if (active) {
      link.setAttribute('aria-current', 'page')
    } else {
      link.removeAttribute('aria-current')
    }
  })
  const viewInput = root.querySelector('input[name="view"]')
  if (viewInput) viewInput.value = view
  syncTenantFilterLinks(root, view)
  const activeLink = root.querySelector(`[data-tenant-view-link][data-tenant-view-value="${view}"]`)
  if (replaceURL && window.history?.replaceState) {
    window.history.replaceState(null, '', activeLink?.href || tenantURLWithView(window.location.href, view))
  }
}

const scrollToTenantView = (root, view, behavior = 'auto') => {
  const swipe = root?.querySelector('[data-tenant-view-swipe]')
  const pane = root?.querySelector(`[data-tenant-view-pane="${view}"]`)
  if (!swipe || !pane) return
  swipe.scrollTo({ left: pane.offsetLeft, behavior })
}

const initTenantViews = (scope = document) => {
  tenantViewRoots(scope).forEach((root) => {
    if (root.dataset.tenantViewReady === 'true') return
    root.dataset.tenantViewReady = 'true'
    const initialView = root.dataset.tenantView || 'list'
    window.requestAnimationFrame(() => scrollToTenantView(root, initialView))
    const swipe = root.querySelector('[data-tenant-view-swipe]')
    swipe?.addEventListener('scroll', () => {
      window.clearTimeout(tenantViewTimers.get(root))
      const timer = window.setTimeout(() => {
        const panes = Array.from(root.querySelectorAll('[data-tenant-view-pane]'))
        const nearestPane = panes.reduce((nearest, pane) => {
          if (!nearest) return pane
          const currentDistance = Math.abs(pane.offsetLeft - swipe.scrollLeft)
          const nearestDistance = Math.abs(nearest.offsetLeft - swipe.scrollLeft)
          return currentDistance < nearestDistance ? pane : nearest
        }, null)
        setTenantView(root, nearestPane?.dataset.tenantViewPane || 'list', true)
      }, 120)
      tenantViewTimers.set(root, timer)
    })
  })
}

document.addEventListener('click', (event) => {
  const link = event.target.closest('[data-tenant-view-link]')
  if (!link) return
  const root = link.closest('[data-tenant-view-root]')
  if (!root) return
  event.preventDefault()
  const nextView = link.dataset.tenantViewValue || 'list'
  setTenantView(root, nextView, true)
  scrollToTenantView(root, nextView, 'smooth')
})

document.addEventListener('htmx:afterSwap', (event) => {
  initRoomViews(event.target)
  initTenantViews(event.target)
  initRoomComboBoxes(event.target)
  initTenantSearchComboBoxes(event.target)
})

initRoomViews()
initTenantViews()

document.querySelectorAll('[data-gallery]').forEach((gallery) => {
  const frames = gallery.querySelectorAll('[data-gallery-main]')
  const thumbs = gallery.querySelectorAll('[data-gallery-thumb]')
  const scrollArea = gallery.querySelector('[data-gallery-scroll]')
  const selectFrame = (index) => {
    frames.forEach((frame) => {
      if (!scrollArea) frame.classList.toggle('hidden', frame.dataset.galleryMain !== index)
      if (scrollArea && frame.dataset.galleryMain === index) frame.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' })
    })
    thumbs.forEach((thumb) => {
      const active = thumb.dataset.galleryThumb === index
      thumb.classList.toggle('border-amber-700', active)
      thumb.classList.toggle('scale-105', active)
    })
  }
  thumbs.forEach((thumb) => {
    thumb.addEventListener('click', () => selectFrame(thumb.dataset.galleryThumb))
  })
  if (thumbs[0]) selectFrame(thumbs[0].dataset.galleryThumb)
})

document.addEventListener('click', (event) => {
  const target = event.target.closest('[data-lightbox-src]')
  if (!target) return
  const src = target.dataset.lightboxSrc
  if (!src) return
  const overlay = document.createElement('div')
  overlay.className = 'fixed inset-0 z-50 flex items-center justify-center bg-black/85 p-4'
  const closeButton = document.createElement('button')
  closeButton.className = 'absolute right-5 top-5 rounded-full bg-white px-4 py-2 font-black text-stone-900'
  closeButton.type = 'button'
  closeButton.textContent = '关闭'
  const image = document.createElement('img')
  image.className = 'max-h-full max-w-full rounded-2xl object-contain'
  image.src = src
  image.alt = '房源图片'
  overlay.append(closeButton, image)
  overlay.addEventListener('click', (overlayEvent) => {
    if (overlayEvent.target === overlay || overlayEvent.target === closeButton) overlay.remove()
  })
  document.body.appendChild(overlay)
})

document.querySelectorAll('[data-counter-target]').forEach((field) => {
  const target = document.getElementById(field.dataset.counterTarget)
  const update = () => {
    if (!target) return
    const length = [...field.value].length
    target.textContent = `${length} / 500 字`
    target.classList.toggle('text-red-700', length > 500)
  }
  field.addEventListener('input', update)
  update()
})

document.querySelectorAll('[data-tenant-filter]').forEach((field) => {
  const select = document.getElementById(field.dataset.tenantFilter)
  if (!select) return
  field.addEventListener('input', () => {
    const query = field.value.trim().toLowerCase()
    Array.from(select.options).forEach((option) => {
      if (!option.value) return
      const text = `${option.textContent} ${option.dataset.search || ''}`.toLowerCase()
      option.hidden = query !== '' && !text.includes(query)
    })
  })
})

const updateRentLabels = () => {
  document.querySelectorAll('[data-rent-type]').forEach((select) => {
    const daily = select.value === 'daily'
    const form = select.closest('form') || document
    const roomLabel = form.querySelector('#rentAmountLabel')
    const tenantLabel = form.querySelector('#tenantRentAmountLabel')
    if (roomLabel) roomLabel.innerHTML = daily ? '租金金额（元/天） <span class="text-red-600">*</span>' : '租金金额（元/月） <span class="text-red-600">*</span>'
    if (tenantLabel) tenantLabel.innerHTML = daily ? '约定租金（元/天） <span class="text-red-600">*</span>' : '约定租金（元/月） <span class="text-red-600">*</span>'
  })
}

document.querySelectorAll('[data-rent-type]').forEach((select) => {
  select.addEventListener('change', updateRentLabels)
})
updateRentLabels()

const extractURL = (value) => {
  const match = String(value || '').match(/https?:\/\/[^\s<>"']+/i)
  if (!match) return ''
  return match[0].replace(/[，。；、）】),.;!?]+$/, '')
}

const normalizeURLField = (field) => {
  const url = extractURL(field.value)
  if (url && field.value !== url) {
    field.value = url
    field.dispatchEvent(new Event('input', { bubbles: true }))
  }
  return url
}

const validators = {
  required: (value, label) => value.trim() ? '' : `${label}不能为空`,
  name: (value) => /^[一-龥A-Za-z]{2,20}$/.test(value.trim()) ? '' : '姓名需填写 2-20 个中文或英文字母',
  phone: (value) => /^1[3-9]\d{9}$/.test(value.trim()) ? '' : '手机号格式不正确',
  optional_phone: (value) => value.trim() === '' || /^1[3-9]\d{9}$/.test(value.trim()) ? '' : '紧急联系人手机号格式不正确',
  room_no: (value) => /^[A-Za-z0-9]{1,10}$/.test(value.trim()) ? '' : '房间号需为 1-10 位字母或数字',
  address: (value) => value.trim().length >= 5 ? '' : '地址至少需要 5 个字符',
  positive_integer: (value, label) => /^[1-9]\d*$/.test(value.trim()) ? '' : `${label}需为大于 0 的整数`,
  non_negative_integer: (value, label) => /^\d+$/.test(value.trim()) ? '' : `${label}需为大于或等于 0 的整数`,
  positive_amount: (value, label) => /^(?:[1-9]\d*|0?\.\d{1,2}|[1-9]\d*\.\d{1,2})$/.test(value.trim()) ? '' : `${label}需为大于 0 的金额`,
  area: (value) => /^\d+$/.test(value.trim()) && Number(value) >= 1 && Number(value) <= 9999 ? '' : '面积需为 1-9999 之间的整数',
  integer: (value, label) => /^-?\d+$/.test(value.trim()) ? '' : `${label}需为整数`,
  range_0_20: (value, label) => /^\d+$/.test(value.trim()) && Number(value) >= 0 && Number(value) <= 20 ? '' : `${label}需为 0-20 之间的整数`,
  video_link: (value) => value.trim() === '' || /^https?:\/\//.test(value.trim()) ? '' : '视频链接需以 http:// 或 https:// 开头',
  notes: (value) => [...value.trim()].length <= 1000 ? '' : '备注不能超过 1000 字'
}

const showFieldError = (field, message) => {
  const parent = field.closest('label') || field.parentElement
  let error = parent.querySelector('[data-field-error]')
  if (!error) {
    error = document.createElement('p')
    error.dataset.fieldError = 'true'
    error.className = 'mt-1 text-xs font-bold text-red-700'
    parent.appendChild(error)
  }
  error.textContent = message
  error.classList.toggle('hidden', message === '')
  field.classList.toggle('border-red-500', message !== '')
}

const validateField = (field) => {
  const rule = field.dataset.validate
  const validator = validators[rule]
  if (!validator) return true
  const label = field.dataset.label || '该字段'
  const message = validator(field.value, label)
  showFieldError(field, message)
  return message === ''
}

document.querySelectorAll('[data-url-extract]').forEach((field) => {
  let checkedClipboard = false
  const readClipboardURL = () => {
    if (checkedClipboard || field.value.trim() !== '' || !navigator.clipboard?.readText) return
    navigator.clipboard.readText()
      .then((text) => {
        const url = extractURL(text)
        checkedClipboard = true
        if (url && field.value.trim() === '') {
          field.value = url
          field.dispatchEvent(new Event('input', { bubbles: true }))
          validateField(field)
        }
      })
      .catch(() => {})
  }
  field.addEventListener('click', readClipboardURL)
  field.addEventListener('input', () => normalizeURLField(field))
  field.addEventListener('blur', () => normalizeURLField(field))
})

document.querySelectorAll('[data-validate]').forEach((field) => {
  field.addEventListener('blur', () => validateField(field))
})

document.querySelectorAll('[data-validate-form]').forEach((form) => {
  form.addEventListener('submit', (event) => {
    const results = Array.from(form.querySelectorAll('[data-validate]')).map(validateField)
    if (!results.every(Boolean)) event.preventDefault()
  })
})

const escapeHTML = (value) => String(value ?? '').replace(/[&<>'"]/g, (char) => ({
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  "'": '&#39;',
  '"': '&quot;'
}[char]))

const formatDisplayDate = (value) => {
  if (!value || value === '-') return '-'
  const text = String(value)
  if (/^\d{4}-\d{2}-\d{2}/.test(text)) return text.slice(0, 10).replaceAll('-', '/')
  return text
}

let activeTenantsRequest
const loadActiveTenants = () => {
  if (!activeTenantsRequest) {
    activeTenantsRequest = fetch('/api/tenants/active')
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('读取租客失败')))
      .catch(() => [])
  }
  return activeTenantsRequest
}

const tenantSearchText = (tenant) => `${tenant.name || ''} ${tenant.room_no || ''} ${tenant.phone || ''}`.toLowerCase()

const initTenantComboBox = (root) => {
  const input = root.querySelector('[data-tenant-input]')
  const hidden = root.querySelector('[data-tenant-hidden]')
  const list = root.querySelector('[data-tenant-list]')
  if (!input || !hidden || !list) return

  let tenants = []
  let filtered = []
  let activeIndex = 0
  const allowAll = root.dataset.tenantAllowAll === 'true'

  const close = () => list.classList.add('hidden')
  const open = () => {
    list.classList.remove('hidden')
    render()
  }
  const selectTenant = (tenant) => {
    hidden.value = tenant.id || ''
    input.value = tenant.id ? tenant.name : ''
    showFieldError(hidden, '')
    hidden.dispatchEvent(new Event('change', { bubbles: true }))
    close()
  }
  const updateFiltered = () => {
    const query = input.value.trim().toLowerCase()
    filtered = tenants.filter((tenant) => query === '' || tenantSearchText(tenant).includes(query))
    if (allowAll) filtered = [{ id: '', name: '全部租客', room_no: '', phone: '' }, ...filtered]
    activeIndex = Math.min(activeIndex, Math.max(filtered.length - 1, 0))
  }
  const render = () => {
    updateFiltered()
    if (filtered.length === 0) {
      list.innerHTML = '<div class="px-4 py-3 text-sm text-stone-500">未找到匹配租客</div>'
      return
    }
    list.innerHTML = filtered.map((tenant, index) => {
      const active = index === activeIndex ? 'bg-amber-50 text-amber-900' : 'hover:bg-stone-50'
      const label = tenant.id ? `${escapeHTML(tenant.name)} — ${escapeHTML(tenant.room_no)} — ${escapeHTML(tenant.phone)}` : '全部租客'
      return `<button class="block w-full px-4 py-3 text-left text-sm font-semibold ${active}" type="button" data-tenant-option="${index}">${label}</button>`
    }).join('')
  }

  loadActiveTenants().then((items) => {
    tenants = Array.isArray(items) ? items : []
  })

  input.addEventListener('focus', () => {
    loadActiveTenants().then((items) => {
      tenants = Array.isArray(items) ? items : []
      open()
    })
  })
  input.addEventListener('input', () => {
    hidden.value = ''
    open()
  })
  input.addEventListener('keydown', (event) => {
    if (list.classList.contains('hidden') && ['ArrowDown', 'ArrowUp', 'Enter'].includes(event.key)) open()
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      activeIndex = Math.min(activeIndex + 1, Math.max(filtered.length - 1, 0))
      render()
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      activeIndex = Math.max(activeIndex - 1, 0)
      render()
    }
    if (event.key === 'Enter' && !list.classList.contains('hidden')) {
      event.preventDefault()
      if (filtered[activeIndex]) selectTenant(filtered[activeIndex])
    }
    if (event.key === 'Escape') close()
  })
  list.addEventListener('click', (event) => {
    const option = event.target.closest('[data-tenant-option]')
    if (!option) return
    const tenant = filtered[Number(option.dataset.tenantOption)]
    if (tenant) selectTenant(tenant)
  })
  document.addEventListener('click', (event) => {
    if (!root.contains(event.target)) close()
  })
}

document.querySelectorAll('[data-tenant-combobox]').forEach(initTenantComboBox)

const roomOptionLabel = (room) => `${room.room_no || ''} - ${room.title || ''}`
const roomSearchText = (room) => `${room.room_no || ''} ${room.title || ''} ${room.status_label || ''} ${roomOptionLabel(room)}`.toLowerCase()

const roomOptionsURL = (status) => {
  const params = new URLSearchParams()
  if (!status || status === 'all') {
    params.set('include_all', 'true')
  } else {
    params.set('status', status)
  }
  return `/api/rooms?${params.toString()}`
}

const initRoomComboBox = (root) => {
  if (root.dataset.roomComboboxReady === 'true') return
  root.dataset.roomComboboxReady = 'true'
  const input = root.querySelector('[data-room-input]')
  const list = root.querySelector('[data-room-list]')
  const form = input?.closest('form')
  if (!input || !list) return

  let rooms = []
  let filtered = []
  let activeIndex = 0
  let loadedStatus = null
  let loading = false
  let loadFailed = false
  let suppressOpen = false

  const currentStatus = () => form?.querySelector('input[name="status"]')?.value || ''
  const close = () => list.classList.add('hidden')
  const message = (text, extraClass = 'text-stone-500') => {
    list.innerHTML = `<div class="px-4 py-3 text-sm ${extraClass}">${escapeHTML(text)}</div>`
  }
  const loadRooms = () => {
    const status = currentStatus()
    if (loadedStatus === status && !loading) return Promise.resolve(rooms)
    loadedStatus = status
    loading = true
    loadFailed = false
    message('正在加载房源...')
    return fetch(roomOptionsURL(status))
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('读取房源失败')))
      .then((payload) => {
        rooms = Array.isArray(payload.data) ? payload.data : []
        return rooms
      })
      .catch(() => {
        rooms = []
        loadFailed = true
        return rooms
      })
      .finally(() => {
        loading = false
      })
  }
  const updateFiltered = () => {
    const query = input.value.trim().toLowerCase()
    filtered = rooms.filter((room) => query === '' || roomSearchText(room).includes(query))
    activeIndex = Math.min(activeIndex, Math.max(filtered.length - 1, 0))
  }
  const render = () => {
    if (loading && rooms.length === 0) {
      message('正在加载房源...')
      return
    }
    if (loadFailed) {
      message('读取房源失败，请稍后重试。', 'text-red-700')
      return
    }
    updateFiltered()
    if (filtered.length === 0) {
      message('未找到匹配房源')
      return
    }
    list.innerHTML = filtered.map((room, index) => {
      const active = index === activeIndex ? 'bg-amber-50 text-amber-900' : 'hover:bg-stone-50'
      const status = room.status_label ? `<span class="mt-1 block text-xs font-bold text-stone-500">${escapeHTML(room.status_label)}</span>` : ''
      return `<button class="block w-full px-4 py-3 text-left text-sm font-semibold ${active}" type="button" data-room-option="${index}"><span>${escapeHTML(roomOptionLabel(room))}</span>${status}</button>`
    }).join('')
  }
  const open = () => {
    list.classList.remove('hidden')
    render()
  }
  const openWithRooms = () => {
    list.classList.remove('hidden')
    loadRooms().then(() => {
      if (!list.classList.contains('hidden')) render()
    })
  }
  const submitFilter = () => {
    if (!form) return
    if (typeof form.requestSubmit === 'function') {
      form.requestSubmit()
      return
    }
    const event = new Event('submit', { bubbles: true, cancelable: true })
    if (form.dispatchEvent(event)) form.submit()
  }
  const selectRoom = (room) => {
    input.value = roomOptionLabel(room)
    suppressOpen = true
    input.dispatchEvent(new Event('input', { bubbles: true }))
    close()
    submitFilter()
  }

  input.addEventListener('focus', openWithRooms)
  input.addEventListener('input', () => {
    if (suppressOpen) {
      suppressOpen = false
      return
    }
    activeIndex = 0
    if (loadedStatus === currentStatus() && !loading) {
      open()
      return
    }
    openWithRooms()
  })
  input.addEventListener('keydown', (event) => {
    if (list.classList.contains('hidden') && ['ArrowDown', 'ArrowUp'].includes(event.key)) openWithRooms()
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      activeIndex = Math.min(activeIndex + 1, Math.max(filtered.length - 1, 0))
      render()
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      activeIndex = Math.max(activeIndex - 1, 0)
      render()
    }
    if (event.key === 'Enter' && !list.classList.contains('hidden')) {
      event.preventDefault()
      if (filtered[activeIndex]) selectRoom(filtered[activeIndex])
    }
    if (event.key === 'Escape') close()
  })
  list.addEventListener('click', (event) => {
    const option = event.target.closest('[data-room-option]')
    if (!option) return
    const room = filtered[Number(option.dataset.roomOption)]
    if (room) selectRoom(room)
  })
  document.addEventListener('click', (event) => {
    if (!root.contains(event.target)) close()
  })
}

const initRoomComboBoxes = (scope = document) => {
  const roots = []
  if (scope?.matches?.('[data-room-combobox]')) roots.push(scope)
  if (scope?.querySelectorAll) roots.push(...scope.querySelectorAll('[data-room-combobox]'))
  roots.forEach(initRoomComboBox)
}

initRoomComboBoxes()

const tenantSearchOptionLabel = (tenant) => `${tenant.name || ''} - ${tenant.room_no || ''} - ${tenant.phone || ''}`
const tenantListSearchText = (tenant) => `${tenant.name || ''} ${tenant.room_no || ''} ${tenant.phone || ''} ${tenantSearchOptionLabel(tenant)}`.toLowerCase()

const tenantSearchOptionsURL = (status) => {
  const params = new URLSearchParams()
  params.set('status', status || 'active')
  return `/api/tenants?${params.toString()}`
}

const initTenantSearchComboBox = (root) => {
  if (root.dataset.tenantSearchComboboxReady === 'true') return
  root.dataset.tenantSearchComboboxReady = 'true'
  const input = root.querySelector('[data-tenant-search-input]')
  const list = root.querySelector('[data-tenant-search-list]')
  const form = input?.closest('form')
  if (!input || !list) return

  let tenants = []
  let filtered = []
  let activeIndex = 0
  let loadedStatus = null
  let loading = false
  let loadFailed = false
  let suppressOpen = false

  const currentStatus = () => {
    if (root.dataset.searchStatus) return root.dataset.searchStatus
    return form?.querySelector('input[name="status"]')?.value || 'active'
  }
  const close = () => list.classList.add('hidden')
  const message = (text, extraClass = 'text-stone-500') => {
    list.innerHTML = `<div class="px-4 py-3 text-sm ${extraClass}">${escapeHTML(text)}</div>`
  }
  const loadTenants = () => {
    const status = currentStatus()
    if (loadedStatus === status && !loading) return Promise.resolve(tenants)
    loadedStatus = status
    loading = true
    loadFailed = false
    message('正在加载租客...')
    return fetch(tenantSearchOptionsURL(status))
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('读取租客失败')))
      .then((payload) => {
        tenants = Array.isArray(payload.data) ? payload.data : []
        return tenants
      })
      .catch(() => {
        tenants = []
        loadFailed = true
        return tenants
      })
      .finally(() => {
        loading = false
      })
  }
  const updateFiltered = () => {
    const query = input.value.trim().toLowerCase()
    filtered = tenants.filter((tenant) => query === '' || tenantListSearchText(tenant).includes(query))
    activeIndex = Math.min(activeIndex, Math.max(filtered.length - 1, 0))
  }
  const render = () => {
    if (loading && tenants.length === 0) {
      message('正在加载租客...')
      return
    }
    if (loadFailed) {
      message('读取租客失败，请稍后重试。', 'text-red-700')
      return
    }
    updateFiltered()
    if (filtered.length === 0) {
      message('未找到匹配租客')
      return
    }
    list.innerHTML = filtered.map((tenant, index) => {
      const active = index === activeIndex ? 'bg-amber-50 text-amber-900' : 'hover:bg-stone-50'
      return `<button class="block w-full px-4 py-3 text-left text-sm font-semibold ${active}" type="button" data-tenant-search-option="${index}">${escapeHTML(tenantSearchOptionLabel(tenant))}</button>`
    }).join('')
  }
  const open = () => {
    list.classList.remove('hidden')
    render()
  }
  const openWithTenants = () => {
    list.classList.remove('hidden')
    loadTenants().then(() => {
      if (!list.classList.contains('hidden')) render()
    })
  }
  const submitFilter = () => {
    if (!form) return
    if (typeof form.requestSubmit === 'function') {
      form.requestSubmit()
      return
    }
    const event = new Event('submit', { bubbles: true, cancelable: true })
    if (form.dispatchEvent(event)) form.submit()
  }
  const selectTenant = (tenant) => {
    input.value = tenantSearchOptionLabel(tenant)
    suppressOpen = true
    input.dispatchEvent(new Event('input', { bubbles: true }))
    close()
    submitFilter()
  }

  input.addEventListener('focus', openWithTenants)
  input.addEventListener('input', () => {
    if (suppressOpen) {
      suppressOpen = false
      return
    }
    activeIndex = 0
    if (loadedStatus === currentStatus() && !loading) {
      open()
      return
    }
    openWithTenants()
  })
  input.addEventListener('keydown', (event) => {
    if (list.classList.contains('hidden') && ['ArrowDown', 'ArrowUp'].includes(event.key)) openWithTenants()
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      activeIndex = Math.min(activeIndex + 1, Math.max(filtered.length - 1, 0))
      render()
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault()
      activeIndex = Math.max(activeIndex - 1, 0)
      render()
    }
    if (event.key === 'Enter' && !list.classList.contains('hidden')) {
      event.preventDefault()
      if (filtered[activeIndex]) selectTenant(filtered[activeIndex])
    }
    if (event.key === 'Escape') close()
  })
  list.addEventListener('click', (event) => {
    const option = event.target.closest('[data-tenant-search-option]')
    if (!option) return
    const tenant = filtered[Number(option.dataset.tenantSearchOption)]
    if (tenant) selectTenant(tenant)
  })
  document.addEventListener('click', (event) => {
    if (!root.contains(event.target)) close()
  })
}

const initTenantSearchComboBoxes = (scope = document) => {
  const roots = []
  if (scope?.matches?.('[data-tenant-search-combobox]')) roots.push(scope)
  if (scope?.querySelectorAll) roots.push(...scope.querySelectorAll('[data-tenant-search-combobox]'))
  roots.forEach(initTenantSearchComboBox)
}

initTenantSearchComboBoxes()

const initPaymentsExcludeModal = () => {
  const modal = document.querySelector('[data-payment-exclude-modal]')
  if (!modal) return
  const form = modal.querySelector('[data-payment-exclude-form]')
  const cancel = modal.querySelector('[data-payment-exclude-cancel]')
  const open = (paymentID) => {
    if (!form || !paymentID) return
    const action = `/admin/payments/${paymentID}/exclusion`
    form.setAttribute('action', action)
    form.setAttribute('hx-post', action)
    if (window.htmx?.process) window.htmx.process(form)
    modal.classList.remove('hidden')
    modal.classList.add('flex')
  }
  const close = () => {
    modal.classList.add('hidden')
    modal.classList.remove('flex')
  }
  document.addEventListener('click', (event) => {
    const trigger = event.target.closest('[data-payment-exclude]')
    if (trigger) {
      event.preventDefault()
      open(trigger.dataset.paymentExclude)
      return
    }
    if (event.target === modal) close()
  })
  cancel?.addEventListener('click', close)
  form?.addEventListener('htmx:afterRequest', (event) => {
    if (event.detail?.successful !== false) close()
  })
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && !modal.classList.contains('hidden')) close()
  })
}

initPaymentsExcludeModal()

const initPaymentCreateForm = () => {
  const form = document.querySelector('[data-payment-create-form]')
  if (!form) return
  form.addEventListener('htmx:afterRequest', (event) => {
    if (event.detail?.successful === false) return
    if (event.detail?.xhr?.status >= 400) return
    form.reset()
    form.querySelectorAll('[data-tenant-hidden]').forEach((field) => { field.value = '' })
    form.querySelectorAll('[data-counter-target]').forEach((field) => field.dispatchEvent(new Event('input', { bubbles: true })))
  })
}

initPaymentCreateForm()

const initPaymentMobileFilters = () => {
  document.addEventListener('change', (event) => {
    const select = event.target.closest('[data-payment-mobile-filter]')
    if (!select) return
    const url = select.value
    if (!url) return
    if (window.htmx) {
      window.htmx.ajax('GET', url, {
        target: '#payment-list-section',
        select: '#payment-list-section',
        swap: 'outerHTML'
      })
      if (window.history?.pushState) window.history.pushState({}, '', url)
    } else {
      window.location.href = url
    }
  })
}

initPaymentMobileFilters()

const initPaymentCreateToggle = () => {
  const card = document.querySelector('[data-payment-create-card]')
  if (!card) return
  const toggle = card.querySelector('[data-payment-create-toggle]')
  const desktop = window.matchMedia('(min-width: 768px)')
  const apply = () => {
    if (desktop.matches) {
      card.open = true
    }
    if (toggle) toggle.textContent = card.open ? '收起' : '展开'
  }
  card.addEventListener('toggle', apply)
  desktop.addEventListener?.('change', apply)
  apply()
}

initPaymentCreateToggle()

const initPaymentsTable_DISABLED = () => {
  const root = document.querySelector('[data-payments-table]')
  if (!root) return
  const body = root.querySelector('[data-payments-body]')
  const grid = root.querySelector('[data-payments-grid]')
  const cards = root.querySelector('[data-payments-cards]')
  const footer = root.querySelector('[data-payments-footer]')
  const filterForm = root.querySelector('[data-payments-filter]')
  const summaryRoot = root.querySelector('[data-payments-summary]')
  const sortSelect = root.querySelector('[data-payment-sort-select]')
  const viewLinks = root.querySelectorAll('[data-payment-view-link]')
  const swipe = root.querySelector('[data-payment-swipe]')
  const viewPanes = root.querySelectorAll('[data-payment-view-pane]')
  const modal = document.querySelector('[data-payment-exclude-modal]')
  const modalCancel = modal?.querySelector('[data-payment-exclude-cancel]')
  const modalConfirm = modal?.querySelector('[data-payment-exclude-confirm]')
  const initialSort = paymentSortFromSelect(sortSelect)
  let viewMode = root.dataset.paymentView || 'list'
  let pendingExcludeID = ''
  let requestSeq = 0
  let swipeTimer = 0
  let state = { page: 1, limit: 20, hasMore: true, loading: false, sortBy: initialSort.sortBy, sortDir: initialSort.sortDir }
  if (grid) grid.className = paymentCardsContainerClass('grid')
  if (cards) cards.className = paymentCardsContainerClass('card')

  const setState = (updates) => {
    state = { ...state, ...updates }
  }
  const filters = () => {
    const data = new FormData(filterForm)
    const params = new URLSearchParams(window.location.search)
    return {
      q: data.get('q') || '',
      tenant_id: params.get('tenant_id') || '',
      paid: data.get('paid') || 'false',
      period: data.get('period') || '',
      type: data.get('type') || '',
      excluded: data.get('excluded') || 'false',
      overdue: params.get('overdue') || ''
    }
  }
  const updateIndicators = () => {
    root.querySelectorAll('[data-sort-indicator]').forEach((indicator) => {
      indicator.textContent = indicator.dataset.sortIndicator === state.sortBy ? (state.sortDir === 'asc' ? '↑' : '↓') : ''
    })
  }
  const paymentPageParams = (targetView = viewMode) => {
    const params = new URLSearchParams()
    Object.entries(filters()).forEach(([key, value]) => {
      if (value !== '') params.set(key, value)
    })
    params.set('sort_by', state.sortBy)
    params.set('sort_dir', state.sortDir)
    params.set('view', targetView)
    return params
  }
  const paymentPageURL = (targetView = viewMode) => {
    const query = paymentPageParams(targetView).toString()
    return query ? `/admin/payments?${query}` : '/admin/payments'
  }
  const syncPaymentNavigation = (replaceURL = false) => {
    viewLinks.forEach((link) => {
      const targetView = link.dataset.paymentViewValue || viewMode
      link.href = paymentPageURL(targetView)
      link.classList.toggle('public-room-view-active', targetView === viewMode)
      if (targetView === viewMode) {
        link.setAttribute('aria-current', 'page')
      } else {
        link.removeAttribute('aria-current')
      }
    })
    const viewInput = filterForm.querySelector('input[name="view"]')
    if (viewInput) viewInput.value = viewMode
    if (replaceURL && window.history?.replaceState) {
      window.history.replaceState(null, '', paymentPageURL(viewMode))
    }
  }
  const setPaymentView = (nextView, replaceURL = false) => {
    if (!nextView || nextView === viewMode) return
    viewMode = nextView
    root.dataset.paymentView = nextView
    syncPaymentNavigation(replaceURL)
  }
  const scrollToPaymentView = (nextView, behavior = 'auto') => {
    const pane = root.querySelector(`[data-payment-view-pane="${nextView}"]`)
    if (!pane || !swipe) return
    swipe.scrollTo({ left: pane.offsetLeft, behavior })
  }
  const paymentURL = () => {
    const params = new URLSearchParams({ page: String(state.page), limit: String(state.limit), sort_by: state.sortBy, sort_dir: state.sortDir })
    Object.entries(filters()).forEach(([key, value]) => {
      if (value !== '') params.set(key, value)
    })
    return `/api/payments?${params.toString()}`
  }
  const setFooter = (message) => {
    if (footer) footer.textContent = message
  }
  const showLoading = () => {
    body.innerHTML = '<tr><td class="px-5 py-10 text-center text-stone-500" colspan="6">正在加载收款记录...</td></tr>'
    if (grid) grid.innerHTML = paymentCardMessage('正在加载收款记录...')
    if (cards) cards.innerHTML = paymentCardMessage('正在加载收款记录...')
  }
  const showEmpty = () => {
    body.innerHTML = '<tr><td class="px-5 py-10 text-center text-stone-500" colspan="6">还没有收款记录。</td></tr>'
    if (grid) grid.innerHTML = paymentCardMessage('还没有收款记录。')
    if (cards) cards.innerHTML = paymentCardMessage('还没有收款记录。')
  }
  const showError = () => {
    body.innerHTML = '<tr><td class="px-5 py-10 text-center text-red-700" colspan="6">读取收款记录失败，请稍后重试。</td></tr>'
    if (grid) grid.innerHTML = paymentCardMessage('读取收款记录失败，请稍后重试。', 'text-red-700 border-red-200 bg-red-50')
    if (cards) cards.innerHTML = paymentCardMessage('读取收款记录失败，请稍后重试。', 'text-red-700 border-red-200 bg-red-50')
  }
  const reload = () => {
    requestSeq += 1
    setState({ page: 1, hasMore: true, loading: false })
    syncPaymentNavigation(true)
    showLoading()
    loadPage(true, requestSeq)
  }
  const loadPage = (replace = false, version = requestSeq) => {
    if (state.loading || !state.hasMore) return
    setState({ loading: true })
    setFooter('正在加载...')
    fetch(paymentURL())
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('读取收款记录失败')))
      .then((payload) => {
        if (version !== requestSeq) return
        const payments = payload.data || []
        const rows = payments.map(renderPaymentRow).join('')
        const gridItems = payments.map((payment) => renderPaymentCard(payment, 'grid')).join('')
        const cardItems = payments.map((payment) => renderPaymentCard(payment, 'card')).join('')
        if (payload.summary) renderPaymentSummary(summaryRoot, payload.summary)
        if (replace) {
          body.innerHTML = ''
          if (grid) grid.innerHTML = ''
          if (cards) cards.innerHTML = ''
        }
        if (rows) body.insertAdjacentHTML('beforeend', rows)
        if (grid && gridItems) grid.insertAdjacentHTML('beforeend', gridItems)
        if (cards && cardItems) cards.insertAdjacentHTML('beforeend', cardItems)
        if (replace && !payments.length) showEmpty()
        setState({ hasMore: Boolean(payload.has_more), page: state.page + 1 })
        setFooter(state.hasMore ? '下拉加载更多记录' : '已显示全部记录')
      })
      .catch(() => {
        if (version !== requestSeq) return
        if (replace) showError()
        setFooter('读取失败')
      })
      .finally(() => {
        if (version !== requestSeq) return
        setState({ loading: false })
        updateIndicators()
      })
  }
  const openExcludeModal = (paymentID) => {
    pendingExcludeID = paymentID
    modal?.classList.remove('hidden')
    modal?.classList.add('flex')
  }
  const closeExcludeModal = () => {
    pendingExcludeID = ''
    modal?.classList.add('hidden')
    modal?.classList.remove('flex')
  }

  filterForm.addEventListener('submit', (event) => {
    event.preventDefault()
    reload()
  })
  filterForm.addEventListener('change', (event) => {
    if (event.target.closest('[data-payment-sort-select]')) {
      const nextSort = paymentSortFromSelect(sortSelect)
      setState({ sortBy: nextSort.sortBy, sortDir: nextSort.sortDir })
    }
    reload()
  })
  root.querySelectorAll('[data-payment-sort]').forEach((button) => {
    button.addEventListener('click', () => {
      const nextSort = button.dataset.paymentSort
      const nextDir = state.sortBy === nextSort && state.sortDir === 'asc' ? 'desc' : 'asc'
      setState({ sortBy: nextSort, sortDir: nextDir })
      if (sortSelect) {
        const nextValue = `${nextSort}:${nextDir}`
        if (Array.from(sortSelect.options).some((option) => option.value === nextValue)) sortSelect.value = nextValue
      }
      reload()
    })
  })
  viewLinks.forEach((link) => {
    link.addEventListener('click', (event) => {
      event.preventDefault()
      const nextView = link.dataset.paymentViewValue || 'list'
      setPaymentView(nextView, true)
      scrollToPaymentView(nextView, 'smooth')
    })
  })
  swipe?.addEventListener('scroll', () => {
    window.clearTimeout(swipeTimer)
    swipeTimer = window.setTimeout(() => {
      const nearestPane = Array.from(viewPanes).reduce((nearest, pane) => {
        if (!nearest) return pane
        const currentDistance = Math.abs(pane.offsetLeft - swipe.scrollLeft)
        const nearestDistance = Math.abs(nearest.offsetLeft - swipe.scrollLeft)
        return currentDistance < nearestDistance ? pane : nearest
      }, null)
      setPaymentView(nearestPane?.dataset.paymentViewPane, true)
    }, 120)
  })
  root.addEventListener('click', (event) => {
    const trigger = event.target.closest('[data-payment-exclude]')
    if (!trigger) return
    openExcludeModal(trigger.dataset.paymentExclude)
  })
  modalCancel?.addEventListener('click', closeExcludeModal)
  modal?.addEventListener('click', (event) => {
    if (event.target === modal) closeExcludeModal()
  })
  modalConfirm?.addEventListener('click', () => {
    if (!pendingExcludeID) return
    modalConfirm.disabled = true
    fetch(`/admin/payments/${pendingExcludeID}/exclude`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ excluded: true })
    })
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('操作失败')))
      .then(() => {
        closeExcludeModal()
        reload()
      })
      .catch(() => setFooter('操作失败，请稍后重试'))
      .finally(() => {
        modalConfirm.disabled = false
      })
  })

  if ('IntersectionObserver' in window && footer) {
    const observer = new IntersectionObserver((entries) => {
      if (entries.some((entry) => entry.isIntersecting)) loadPage()
    }, { rootMargin: '200px' })
    observer.observe(footer)
  } else if (footer) {
    footer.textContent = '点击加载更多'
    footer.addEventListener('click', () => loadPage())
  }

  syncPaymentNavigation(false)
  window.requestAnimationFrame(() => scrollToPaymentView(viewMode))
  reload()
}

const paymentSortFromSelect = (select) => {
  const [sortBy, sortDir] = String(select?.value || 'next_pay_date:asc').split(':')
  return { sortBy: sortBy || 'next_pay_date', sortDir: sortDir || 'asc' }
}

const renderPaymentSummary = (root, summary) => {
  if (!root) return
  root.innerHTML = [
    paymentSummaryCard('总待付款', `¥${summary.total_unpaid_amount_text || '0.00'}`),
    paymentSummaryCard('已付款', `¥${summary.total_paid_amount_text || '0.00'}`),
    paymentSummaryCard('已退租待处理', summary.checkout_pending_count || 0),
    paymentSummaryCard('排除记录', summary.excluded_count || 0)
  ].join('')
}

const paymentSummaryCard = (label, value) => `
  <div class="rounded-2xl border border-stone-200 bg-white/55 p-4">
    <p class="text-sm font-bold text-stone-500">${escapeHTML(label)}</p>
    <p class="mt-2 text-2xl font-black">${escapeHTML(value)}</p>
  </div>`

const paymentID = (payment) => payment.payment_id || payment.id
const paymentRoom = (payment) => payment.room || payment.room_no || '-'
const paymentStatusValue = (payment) => payment.payment_status || (payment.paid ? 'paid' : 'unpaid')
const paymentCardsContainerClass = (viewMode) => {
  if (viewMode === 'grid') return 'grid gap-5 p-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
  if (viewMode === 'card') return 'public-room-masonry p-4'
  return 'space-y-4 p-4 md:hidden'
}
const paymentCardMessage = (message, extraClass = 'text-stone-500') => `<div class="rounded-2xl border border-stone-200 bg-white/60 p-4 text-center ${extraClass}">${escapeHTML(message)}</div>`
const paymentReturnTo = () => `${window.location.pathname}${window.location.search}`
const paymentStatusLabel = (payment) => paymentStatusValue(payment) === 'paid' ? '已付款' : '未付款'
const tenantStatusLabel = (status) => {
  if (status === 'active') return '在租'
  if (status === 'checkout') return '已退租'
  return '未知'
}
const paymentMonth = (payment) => {
  const value = payment.pay_date || payment.next_pay_date_label || ''
  return value && value.length >= 7 ? value.slice(0, 7) : '-'
}

const renderPaymentRow = (payment) => {
  const rowClass = payment.excluded ? 'bg-stone-100/80 text-stone-500' : ''
  const statusClass = paymentStatusValue(payment) === 'paid' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
  const note = payment.note || '暂无备注'
  const badges = renderPaymentBadges(payment)
  return `
    <tr class="${rowClass}">
      <td class="payment-sticky-left col-tenant px-5 py-4">
        <a class="payment-tenant-link font-semibold" href="/admin/tenants/${escapeHTML(payment.tenant_id)}">${escapeHTML(payment.tenant_name)}</a>
        <p class="mt-1 text-xs text-stone-500">${escapeHTML(paymentRoom(payment))} · ${escapeHTML(tenantStatusLabel(payment.tenant_status))}</p>
      </td>
      <td class="col-month px-5 py-4">${escapeHTML(paymentMonth(payment))}</td>
      <td class="col-amount px-5 py-4 font-black">¥${escapeHTML(payment.amount_text)}</td>
      <td class="col-status px-5 py-4"><span class="badge ${statusClass}">${escapeHTML(paymentStatusLabel(payment))}</span>${badges}</td>
      <td class="col-note px-5 py-4"><p class="font-bold text-stone-700">${escapeHTML(payment.type_label)}</p><p class="mt-1 text-stone-600">${escapeHTML(note)}</p></td>
      <td class="payment-sticky-right col-actions space-y-2 px-5 py-4">${renderPaymentActions(payment)}</td>
    </tr>`
}

const renderPaymentCard = (payment, viewMode) => {
  const statusClass = paymentStatusValue(payment) === 'paid' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
  const note = payment.note || '暂无备注'
  const badges = renderPaymentBadges(payment)
  const mutedClass = payment.excluded ? 'opacity-75' : ''
  const shellClass = viewMode === 'card' ? 'card public-room-masonry-card overflow-hidden p-4' : 'card flex h-full flex-col overflow-hidden p-4'
  return `
    <article class="${shellClass} ${mutedClass}">
      <div class="flex flex-wrap items-center gap-2">
        <span class="badge ${statusClass}">${escapeHTML(paymentStatusLabel(payment))}</span>
        <span class="rounded-full bg-stone-900/5 px-3 py-1 text-xs font-black text-stone-600">${escapeHTML(payment.type_label)}</span>
        ${badges}
      </div>
      <div class="mt-4">
        <a class="payment-tenant-link text-xl font-black" href="/admin/tenants/${escapeHTML(payment.tenant_id)}">${escapeHTML(payment.tenant_name)}</a>
        <p class="mt-1 text-sm font-bold text-stone-500">${escapeHTML(paymentRoom(payment))} · ${escapeHTML(tenantStatusLabel(payment.tenant_status))}</p>
      </div>
      <div class="mt-5 flex flex-wrap gap-2 text-sm font-bold text-stone-600">
        <span>${escapeHTML(paymentMonth(payment))}</span>
        <span>${escapeHTML(payment.payment_terms_label || '')}</span>
        <span>${escapeHTML(payment.rent_type_label || '')}</span>
      </div>
      <p class="mt-5 text-3xl font-black text-amber-800">¥${escapeHTML(payment.amount_text)}</p>
      <div class="mt-4 border-t border-stone-200 pt-4">
        <p class="text-sm font-bold text-stone-500">备注</p>
        <p class="mt-1 text-sm text-stone-700">${escapeHTML(note)}</p>
      </div>
      <div class="mt-auto space-y-2 pt-5">${renderPaymentActions(payment)}</div>
    </article>`
}

const renderPaymentBadges = (payment) => {
  const excludedBadge = payment.excluded ? '<span class="badge ml-2 bg-stone-200 text-stone-600">已排除</span>' : ''
  const autoBadge = payment.auto_generated ? '<span class="badge ml-2 bg-amber-100 text-amber-800">自动生成</span>' : ''
  const overdueBadge = payment.overdue_label ? `<span class="badge ml-2 bg-red-100 text-red-700">${escapeHTML(payment.overdue_label)}</span>` : ''
  return `${excludedBadge}${autoBadge}${overdueBadge}`
}

const renderPaymentActions = (payment) => {
  const id = paymentID(payment)
  const status = paymentStatusValue(payment)
  const canRemind = status === 'unpaid' && !payment.excluded
  const canExclude = payment.tenant_status === 'checkout' && !payment.excluded
  const actionClass = status === 'paid' ? 'bg-emerald-600 hover:bg-emerald-700' : 'bg-red-600 hover:bg-red-700'
  const returnTo = escapeHTML(paymentReturnTo())
  const remindButton = canRemind ? '<button class="btn-secondary payment-action-button text-xs opacity-70" type="button" disabled>催缴</button>' : ''
  const excludeButton = canExclude ? `<button class="btn-secondary payment-action-button text-xs" type="button" data-payment-exclude="${escapeHTML(id)}">不再记录</button>` : ''
  const restoreForm = payment.excluded ? `
    <form method="post" action="/admin/payments/${escapeHTML(id)}/exclusion">
      <input type="hidden" name="excluded" value="false">
      <input type="hidden" name="return_to" value="${returnTo}">
      <button class="btn-secondary payment-action-button text-xs" type="submit">恢复记录</button>
    </form>` : ''
  const toggleForm = `
    <form method="post" action="/admin/payments/${escapeHTML(id)}/toggle">
      <input type="hidden" name="return_to" value="${returnTo}">
      <button class="payment-action-button rounded-full px-4 py-2 text-xs font-black text-white ${actionClass}" type="submit">${status === 'paid' ? '已付款，改为未付款' : '未付款，标记已付款'}</button>
    </form>`
  return `${remindButton}${excludeButton}${restoreForm}${toggleForm}`
}

// initPaymentsTable() — legacy client-side renderer, kept for reference but no
// longer invoked. The payments page now renders server-side via HTMX swaps.

const dashboardDetailConfig = {
  'rooms-total': { title: '总房源 — 明细', url: '/api/rooms?include_all=true', type: 'rooms', fullURL: '/admin/rooms' },
  'rooms-vacant': { title: '空置房源 — 明细', url: '/api/rooms?status=vacant', type: 'vacantRooms', fullURL: '/admin/rooms?status=vacant' },
  'rooms-occupied': { title: '已出租 — 明细', url: '/api/rooms?status=occupied', type: 'occupiedRooms', fullURL: '/admin/rooms?status=occupied' },
  'tenants-active': { title: '在租租客 — 明细', url: '/api/tenants?status=active', type: 'tenants', fullURL: '/admin/tenants?status=active' },
  'tenants-expired': { title: '尚未办理退租的租客 — 明细', url: '/api/tenants?status=active&lease_expired=true', type: 'expiredTenants', fullURL: '/admin/tenants?status=active' },
  'income-month': { title: '本月已收 — 明细', url: '/api/payments?paid=true&period=current_month&sort_by=next_pay_date&sort_dir=desc&limit=100', type: 'payments', fullURL: '/admin/payments?paid=true&period=month' },
  'payments-overdue': { title: '未收金额 — 明细', url: '/api/payments?paid=false&excluded=false&overdue=true&sort_by=next_pay_date&sort_dir=asc&limit=100', type: 'overduePayments', fullURL: '/admin/payments?paid=false&excluded=false&overdue=true' },
  'projection-month': { title: '本月应收 — 明细', url: '/api/stats/projection?period=month', type: 'projection', fullURL: '/admin/payments?period=month' },
  'projection-6months': { title: '半年应收 — 明细', url: '/api/stats/projection?period=6months', type: 'projectionMonths', fullURL: '/admin/stats' },
  'projection-12months': { title: '全年应收 — 明细', url: '/api/stats/projection?period=12months', type: 'projectionMonths', fullURL: '/admin/stats' }
}

const initDashboardDetails = () => {
  const panel = document.querySelector('[data-dashboard-panel]')
  if (!panel) return
  const title = panel.querySelector('[data-dashboard-panel-title]')
  const body = panel.querySelector('[data-dashboard-panel-body]')
  const lockPageScroll = () => {
    if (document.body.dataset.dashboardScrollY !== undefined) return
    const scrollY = window.scrollY
    document.body.dataset.dashboardScrollY = String(scrollY)
    document.documentElement.style.overflow = 'hidden'
    document.body.style.position = 'fixed'
    document.body.style.top = `-${scrollY}px`
    document.body.style.left = '0'
    document.body.style.right = '0'
    document.body.style.width = '100%'
    document.body.style.overflow = 'hidden'
  }
  const unlockPageScroll = () => {
    const scrollY = Number(document.body.dataset.dashboardScrollY || '0')
    delete document.body.dataset.dashboardScrollY
    document.documentElement.style.overflow = ''
    document.body.style.position = ''
    document.body.style.top = ''
    document.body.style.left = ''
    document.body.style.right = ''
    document.body.style.width = ''
    document.body.style.overflow = ''
    window.scrollTo(0, scrollY)
  }
  const openPanel = (label) => {
    title.textContent = label
    body.innerHTML = '<p class="text-sm text-stone-500">正在加载明细...</p>'
    lockPageScroll()
    panel.classList.remove('hidden')
  }
  const closePanel = () => {
    if (panel.classList.contains('hidden')) return
    panel.classList.add('hidden')
    unlockPageScroll()
  }
  panel.querySelectorAll('[data-dashboard-close]').forEach((button) => button.addEventListener('click', closePanel))
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') closePanel()
  })
  document.querySelectorAll('[data-dashboard-detail]').forEach((button) => {
    button.addEventListener('click', () => {
      const config = dashboardDetailConfig[button.dataset.dashboardDetail]
      if (!config) return
      openPanel(config.title)
      fetch(config.url)
        .then((response) => response.ok ? response.json() : Promise.reject(new Error('读取明细失败')))
        .then((payload) => {
          body.innerHTML = renderDashboardDetail(config, payload)
          initDashboardSortableTables(body)
        })
        .catch(() => {
          body.innerHTML = '<p class="text-sm font-bold text-red-700">读取明细失败，请稍后重试。</p>'
        })
    })
  })
}

const renderDashboardDetail = (config, payload) => {
  if (config.type === 'rooms') return renderRoomsDetail(payload.data || [], config.fullURL)
  if (config.type === 'vacantRooms') return renderVacantRoomsDetail(payload.data || [], config.fullURL)
  if (config.type === 'occupiedRooms') return renderOccupiedRoomsDetail(payload.data || [], config.fullURL)
  if (config.type === 'tenants') return renderTenantsDetail(payload.data || [], config.fullURL)
  if (config.type === 'expiredTenants') return renderExpiredTenantsDetail(payload.data || [], config.fullURL)
  if (config.type === 'payments') return renderPaymentsDetail(payload.data || [], config.fullURL)
  if (config.type === 'overduePayments') return renderOverduePaymentsDetail(payload.data || [], config.fullURL)
  if (config.type === 'projection') return renderProjectionDetail(payload, config.fullURL)
  if (config.type === 'projectionMonths') return renderProjectionMonthsDetail(payload, config.fullURL)
  return '<p class="text-sm text-stone-500">暂无明细。</p>'
}

const dashboardTable = (headers, rows, footer = '', options = {}) => {
  if (!rows.length) return `${footer}<p class="rounded-2xl bg-white/60 p-6 text-center text-sm text-stone-500">暂无记录。</p>`
  const defaultSortIndex = Number.isInteger(options.defaultSortIndex) ? options.defaultSortIndex : 0
  const defaultSortDir = options.defaultSortDir === 'desc' ? 'desc' : 'asc'
  const headerCells = headers.map((header, index) => {
    const isDefault = index === defaultSortIndex
    const indicator = isDefault ? (defaultSortDir === 'asc' ? '↑' : '↓') : ''
    return `<th class="border-b border-stone-200 px-3 py-3" aria-sort="${isDefault ? (defaultSortDir === 'asc' ? 'ascending' : 'descending') : 'none'}"><button class="inline-flex items-center gap-1 font-black text-stone-600 transition hover:text-stone-950" type="button" data-dashboard-sort-column="${index}" data-dashboard-sort-dir="${isDefault ? defaultSortDir : 'none'}" aria-label="按${escapeHTML(header)}排序"><span>${escapeHTML(header)}</span><span data-dashboard-sort-indicator>${indicator}</span></button></th>`
  }).join('')
  return `${footer}<div class="overflow-x-auto"><table class="w-full min-w-[620px] text-left text-sm" data-dashboard-sortable data-dashboard-default-sort-index="${defaultSortIndex}" data-dashboard-default-sort-dir="${defaultSortDir}"><thead class="text-stone-500"><tr>${headerCells}</tr></thead><tbody class="divide-y divide-stone-100">${rows.join('')}</tbody></table></div>`
}

const initDashboardSortableTables = (root) => {
  root.querySelectorAll('[data-dashboard-sortable]').forEach((table) => {
    table.querySelectorAll('[data-dashboard-sort-column]').forEach((button) => {
      button.addEventListener('click', () => {
        const currentDir = button.dataset.dashboardSortDir
        const nextDir = currentDir === 'asc' ? 'desc' : 'asc'
        sortDashboardTable(table, Number(button.dataset.dashboardSortColumn), nextDir)
      })
    })
    sortDashboardTable(table, Number(table.dataset.dashboardDefaultSortIndex), table.dataset.dashboardDefaultSortDir)
  })
}

const sortDashboardTable = (table, columnIndex, direction) => {
  const tbody = table.querySelector('tbody')
  if (!tbody) return
  const rows = Array.from(tbody.querySelectorAll('tr'))
  const sortedRows = rows
    .map((row, originalIndex) => ({ row, originalIndex, value: dashboardCellSortValue(row.children[columnIndex]) }))
    .sort((left, right) => {
      const result = compareDashboardSortValues(left.value, right.value)
      return result === 0 ? left.originalIndex - right.originalIndex : (direction === 'desc' ? -result : result)
    })
    .map((item) => item.row)
  sortedRows.forEach((row) => tbody.appendChild(row))
  updateDashboardSortHeaders(table, columnIndex, direction)
}

const dashboardCellSortValue = (cell) => (cell?.dataset.sortValue || cell?.textContent || '').trim()

const compareDashboardSortValues = (left, right) => {
  const leftDate = dashboardDateSortValue(left)
  const rightDate = dashboardDateSortValue(right)
  if (leftDate !== null && rightDate !== null) return leftDate - rightDate
  const leftNumber = dashboardNumberSortValue(left)
  const rightNumber = dashboardNumberSortValue(right)
  if (leftNumber !== null && rightNumber !== null) return leftNumber - rightNumber
  return String(left).localeCompare(String(right), 'zh-Hans-CN', { numeric: true, sensitivity: 'base' })
}

const dashboardDateSortValue = (value) => {
  const text = String(value).trim()
  if (!/^\d{4}[/-]\d{2}[/-]\d{2}/.test(text)) return null
  const time = Date.parse(text.replaceAll('/', '-'))
  return Number.isNaN(time) ? null : time
}

const dashboardNumberSortValue = (value) => {
  const match = String(value).replaceAll(',', '').trim().match(/^-?¥?\s*(\d+(?:\.\d+)?)/)
  return match ? Number(match[1]) : null
}

const updateDashboardSortHeaders = (table, columnIndex, direction) => {
  table.querySelectorAll('th').forEach((header, index) => {
    const button = header.querySelector('[data-dashboard-sort-column]')
    const indicator = header.querySelector('[data-dashboard-sort-indicator]')
    const active = index === columnIndex
    header.setAttribute('aria-sort', active ? (direction === 'asc' ? 'ascending' : 'descending') : 'none')
    if (button) button.dataset.dashboardSortDir = active ? direction : 'none'
    if (indicator) indicator.textContent = active ? (direction === 'asc' ? '↑' : '↓') : ''
  })
}

const dashboardRoomLink = (room) => {
  const roomNo = escapeHTML(room.room_no)
  if (!room.detail_url) return `<span class="font-black">${roomNo}</span>`
  return `<a class="font-black text-amber-800" href="${escapeHTML(room.detail_url)}">${roomNo}</a>`
}

const renderRoomsDetail = (rooms, fullURL) => dashboardTable(
  ['房间号', '标题', '状态', '楼层', '面积'],
  rooms.map((room) => `<tr><td class="px-3 py-3">${dashboardRoomLink(room)}</td><td class="px-3 py-3">${escapeHTML(room.title)}</td><td class="px-3 py-3">${escapeHTML(room.status_label)}</td><td class="px-3 py-3">${escapeHTML(room.floor)}层</td><td class="px-3 py-3">${escapeHTML(room.area)}㎡</td></tr>`),
  `<p class="mb-4 text-sm font-bold text-stone-600">共 ${rooms.length} 间</p>`
) + fullListLink(fullURL)

const renderVacantRoomsDetail = (rooms, fullURL) => {
  const totalLoss = rooms.reduce((sum, room) => sum + Number(room.rent_price_fen || 0), 0)
  const summary = `<p class="mb-4 text-sm font-bold text-stone-600">共空置 ${rooms.length} 间，每月潜在损失：¥${formatYuanIntText(totalLoss)}</p>`
  return dashboardTable(
    ['房间号', '标题', '租金', '楼层', '面积'],
    rooms.map((room) => `<tr><td class="px-3 py-3">${dashboardRoomLink(room)}</td><td class="px-3 py-3">${escapeHTML(room.title)}</td><td class="px-3 py-3">¥${escapeHTML(room.rent_price_text)} / ${escapeHTML(room.rent_type_label)}</td><td class="px-3 py-3">${escapeHTML(room.floor)}层</td><td class="px-3 py-3">${escapeHTML(room.area)}㎡</td></tr>`),
    summary
  ) + fullListLink(fullURL)
}

const renderOccupiedRoomsDetail = (rooms, fullURL) => dashboardTable(
  ['房间号', '当前租客', '租约到期日', '租金'],
  rooms.map((room) => `<tr><td class="px-3 py-3">${dashboardRoomLink(room)}</td><td class="px-3 py-3">${escapeHTML(room.tenant_name || '-')}</td><td class="px-3 py-3">${escapeHTML(room.lease_end_date ? formatDisplayDate(room.lease_end_date) : '长期')}</td><td class="px-3 py-3">¥${escapeHTML(room.rent_price_text)} / ${escapeHTML(room.rent_type_label)}</td></tr>`)
) + fullListLink(fullURL)

const renderTenantsDetail = (tenants, fullURL) => dashboardTable(
  ['姓名', '手机号', '房间号', '入住日期', '租金'],
  tenants.map((tenant) => `<tr><td class="px-3 py-3"><a class="font-black text-amber-800" href="${escapeHTML(tenant.detail_url)}">${escapeHTML(tenant.name)}</a></td><td class="px-3 py-3">${escapeHTML(tenant.phone)}</td><td class="px-3 py-3">${escapeHTML(tenant.room_no)}</td><td class="px-3 py-3">${escapeHTML(formatDisplayDate(tenant.checkin_date))}</td><td class="px-3 py-3">¥${escapeHTML(tenant.rent_price_text)} / ${escapeHTML(tenant.rent_type_label)}</td></tr>`),
  '',
  { defaultSortIndex: 3, defaultSortDir: 'desc' }
) + fullListLink(fullURL)

const renderExpiredTenantsDetail = (tenants, fullURL) => dashboardTable(
  ['姓名', '房间号', '原租约到期日', '已超期天数', '操作'],
  tenants.map((tenant) => `<tr><td class="px-3 py-3"><a class="font-black text-amber-800" href="${escapeHTML(tenant.detail_url)}">${escapeHTML(tenant.name)}</a></td><td class="px-3 py-3">${escapeHTML(tenant.room_no)}</td><td class="px-3 py-3">${escapeHTML(formatDisplayDate(tenant.lease_end_date))}</td><td class="px-3 py-3 font-black text-red-700">${escapeHTML(tenant.overdue_days)} 天</td><td class="px-3 py-3"><a class="btn-secondary text-xs" href="${escapeHTML(tenant.checkout_url)}">立即办理退租</a></td></tr>`)
) + fullListLink(fullURL)

const renderPaymentsDetail = (payments, fullURL) => {
  const total = payments.reduce((sum, payment) => sum + Number(payment.amount_fen || 0), 0)
  const summary = `<p class="mb-4 text-sm font-bold text-stone-600">共收款 ${payments.length} 笔，合计 ¥${formatFenText(total)}</p>`
  return dashboardTable(
    ['租客姓名', '房间号', '类型', '金额', '收款日期'],
    payments.map((payment) => `<tr><td class="px-3 py-3 font-black">${escapeHTML(payment.tenant_name)}</td><td class="px-3 py-3">${escapeHTML(payment.room_no)}</td><td class="px-3 py-3">${escapeHTML(payment.type_label)}</td><td class="px-3 py-3">¥${escapeHTML(payment.amount_text)}</td><td class="px-3 py-3">${escapeHTML(formatDisplayDate(payment.pay_date))}</td></tr>`),
    summary
  ) + fullListLink(fullURL)
}

const renderOverduePaymentsDetail = (payments, fullURL) => {
  const total = payments.reduce((sum, payment) => sum + Number(payment.amount_fen || 0), 0)
  const summary = `<p class="mb-4 text-sm font-bold text-stone-600">共 ${payments.length} 笔欠款，合计 ¥${formatFenText(total)}</p>`
  return dashboardTable(
    ['租客姓名', '房间号', '类型', '欠款金额', '应付日期', '逾期天数', '操作'],
    payments.map((payment) => `<tr><td class="px-3 py-3 font-black">${escapeHTML(payment.tenant_name)}</td><td class="px-3 py-3">${escapeHTML(payment.room_no)}</td><td class="px-3 py-3">${escapeHTML(payment.type_label)}</td><td class="px-3 py-3">¥${escapeHTML(payment.amount_text)}</td><td class="px-3 py-3 text-red-700">${escapeHTML(formatDisplayDate(payment.pay_date))}</td><td class="px-3 py-3 font-black text-red-700">${escapeHTML(payment.overdue_days)} 天</td><td class="px-3 py-3"><form method="post" action="/admin/payments/${payment.id}/toggle"><button class="btn-secondary text-xs" type="submit">标记已收</button></form></td></tr>`),
    summary
  ) + fullListLink(fullURL)
}

const renderProjectionDetail = (projection, fullURL) => {
  const summary = `<div class="mb-4 rounded-2xl bg-white/60 p-4 text-sm font-bold text-stone-700">已收：¥${formatFenText(projection.collected_fen)}｜未收：¥${formatFenText(projection.unpaid_fen)}｜未到期：¥${formatFenText(projection.not_due_fen)}</div>`
  return dashboardTable(
    ['租客姓名', '房间号', '应收金额', '支付状态', '应付日期'],
    (projection.items || []).map((item) => `<tr><td class="px-3 py-3 font-black">${escapeHTML(item.tenant_name)}</td><td class="px-3 py-3">${escapeHTML(item.room_no)}</td><td class="px-3 py-3">¥${escapeHTML(item.amount_text)}</td><td class="px-3 py-3">${escapeHTML(item.status)}</td><td class="px-3 py-3">${escapeHTML(formatDisplayDate(item.due_date))}</td></tr>`),
    summary
  ) + fullListLink(fullURL)
}

const renderProjectionMonthsDetail = (projection, fullURL) => {
  const summary = `<p class="mb-4 text-sm font-bold text-stone-600">合计应收 ¥${escapeHTML(projection.total_text)}，仅含当前在租租客，不含未来新签租约。</p>`
  return dashboardTable(
    ['月份', '应收总额', '参与租客数'],
    (projection.months || []).map((month) => `<tr><td class="px-3 py-3 font-black">${escapeHTML(month.month)}</td><td class="px-3 py-3">¥${escapeHTML(month.total_text)}</td><td class="px-3 py-3">${escapeHTML(month.tenant_count)}</td></tr>`),
    summary
  ) + fullListLink(fullURL)
}

const fullListLink = (url) => `<div class="mt-5"><a class="btn-primary inline-flex" href="${escapeHTML(url)}">查看完整列表</a></div>`

const formatFenText = (fen) => (Number(fen || 0) / 100).toFixed(2)
const formatYuanIntText = (fen) => String(Math.trunc(Number(fen || 0) / 100))

initDashboardDetails()

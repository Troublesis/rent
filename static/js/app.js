document.addEventListener('click', (event) => {
  const target = event.target.closest('[data-confirm]')
  if (!target) return
  if (!window.confirm(target.dataset.confirm)) {
    event.preventDefault()
  }
})

document.addEventListener('click', (event) => {
  const target = event.target.closest('[data-toggle-panel]')
  if (!target) return
  const panel = document.getElementById(target.dataset.togglePanel)
  if (panel) panel.classList.toggle('hidden')
})

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

const initPaymentsTable = () => {
  const root = document.querySelector('[data-payments-table]')
  if (!root) return
  const body = root.querySelector('[data-payments-body]')
  const cards = root.querySelector('[data-payments-cards]')
  const footer = root.querySelector('[data-payments-footer]')
  const filterForm = root.querySelector('[data-payments-filter]')
  const summaryRoot = root.querySelector('[data-payments-summary]')
  const sortSelect = root.querySelector('[data-payment-sort-select]')
  const modal = document.querySelector('[data-payment-exclude-modal]')
  const modalCancel = modal?.querySelector('[data-payment-exclude-cancel]')
  const modalConfirm = modal?.querySelector('[data-payment-exclude-confirm]')
  const initialSort = paymentSortFromSelect(sortSelect)
  let pendingExcludeID = ''
  let state = { page: 1, limit: 20, hasMore: true, loading: false, sortBy: initialSort.sortBy, sortDir: initialSort.sortDir }

  const setState = (updates) => {
    state = { ...state, ...updates }
  }
  const filters = () => {
    const data = new FormData(filterForm)
    const params = new URLSearchParams(window.location.search)
    return {
      q: data.get('q') || '',
      tenant_id: params.get('tenant_id') || '',
      paid: data.get('paid') || '',
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
    if (cards) cards.innerHTML = '<div class="rounded-2xl border border-stone-200 bg-white/60 p-4 text-center text-stone-500">正在加载收款记录...</div>'
  }
  const showEmpty = () => {
    body.innerHTML = '<tr><td class="px-5 py-10 text-center text-stone-500" colspan="6">还没有收款记录。</td></tr>'
    if (cards) cards.innerHTML = '<div class="rounded-2xl border border-stone-200 bg-white/60 p-4 text-center text-stone-500">还没有收款记录。</div>'
  }
  const showError = () => {
    body.innerHTML = '<tr><td class="px-5 py-10 text-center text-red-700" colspan="6">读取收款记录失败，请稍后重试。</td></tr>'
    if (cards) cards.innerHTML = '<div class="rounded-2xl border border-red-200 bg-red-50 p-4 text-center text-red-700">读取收款记录失败，请稍后重试。</div>'
  }
  const reload = () => {
    setState({ page: 1, hasMore: true })
    showLoading()
    loadPage(true)
  }
  const loadPage = (replace = false) => {
    if (state.loading || !state.hasMore) return
    setState({ loading: true })
    setFooter('正在加载...')
    fetch(paymentURL())
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('读取收款记录失败')))
      .then((payload) => {
        const payments = payload.data || []
        const rows = payments.map(renderPaymentRow).join('')
        const mobileCards = payments.map(renderPaymentCard).join('')
        if (payload.summary) renderPaymentSummary(summaryRoot, payload.summary)
        if (replace) {
          body.innerHTML = ''
          if (cards) cards.innerHTML = ''
        }
        if (rows) body.insertAdjacentHTML('beforeend', rows)
        if (cards && mobileCards) cards.insertAdjacentHTML('beforeend', mobileCards)
        if (replace && !payments.length) showEmpty()
        setState({ hasMore: Boolean(payload.has_more), page: state.page + 1 })
        setFooter(state.hasMore ? '下拉加载更多记录' : '已显示全部记录')
      })
      .catch(() => {
        if (replace) showError()
        setFooter('读取失败')
      })
      .finally(() => {
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

const renderPaymentCard = (payment) => {
  const statusClass = paymentStatusValue(payment) === 'paid' ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'
  const note = payment.note || '暂无备注'
  const cardClass = payment.excluded ? 'opacity-75' : ''
  return `
    <article class="payment-mobile-card ${cardClass}">
      <div>
        <a class="payment-tenant-link font-semibold" href="/admin/tenants/${escapeHTML(payment.tenant_id)}">${escapeHTML(payment.tenant_name)}</a>
        <p class="mt-1 text-xs text-stone-500">${escapeHTML(paymentRoom(payment))} · ${escapeHTML(tenantStatusLabel(payment.tenant_status))}</p>
      </div>
      <dl class="mt-4 space-y-3 text-sm">
        <div class="flex justify-between gap-4"><dt class="text-stone-500">租金月份</dt><dd class="font-bold">${escapeHTML(paymentMonth(payment))}</dd></div>
        <div class="flex justify-between gap-4"><dt class="text-stone-500">金额</dt><dd class="font-bold">¥${escapeHTML(payment.amount_text)}</dd></div>
        <div class="flex justify-between gap-4"><dt class="text-stone-500">状态</dt><dd><span class="badge ${statusClass}">${escapeHTML(paymentStatusLabel(payment))}</span></dd></div>
      </dl>
      <div class="mt-4 border-t border-stone-200 pt-4">
        <p class="text-sm font-bold text-stone-500">备注</p>
        <p class="mt-1 text-sm text-stone-700">${escapeHTML(note)}</p>
      </div>
      <div class="mt-4 space-y-2">${renderPaymentActions(payment)}</div>
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
  const remindButton = canRemind ? '<button class="btn-secondary payment-action-button text-xs opacity-70" type="button" disabled>催缴</button>' : ''
  const excludeButton = canExclude ? `<button class="btn-secondary payment-action-button text-xs" type="button" data-payment-exclude="${escapeHTML(id)}">不再记录</button>` : ''
  const restoreForm = payment.excluded ? `
    <form method="post" action="/admin/payments/${escapeHTML(id)}/exclusion">
      <input type="hidden" name="excluded" value="false">
      <button class="btn-secondary payment-action-button text-xs" type="submit">恢复记录</button>
    </form>` : ''
  const toggleForm = `
    <form method="post" action="/admin/payments/${escapeHTML(id)}/toggle">
      <button class="payment-action-button rounded-full px-4 py-2 text-xs font-black text-white ${actionClass}" type="submit">${status === 'paid' ? '已付款，改为未付款' : '未付款，标记已付款'}</button>
    </form>`
  return `${remindButton}${excludeButton}${restoreForm}${toggleForm}`
}

initPaymentsTable()

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
  const openPanel = (label) => {
    title.textContent = label
    body.innerHTML = '<p class="text-sm text-stone-500">正在加载明细...</p>'
    panel.classList.remove('hidden')
  }
  const closePanel = () => panel.classList.add('hidden')
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

const dashboardTable = (headers, rows, footer = '') => {
  if (!rows.length) return `${footer}<p class="rounded-2xl bg-white/60 p-6 text-center text-sm text-stone-500">暂无记录。</p>`
  return `${footer}<div class="overflow-x-auto"><table class="w-full min-w-[620px] text-left text-sm"><thead class="text-stone-500"><tr>${headers.map((header) => `<th class="border-b border-stone-200 px-3 py-3">${header}</th>`).join('')}</tr></thead><tbody class="divide-y divide-stone-100">${rows.join('')}</tbody></table></div>`
}

const renderRoomsDetail = (rooms, fullURL) => dashboardTable(
  ['房间号', '标题', '状态', '楼层', '面积'],
  rooms.map((room) => `<tr><td class="px-3 py-3 font-black">${escapeHTML(room.room_no)}</td><td class="px-3 py-3">${escapeHTML(room.title)}</td><td class="px-3 py-3">${escapeHTML(room.status_label)}</td><td class="px-3 py-3">${escapeHTML(room.floor)}层</td><td class="px-3 py-3">${escapeHTML(room.area)}㎡</td></tr>`),
  `<p class="mb-4 text-sm font-bold text-stone-600">共 ${rooms.length} 间</p>`
) + fullListLink(fullURL)

const renderVacantRoomsDetail = (rooms, fullURL) => {
  const totalLoss = rooms.reduce((sum, room) => sum + Number(room.rent_price_fen || 0), 0)
  const summary = `<p class="mb-4 text-sm font-bold text-stone-600">共空置 ${rooms.length} 间，每月潜在损失：¥${formatFenText(totalLoss)}</p>`
  return dashboardTable(
    ['房间号', '标题', '租金', '楼层', '面积'],
    rooms.map((room) => `<tr><td class="px-3 py-3"><a class="font-black text-amber-800" href="${escapeHTML(room.detail_url)}">${escapeHTML(room.room_no)}</a></td><td class="px-3 py-3">${escapeHTML(room.title)}</td><td class="px-3 py-3">¥${escapeHTML(room.rent_price_text)} / ${escapeHTML(room.rent_type_label)}</td><td class="px-3 py-3">${escapeHTML(room.floor)}层</td><td class="px-3 py-3">${escapeHTML(room.area)}㎡</td></tr>`),
    summary
  ) + fullListLink(fullURL)
}

const renderOccupiedRoomsDetail = (rooms, fullURL) => dashboardTable(
  ['房间号', '当前租客', '租约到期日', '租金'],
  rooms.map((room) => `<tr><td class="px-3 py-3"><a class="font-black text-amber-800" href="${escapeHTML(room.detail_url)}">${escapeHTML(room.room_no)}</a></td><td class="px-3 py-3">${escapeHTML(room.tenant_name || '-')}</td><td class="px-3 py-3">${escapeHTML(room.lease_end_date ? formatDisplayDate(room.lease_end_date) : '长期')}</td><td class="px-3 py-3">¥${escapeHTML(room.rent_price_text)} / ${escapeHTML(room.rent_type_label)}</td></tr>`)
) + fullListLink(fullURL)

const renderTenantsDetail = (tenants, fullURL) => dashboardTable(
  ['姓名', '手机号', '房间号', '入住日期', '租金'],
  tenants.map((tenant) => `<tr><td class="px-3 py-3"><a class="font-black text-amber-800" href="${escapeHTML(tenant.detail_url)}">${escapeHTML(tenant.name)}</a></td><td class="px-3 py-3">${escapeHTML(tenant.phone)}</td><td class="px-3 py-3">${escapeHTML(tenant.room_no)}</td><td class="px-3 py-3">${escapeHTML(formatDisplayDate(tenant.checkin_date))}</td><td class="px-3 py-3">¥${escapeHTML(tenant.rent_price_text)} / ${escapeHTML(tenant.rent_type_label)}</td></tr>`)
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

initDashboardDetails()

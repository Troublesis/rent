# Payment Record Editing & Enhanced Status Tracking

Date: 2026-06-05

## Summary

Add inline editing for payment records, remove the checkout-only restriction on
exclusion, and track the actual payment received date so users can see when money
was collected.

## Motivation

- Landlords need to correct payment details (amount, notes) after creation.
- Excluding a record is useful for any tenant, not just checkout tenants (e.g.,
  duplicate records, offline settlements).
- When marking a payment as received, the system should record *when* the money
  arrived, not just that it was received.

## Design

### 1. Data Model — `PaidAt` field

Add `PaidAt *time.Time` to the `Payment` model:

```go
type Payment struct {
    // ... existing fields unchanged ...
    Paid          bool       `gorm:"default:false;index"`
    PaidAt        *time.Time // nil when unpaid; set when marked paid
    // ... rest unchanged ...
}
```

- GORM auto-migration handles the new column (NULL-able).
- `PayDate` remains the billing period / due date (unchanged semantics).
- `PaidAt` is the actual date money was received.

### 2. Remove checkout-only restriction on exclusion

Remove the tenant status guard in `PaymentService.SetExcluded`:

```go
// REMOVE this check:
if excluded && payment.Tenant.Status != model.TenantStatusCheckout {
    return fmt.Errorf("只有已退租租客的付款记录可以不再记录")
}
```

Exclusion is now available for all payment records regardless of tenant status.

### 3. Row actions — updated button layout

**Current per-row actions:**
- Normal: `不再记录` | `标为已付`
- Excluded: `恢复记录` | `标为未付`

**New per-row actions:**
- Normal: `编辑` | `标为已付`
- Excluded: `编辑` | `标为未付`

The `不再记录` and `恢复记录` actions both move into the edit panel. Inside
the edit panel:
- Non-excluded records show a `不再记录此记录` danger button.
- Excluded records show a `恢复记录` secondary button.

Applies to both desktop table rows and mobile cards.

### 4. Inline edit panel (slide-down)

Clicking `编辑` inserts an HTMX partial directly below the current row (desktop)
or below the current card (mobile).

**New routes:**

```
GET  /admin/payments/:id/edit    → returns edit panel partial (HTMX)
POST /admin/payments/:id         → updates payment (amount, note)
```

**Edit panel fields:**

| Field | Type | Default |
|-------|------|---------|
| 金额（元） | number input (yuan) | current amount |
| 备注 | textarea | current note |
| 不再记录此记录 / 恢复记录 | danger/secondary button | shown when not excluded / excluded |
| 保存 | submit | POSTs update via HTMX |
| 取消 | button | removes panel |

Implementation: The edit panel is an HTMX partial template
(`templates/components/payment_edit_panel.html`). Clicking `编辑` triggers
`hx-get="/admin/payments/:id/edit"` which returns the panel. The row/card gets
an expanded state (a new `<tr>` after the current row, or a nested `<div>` after
the card). `保存` posts via HTMX and refreshes the list. `取消` removes the
panel.

### 5. Status column — show paid received date

When `Paid = true` and `PaidAt` is set, display the received date in the status
column instead of just the type label:

**Current:**
```
[已付款]
租金
```

**New:**
```
[已付款]
2026年6月5日收款
租金
```

When `Paid = false`, no change (shows `未付款` badge + type label).

The `PaidAt` date is also shown in the mobile card layout.

### 6. TogglePaid updates

`PaymentService.TogglePaid` changes:

- When setting `Paid = true`: set `PaidAt = &now` (today's date).
- When setting `Paid = false`: set `PaidAt = nil`.

The `paymentRow` struct and `paymentToRow` function get a new `PaidAtLabel`
field for the formatted date string.

## Scope

- Model: 1 new field (`PaidAt`)
- Service: update `TogglePaid`, remove guard from `SetExcluded`, add `UpdatePayment`
- Handler: add `Edit` (GET panel partial), `Update` (POST save), update `Toggle`
- Templates: new edit panel partial, update row/card actions, update status column
- JS: wire `编辑` button to load panel, wire `取消` to close panel
- Routes: 2 new routes

## Out of scope

- Editing tenant, type, or pay date (can be added later if needed)
- Batch editing multiple records
- Payment history / audit trail
- PDF export

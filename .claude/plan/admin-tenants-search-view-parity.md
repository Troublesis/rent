# Admin Tenants Search and View Parity Plan

## Operation Type

Requirement Planning Type

## Objective

Update `/admin/tenants` to follow the current `/admin/rooms` page behavior for:

- Search field interaction.
- List/card swipe view state.
- HTMX partial refreshes.
- View preservation across filters, status chips, sort links, and clear actions.

The implementation should stay simple, preserve progressive enhancement, and keep all UI-facing text in Simplified Chinese.

## Current Findings

- `templates/admin/rooms.html` already has:
  - `data-room-view-root`
  - `data-room-view-swipe`
  - `data-room-view-pane="list|card"`
  - combobox-style room search
  - explicit HTMX form targeting `#room-list-section`
- `templates/admin/tenants.html` currently has:
  - `#tenant-list-section` without a view root data attribute
  - a plain search input
  - two horizontal panes without view pane data attributes
  - `hx-boost="true"` on the filter form
- `internal/handler/admin_tenant.go` already supports:
  - `?view=list|card`
  - status links
  - sort links
  - clear URL preserving view mode
- `internal/repository/tenant_repo.go` currently searches tenant name, phone, room number, and room title.
- `static/js/app.js` has reusable patterns but current room view logic is room-specific.

## Recommended Implementation Approach

Use the rooms page as the direct interaction reference while keeping tenant-specific code isolated enough to avoid regressions.

### Search Behavior

- Convert the tenant search input into a combobox-style field similar to the rooms page.
- Suggestions should load from `/api/tenants` and respect the current tenant status filter:
  - `active` → 在租
  - `checkout` → 已退租
  - `all` → 全部
- Suggestion labels should display as `姓名 - 房号 - 电话`.
- Typing should narrow the dropdown locally.
- Selecting a suggestion should submit the tenant filter form dynamically.
- To keep selected labels searchable, extend tenant repository search to also match the combined label expression.

### View Behavior

- Add tenants view root/swipe/pane attributes equivalent to rooms.
- Add tenant view JavaScript that mirrors the rooms page behavior:
  - initial scroll to `list` or `card` pane based on `.ViewMode`
  - update URL when swiping
  - update hidden `view` input
  - preserve current view on filters/status/sort/clear links
  - reinitialize after HTMX swaps

### JavaScript Structure

Prefer a small tenant-specific implementation that mirrors the current rooms code for low regression risk. Avoid broad refactoring of room view logic unless duplication becomes clearly problematic.

## Affected Files

- `templates/admin/tenants.html`
  - Add tenant view data attributes.
  - Replace the plain search input with a combobox-style search field.
  - Use explicit HTMX attributes like rooms page.
- `static/js/app.js`
  - Add tenant view initialization.
  - Add tenant search combobox initialization.
  - Reinitialize tenant widgets after HTMX swaps.
- `internal/repository/tenant_repo.go`
  - Extend query matching to support selected combined labels if needed.
- `internal/repository/tenant_repo_test.go` or existing tenant repository tests
  - Add focused coverage for combined tenant search labels if backend query is adjusted.
- `internal/handler/admin_tenant.go`
  - Verify URL generation; adjust only if view preservation gaps appear.

## Implementation Steps

1. Update `templates/admin/tenants.html`:
   - Add `data-tenant-view-root data-tenant-view="{{.ViewMode}}"` to `#tenant-list-section`.
   - Change the filter form to explicit `hx-get="/admin/tenants"` with the existing target/select/swap/push-url behavior.
   - Replace the plain search input with a tenant search combobox structure.
   - Add `data-tenant-view-swipe` and `data-tenant-view-pane="list|card"` to the result panes.
2. Add tenant view JavaScript in `static/js/app.js`:
   - Mirror the current room view functions with tenant selectors and `/admin/tenants` path.
   - Preserve view state on tenant links.
   - Initialize on page load and `htmx:afterSwap`.
3. Add tenant search combobox JavaScript in `static/js/app.js`:
   - Fetch `/api/tenants` with current status scope.
   - Render `姓名 - 房号 - 电话` options.
   - Filter options locally while typing.
   - Submit the tenant filter form when an option is selected.
   - Keep typing usable as a normal GET search fallback.
4. Verify backend tenant search compatibility:
   - If selecting the combined label does not match current search, extend `TenantRepository.ListTenants` to match `name || ' - ' || room_no || ' - ' || phone`.
   - Add a focused repository test for that search behavior.
5. Validate:
   - Run `gofmt` if Go files change.
   - Run `go test ./...`.
   - Run `node --check "static/js/app.js"`.
   - Manually verify `/admin/tenants` search, list/card view, status chips, sort links, clear link, and room page regression behavior.

## Acceptance Criteria

- `/admin/tenants?view=list` opens on the list pane.
- `/admin/tenants?view=card` opens on the card pane.
- Swiping tenants panes updates the hidden `view` input and browser URL.
- Tenant status chips preserve the active view.
- Tenant sort links preserve the active view.
- Tenant clear action preserves the active view.
- Tenant search uses a searchable dropdown field.
- Tenant suggestions display `姓名 - 房号 - 电话`.
- Typing narrows tenant suggestions.
- Suggestions are scoped by current status: 在租、已退租、全部.
- Selecting a tenant suggestion refreshes only `#tenant-list-section` through HTMX.
- Normal filter submit still works without JavaScript.
- Existing `/admin/rooms` search and list/card behavior remains unchanged.
- Existing tenant form combobox behavior remains unchanged.
- No English UI-facing text is introduced.

## Risks and Mitigations

- Tenant form already uses `data-tenant-combobox`.
  - Mitigation: use distinct attributes like `data-tenant-search-combobox`, `data-tenant-search-input`, and `data-tenant-search-list`.
- Broadly refactoring room view code could regress `/admin/rooms`.
  - Mitigation: add tenant-specific view logic first; extract later only if justified.
- HTMX swaps can remove event listeners.
  - Mitigation: make tenant initializers idempotent and call them after swaps.
- Combined selected labels may not match repository query.
  - Mitigation: add a targeted SQL search expression and test.

## Plan Status

Drafted and ready for execution after user confirmation.

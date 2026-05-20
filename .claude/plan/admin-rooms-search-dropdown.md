# Admin Rooms Search Dropdown Plan

## Operation Type

Requirement Planning Type

## Objective

Update the `/admin/rooms` search field so it behaves like the `/admin/payments` tenant searchable dropdown:

- Show room options as `房间号 - 标题`.
- Narrow dropdown options while typing.
- Limit available options by the current room status filter, such as 空置、已出租、维修.
- Preserve current HTMX dynamic list updates and existing URL query behavior.

## Current Findings

- `templates/admin/rooms.html` currently uses a plain `type="search"` input named `q` with HTMX attributes.
- `templates/admin/payments.html` uses a combobox-style tenant search field for the payment creation form.
- The rooms filter form already carries current state through hidden inputs: `view`, `status`, `sort_by`, and `sort_dir`.
- The existing rooms search backend can continue using `q` for room number/title matching.
- The room dropdown should avoid submitting the combined label if that label does not match backend search directly.

## Recommended Implementation Approach

Use a small room-specific combobox, reusing the existing payments tenant combobox interaction style without adding dependencies.

### Submitted Search Value

- Free typing: submit the typed text through `q` so title/room-number search keeps working.
- Option selection: display `房间号 - 标题` in the visible input, but submit the room number through `q` for reliable existing backend matching.

### Status Scope

- Read current status from the existing hidden `status` input in `#room-filter-form`.
- Request dropdown options according to that status:
  - `vacant` → only 空置 rooms
  - `occupied` → only 已出租 rooms
  - `maintenance` → only 维修/维护 rooms
  - `all` or empty → all rooms, using the existing API include-all behavior if available

## Affected Files

- `templates/admin/rooms.html`
  - Replace the plain search input with a combobox structure.
  - Preserve hidden state inputs and HTMX form behavior.
- `static/js/app.js`
  - Add room combobox initialization and filtering behavior.
  - Reinitialize after HTMX swaps.
- `internal/handler/admin_room.go` optional
  - Only adjust if the existing `/api/rooms` endpoint does not already support the required status/query combination.
- `internal/repository/room_repo.go` optional
  - Only adjust if exact filtering or API search support is insufficient.

## Implementation Steps

1. Inspect the existing tenant combobox JavaScript in `static/js/app.js` and mirror the interaction pattern for rooms.
2. Update `templates/admin/rooms.html`:
   - Keep `name="q"` on the effective submitted field or introduce a hidden `q` field synchronized by JavaScript.
   - Add a visible search input and dropdown list container.
   - Keep all visible UI copy in Simplified Chinese.
3. Implement room option loading in `static/js/app.js`:
   - Fetch room options using the current status and typed query.
   - Render labels as `room_no - title`.
   - Filter/narrow options while typing.
   - Support click selection, Escape close, outside-click close, and keyboard navigation where practical.
4. Ensure selection and typing still update the room list dynamically:
   - Typing should continue to refresh `#room-list-section` through HTMX or an equivalent existing fetch pattern.
   - Selecting a room should update `q`, push URL state, and refresh the list without a full-page reload.
5. Verify status-chip changes limit subsequent dropdown options to the active status after the HTMX swap.
6. Run lightweight validation:
   - `gofmt` only if Go files change.
   - `go test ./...` if backend code changes.
   - Browser/manual smoke check for `/admin/rooms` if available.

## Acceptance Criteria

- `/admin/rooms` search appears as a searchable dropdown.
- Dropdown options display `房间号 - 标题`.
- Typing narrows the dropdown options.
- Selecting an option filters the rooms list to the selected room.
- Free typing still searches by room number or title.
- 空置 status only offers vacant rooms.
- 已出租 status only offers occupied rooms.
- 维修/维护 status only offers maintenance rooms.
- 全部 status offers rooms across statuses.
- Status chips, sort links, clear link, and view switching still work.
- Payment tenant dropdown behavior remains unchanged.
- No English UI-facing text is introduced.

## Risks and Mitigations

- Combined display labels may not match backend `q` search.
  - Mitigation: submit room number on option selection, not the combined label.
- HTMX swaps may remove event listeners.
  - Mitigation: initialize room combobox on initial load and after HTMX swaps.
- JavaScript duplication with tenant combobox may grow.
  - Mitigation: keep the first implementation focused; extract shared helpers only if duplication becomes significant.

## Plan Status

Drafted and ready for execution after user confirmation.

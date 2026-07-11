# Plan: Fix department checkbox + show member list

## Problems
1. `toggleDepartment` calls `getMembers(deptId)` with `directOnly: true` — departments like 总公司/技术部 have 0 direct members, so the checkbox does nothing.
2. No member list is displayed — users can't select individual members.

## Solution

### 1. Fix department toggle to load ALL members (recursive)
- Add a new `getAllMembers` prop that calls `memberApi.list({ departmentId, page: 1, pageSize: 200 })` WITHOUT `directOnly: true`
- `toggleDepartment` uses `getAllMembers` to get all members under that dept (recursive)
- `getMembers` (directOnly) is still used for displaying the member list under a leaf node

### 2. Show member list on expand
- When a department is expanded, show its direct members (loaded via `getMembers` with `directOnly: true`)
- Each member has a Checkbox for individual selection/deselection
- Department checkbox reflects the state of ALL recursive members (uses `getAllMembers` result)

### Actually simpler approach:
Since the user said "勾选部门则选择部门下面所有的人" — the department checkbox should select ALL people recursively. But showing individual members should also work.

**Revised design:**
- Expand a department → show sub-departments + direct members (with `directOnly: true`)
- Click department checkbox → load ALL members recursively (without `directOnly`), select/deselect all
- Click individual member checkbox → toggle that single member
- Department checkbox state: all selected / indeterminate / none — based on ALL recursive members

### Implementation changes:

1. **`use-budget-page.ts`**: Add `getAllDeptMembers` callback without `directOnly`
2. **`budget-org-member-picker.tsx`**:
   - Add `getAllDeptMembers` prop
   - `toggleDepartment` uses `getAllDeptMembers` (recursive) for select/deselect all
   - On expand, call `loadDeptMembers` (directOnly) to show individual members
   - Render member rows below sub-departments when expanded
   - Department checkbox state based on recursive member loading
3. **Prop threading**: `budget-project-dialog.tsx` → `budget-detail-team.tsx` → `budget-page-shell.tsx`

### Simplified alternative (fewer API calls):
Since checking "all recursive members" requires an extra API call for state, let's simplify:
- Department checkbox toggles ALL members (loads without directOnly on click)
- Expand shows direct members (loaded with directOnly)
- Checkbox state: unchecked by default, checked after toggle, indeterminate if some members manually deselected

This avoids needing to pre-load all recursive members just to show checkbox state.

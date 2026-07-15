# Frontend Notification Scroll Area Bugfix Design

## Overview

The notification inbox is blocked because `apps/frontend/src/components/layout/notification-inbox.tsx` imports `@/components/ui/scroll-area`, but the shared primitive file is missing from `apps/frontend/src/components/ui`. The previous design proposed a temporary native `div` wrapper; that approach is rejected because it bypasses the repository's React + Radix/shadcn UI convention and does not provide the standard ScrollArea primitive contract.

The corrected design adds a standard Radix ScrollArea-based shared primitive with `Root`, `Viewport`, `Scrollbar`, and `Thumb` (and the optional Radix `Corner` where supported by the installed version). `notification-inbox.tsx` remains unchanged and continues to use `@/components/ui/scroll-area` with `className="h-80"`. The primitive owns scrolling mechanics and scrollbar rendering, while the consumer continues to own height, content, notification state, empty state, actions, and Popover composition.

This design updates only the spec document. It does not modify implementation code, `apps/frontend/package.json`, `pnpm-lock.yaml`, `notification-inbox.tsx`, or `docs/Frontend.md`, and it does not run tests, builds, `tsc`, lint, or a development server.

## Glossary

- **Bug_Condition (C)**: A frontend module graph reaches `notification-inbox.tsx` while `@/components/ui/scroll-area` cannot resolve to the shared ScrollArea module.
- **Property (P)**: The resolved shared module exposes the standard Radix ScrollArea structure, preserves the consumer's `h-80`, renders children unchanged, and supplies a functioning vertical scrollbar for overflowing content.
- **Preservation**: Existing notification inbox behavior remains unchanged, including Popover structure, fixed height, children, empty state, notification rendering, unread actions, mark-all-as-read behavior, and query invalidation.
- **Radix ScrollArea primitive**: The `@radix-ui/react-scroll-area` primitives used as the implementation foundation: `Root`, `Viewport`, `Scrollbar`, `Thumb`, and optionally `Corner`.
- **shadcn ScrollArea wrapper**: The project-local `apps/frontend/src/components/ui/scroll-area.tsx` module that composes Radix primitives, applies project styling through `cn`, and exports the consumer-facing `ScrollArea` API.
- **Notification inbox**: `apps/frontend/src/components/layout/notification-inbox.tsx`, currently the only known consumer of `ScrollArea`.
- **`cn`**: The existing class merge helper at `apps/frontend/src/lib/utils.ts`, used by the local shadcn-style UI primitives.

## Bug Details

### Bug Condition

The bug manifests when the frontend resolves or renders the notification inbox. Its established alias import, `@/components/ui/scroll-area`, targets `apps/frontend/src/components/ui/scroll-area.tsx`, but that module is absent. The previous proposed workaround would resolve the import with a native wrapper, but it would still be the wrong architectural fix because it would not implement the standard Radix Root/Viewport/Scrollbar/Thumb contract required by the project's UI conventions.

**Formal Specification:**

```text
FUNCTION isBugCondition(input)
  INPUT: input of type FrontendModuleResolutionRequest
  OUTPUT: boolean

  RETURN input.importer == "apps/frontend/src/components/layout/notification-inbox.tsx"
         AND input.importPath == "@/components/ui/scroll-area"
         AND input.resolvesTo("apps/frontend/src/components/ui/scroll-area.tsx") == false
END FUNCTION
```

The bug condition is about the missing shared module. The corrected implementation must not change the alias, replace the import with a relative path, add notification-specific scrolling logic, or use a temporary native `div` wrapper as the ScrollArea implementation.

### Examples

- **Populated inbox:** Resolving the notification inbox reaches the missing import before notification rows can render. After the fix, the existing mapped `NotificationItemRow` children render inside the Radix Viewport without changing their order, labels, styles, or click handlers.
- **Empty inbox:** The unresolved import prevents the existing `暂无通知` empty state from rendering. After the fix, that same empty-state `div` remains the direct child content supplied to `ScrollArea` and retains its current centered layout.
- **Long list:** With enough rows to exceed the consumer's `h-80`, the Radix Viewport becomes the scrollable region and the vertical `Scrollbar`/`Thumb` provide the standard UI affordance. The Popover remains bounded instead of growing with the list.
- **Short list:** When the content fits within `h-80`, the Radix scrollbar does not alter the existing visual layout in a material way; the rows and Popover remain as before.
- **Existing interaction:** Clicking an unread row still calls `notificationApi.markRead`, invalidates `['notifications']` and `['notifications', 'unread-count']`, and updates the existing unread presentation. Clicking a read row remains a no-op for marking read; `全部已读` continues to use the existing handler.
- **Non-consumer UI:** Components that do not import this module remain outside `isBugCondition` and must not be changed by the fix.

## Expected Behavior

### Correct Behavior Specification

The expected behavior is defined by the following pseudocode. It describes the public contract of the shared wrapper and the unchanged consumer contract; it does not authorize changing `notification-inbox.tsx`.

```text
FUNCTION expectedBehavior(input)
  INPUT: input of type ScrollAreaRenderRequest
  OUTPUT: boolean

  RETURN input.importResolves == true
         AND input.rootUsesRadixScrollArea == true
         AND input.rootContainsViewport == true
         AND input.viewportContainsUnchangedChildren == true
         AND input.rootPreservesClass("h-80") == true
         AND input.rootProvidesVerticalScrollbarAndThumb == true
END FUNCTION
```

The intended composition is:

```text
<ScrollArea.Root className={cn("relative", className)} {...rootProps}>
  <ScrollArea.Viewport className="size-full rounded-[inherit]">
    {children}
  </ScrollArea.Viewport>
  <ScrollArea.Scrollbar orientation="vertical">
    <ScrollArea.Thumb />
  </ScrollArea.Scrollbar>
  <ScrollArea.Corner />             // only if supported/appropriate in the local Radix version
</ScrollArea.Root>
```

The wrapper may export a `ScrollBar` helper in the same way as standard shadcn components, but the existing consumer must only need the existing `ScrollArea` export.

### Preservation Requirements

**Unchanged Behaviors:**

- `notification-inbox.tsx` continues to import `ScrollArea` from `@/components/ui/scroll-area`; no consumer rewrite or native wrapper is allowed.
- The consumer's `className="h-80"` remains applied to the ScrollArea root and continues to bound the notification list height.
- The existing `children` expression remains unchanged and ordered: either the centered `暂无通知` empty state or the mapped notification row buttons.
- Notification title, optional body, relative time, unread indicator, unread styling, and the `99+` badge cap remain unchanged.
- Clicking an unread notification continues to mark it read and invalidate the existing notification and unread-count queries; clicking an already-read notification remains a no-op for marking read.
- The `全部已读` button, its visibility condition, mutation, and query invalidation remain unchanged.
- The `Popover`, `PopoverTrigger`, `PopoverContent`, header, title, width, padding, and overall inbox placement remain unchanged.
- Components and flows that do not import the missing ScrollArea path remain unaffected.

**Scope:**

The implementation scope is the shared UI primitive plus the dependency metadata required to use the real Radix primitive. It must not change notification APIs, data fetching, query keys, event handlers, Popover composition, alias configuration, Tailwind configuration, or unrelated UI components. For non-buggy inputs (`NOT C`), the fixed implementation must preserve the original behavior (`F(X) = F'(X)`).

### Repository Evidence for Shared Component Documentation

The workspace rule requires updating the §2 shared-component inventory when a shared UI component is added. Inspection of `docs/Frontend.md` shows that §2 contains directory layout, component placement rules, and a component ownership table, but no enumerated shared-component inventory or list of individual files. The document also contains no `shared components`, `共享组件`, or equivalent inventory section. Therefore, no `docs/Frontend.md` change is required for this generic `components/ui/scroll-area.tsx` addition; the implementation task should retain this evidence in its completion notes. If a future documentation change introduces an explicit inventory, this component must then be listed there.

## Hypothesized Root Cause

1. **Missing shared UI primitive**: `notification-inbox.tsx` was authored against the project's shadcn-style `@/components/ui/scroll-area` convention, but `apps/frontend/src/components/ui/scroll-area.tsx` is absent. The module graph fails before either populated or empty notification content can render.

2. **Previous design selected the wrong implementation class**: The prior proposal treated the missing component as a native `div` with `overflow-y-auto`. That would be a hacky compatibility wrapper rather than the established Radix/shadcn primitive and would omit the standard Viewport, Scrollbar, and Thumb behavior.

3. **Required Radix package is not installed**: `apps/frontend/package.json` and the `apps/frontend` importer in `pnpm-lock.yaml` contain other Radix packages but no `@radix-ui/react-scroll-area`. A standard implementation therefore requires a dependency and lockfile update; excluding them would make the proposed implementation incomplete.

4. **Alias and consumer contract are not the root cause**: The consumer uses the established `@/` alias and only requires the existing `ScrollArea` export with `className` and `children`. There is no reason to modify alias configuration or notification logic.

5. **Height/class forwarding is a potential regression point**: A wrapper that drops `className` would remove `h-80`; a wrapper that moves the height to the wrong element could prevent the Viewport from filling the bounded root. The Radix Root must receive the consumer class, and the Viewport must fill the Root with the standard `size-full` behavior.

## Correctness Properties

Property 1: Bug Condition - Radix ScrollArea Resolves and Provides Bounded Vertical Scrolling

_For any_ frontend module-resolution request where `isBugCondition(input)` returns true, the fixed source SHALL resolve `@/components/ui/scroll-area` to a project-local shadcn wrapper backed by `@radix-ui/react-scroll-area`, compose a Root containing a Viewport and vertical Scrollbar/Thumb, preserve the consumer's `h-80`, and render the supplied notification children unchanged so overflowing content can scroll inside the Popover.

**Validates: Requirements 2.1, 2.2, 2.3, 2.4**

Property 2: Preservation - Notification Inbox Contract and Existing Interactions

_For any_ frontend input where `isBugCondition(input)` returns false, the fixed source SHALL produce the same consumer, notification, Popover, layout, and interaction behavior as the original source, preserving children, empty state, fixed height, unread actions, mark-all-as-read behavior, and unrelated frontend behavior.

**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6**

## Fix Implementation

### Changes Required

#### 1. Add the standard shared primitive

**File:** `apps/frontend/src/components/ui/scroll-area.tsx`

Implement the local shadcn wrapper using `@radix-ui/react-scroll-area` and the project's existing `cn` helper. Follow the style of `popover.tsx`, `dialog.tsx`, and other UI primitives: use `import * as React from 'react'`, add `'use client'` if required by the local generated-component convention, use `React.ComponentProps<typeof ...>` types, add `data-slot` markers, merge classes with `cn`, and export named components.

Required structure and behavior:

1. Export `ScrollArea` as the consumer-facing wrapper around `ScrollAreaPrimitive.Root`.
2. Render `ScrollAreaPrimitive.Viewport` inside Root and render `children` only inside Viewport, preserving child order and identity.
3. Provide a standard `ScrollBar` helper around `ScrollAreaPrimitive.Scrollbar`; default its orientation to `vertical` and render `ScrollAreaPrimitive.Thumb` inside it.
4. Apply the consumer `className` to Root through `cn`, preserving `h-80`; do not replace or hardcode the consumer's height.
5. Give Viewport the standard full-size/inherited-radius classes needed for the Root's bounded height to control the scroll region.
6. Use the standard Radix scrollbar orientation classes and thumb styling, with a horizontal branch only if the shared helper follows the generated shadcn contract. The inbox relies on the vertical default.
7. Include `ScrollAreaPrimitive.Corner` only if it is part of the installed primitive API and does not affect the one-axis inbox layout.
8. Forward supported Radix Root/Scrollbar props and preserve `data-slot` markers consistent with local UI primitives. Do not add notification-specific props or behavior.
9. Do not create a native `div` replacement, a second bespoke wrapper layer, or a consumer-specific selector.

#### 2. Add the required dependency metadata

**Files:** `apps/frontend/package.json`, `pnpm-lock.yaml`

`@radix-ui/react-scroll-area` **is required** for the standard implementation; it is not currently present in either file. Add it to `apps/frontend` runtime dependencies and regenerate the workspace lockfile with the repository's package manager in the coding phase. Use the repository's current Radix compatibility strategy: existing Radix dependencies use caret ranges within their compatible major line (for example `^1.1.x`, `^1.2.x`, and `^2.3.x`), while the lockfile records the exact resolved package and peer-compatible React 19 resolution. Select the current compatible `1.x` ScrollArea release at implementation time, use the matching caret specifier rather than an unrelated major or a hand-written lock entry, and let pnpm update the transitive Radix packages/peer snapshots as needed.

The design phase does not edit either dependency file. The implementation task must not manually invent an integrity hash or commit a package entry without the corresponding lockfile resolution.

#### 3. Keep the consumer and documentation scope explicit

**No change:** `apps/frontend/src/components/layout/notification-inbox.tsx`. Preserve its import, `h-80`, children, empty state, read operations, and Popover behavior exactly.

**No change:** `docs/Frontend.md`. The repository evidence above shows that §2 has no individual shared-component inventory to update. The implementation task should record that inspection result rather than add a speculative list.

**No change:** aliases, Tailwind/CSS configuration, notification APIs, query definitions, and unrelated UI primitives.

### Intended Implementation Shape (Pseudocode)

```text
FUNCTION ScrollArea({ className, children, ...rootProps })
  RETURN RadixScrollArea.Root(
    data-slot = "scroll-area",
    className = cn("relative", className),
    rootProps,
    RadixScrollArea.Viewport(
      data-slot = "scroll-area-viewport",
      className = "size-full rounded-[inherit] ...",
      children
    ),
    ScrollBar(orientation = "vertical"),
    RadixScrollArea.Corner(if supported)
  )
END FUNCTION

FUNCTION ScrollBar({ orientation = "vertical", className, ...props })
  RETURN RadixScrollArea.Scrollbar(
    data-slot = "scroll-area-scrollbar",
    orientation,
    className = cn(orientationClasses(orientation), className),
    props,
    RadixScrollArea.Thumb(
      data-slot = "scroll-area-thumb",
      className = "relative flex-1 rounded-full bg-border"
    )
  )
END FUNCTION

EXPORT ScrollArea, ScrollBar
```

This is implementation guidance for the next coding phase, not an implementation change in this design phase.

## Testing Strategy

### Validation Approach

Validation follows the bug-condition methodology: first capture the missing-module counterexample on the unfixed tree, then verify the Radix primitive and dependency resolution after implementation, and finally verify preservation of the notification inbox contract. Per the user's explicit instruction, no tests, builds, type checks, lint, or dev server are run in this design phase.

### Exploratory Bug Condition Checking

**Goal:** Confirm the pre-fix counterexample and ensure the root cause is the missing shared module rather than the alias or notification logic.

**Test Plan:** During the implementation phase, inspect or run the repository-approved frontend validation only after the user allows it. The expected unfixed observation is a module-resolution failure at `@/components/ui/scroll-area`; both populated and empty inbox paths are blocked before their children mount.

**Test Cases:**

1. **Import resolution:** Resolve `notification-inbox.tsx` and record the missing `scroll-area` module diagnostic.
2. **Populated notification path:** Reach the header Popover with notification data and record that rows cannot mount on the unfixed tree.
3. **Empty notification path:** Reach the inbox with no notifications and record that `暂无通知` also cannot mount.
4. **Consumer inventory:** Confirm the known current consumer and ensure no additional import requires a different public API.
5. **Dependency inventory:** Confirm that `@radix-ui/react-scroll-area` is absent before the dependency update and present in both package metadata and lockfile after it.

**Expected Counterexamples:**

- The resolver cannot find `@/components/ui/scroll-area`.
- The header notification experience fails before populated or empty children render.
- A native-wrapper-only solution would resolve the file but fail the required Radix Root/Viewport/Scrollbar/Thumb design contract; this is an architectural counterexample to the previous design, not an acceptable fix.

### Fix Checking

**Goal:** Verify that all inputs satisfying the bug condition resolve to the standard Radix-backed shared primitive and preserve the bounded inbox behavior.

**Pseudocode:**

```text
FOR ALL input WHERE isBugCondition(input) DO
  result := ScrollArea_fixed(input.className, input.children)
  ASSERT expectedBehavior(result)
END FOR
```

Planned checks include:

- `notification-inbox.tsx` resolves its existing alias import.
- `apps/frontend/package.json` declares the Radix ScrollArea runtime dependency and the lockfile resolves it with React 19-compatible peers.
- The rendered structure contains Root, Viewport, vertical Scrollbar, and Thumb.
- `h-80` remains on the Root, the Viewport fills it, and overflowing content is scrollable.
- Empty and populated children render unchanged and in their original order.

### Preservation Checking

**Goal:** Verify that the fixed primitive does not change behavior outside the missing-module condition.

**Pseudocode:**

```text
FOR ALL input WHERE NOT isBugCondition(input) DO
  ASSERT behavior_original(input) = behavior_fixed(input)
END FOR
```

**Testing Approach:** Use focused unit and integration examples plus property-based cases in the implementation phase. Preservation is established by leaving the consumer untouched and checking that the wrapper forwards Root props, merges classes, preserves children, and does not alter notification handlers or Popover composition.

### Unit Tests

- Verify the shared module exports `ScrollArea` and `ScrollBar` and is backed by Radix primitives rather than a native-wrapper implementation.
- Verify the rendered structure includes the Root, Viewport, vertical Scrollbar, and Thumb, with the expected `data-slot` markers.
- Verify caller classes such as `h-80` remain on the Root and are merged with base classes.
- Verify empty-state content and notification row children render unchanged and in order.
- Verify supported Root/Viewport props are forwarded without introducing notification-specific behavior.

### Property-Based Tests

- Generate caller class strings and child sequences; verify class preservation and child order through the Root/Viewport composition.
- Generate short and overflowing content states; verify the Root remains bounded by the caller's `h-80` and the vertical Scrollbar/Thumb contract remains present.
- Generate generic Root props and non-notification children; verify the primitive remains business-agnostic and preserves behavior for unrelated consumers.
- Generate notification read states and actions at the consumer boundary; verify the ScrollArea change does not alter which existing handlers are invoked.

These are planned checks only; no property-based test is run in this design phase.

### Integration Tests

- Resolve and bundle the frontend notification inbox with the existing `@/components/ui/scroll-area` import.
- Open the header notification Popover with no notifications and verify the existing empty state remains visible inside the bounded `h-80` area.
- Open it with enough notifications to overflow `h-80` and verify the Radix Viewport scrolls through the list using the vertical Scrollbar/Thumb without changing Popover header/actions.
- Verify unread row click, read-row no-op behavior, mark-all-as-read, unread badge, and query invalidation remain unchanged.
- Verify unrelated UI consumers and non-notification flows are unaffected.

No tests, builds, type checks, lint, or dev servers are run now, per the user's explicit instruction. The next coding phase should implement only the documented file/dependency changes and then perform the authorized validation separately.

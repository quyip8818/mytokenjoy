# Bugfix Implementation Plan

> **Scope:** Restore the shared Radix/shadcn ScrollArea primitive used by the notification inbox. This task list contains only dependency updates, implementation work, and code/scope checks. Do not create test files.
>
> **Validation restriction:** Do not run tests, property-based tests, TypeScript/`tsc`, builds, lint, or a development server. The checks below are inspection and scope checks only, as explicitly requested.

- [x] 1. Add the Radix ScrollArea dependency and synchronize the lockfile
  - Add `@radix-ui/react-scroll-area` to the `apps/frontend` runtime dependencies using the repository's `pnpm` package manager and the compatible current `1.x` release/caret range used by the other Radix packages.
  - Run the appropriate `pnpm` dependency-update command from the repository so that both `apps/frontend/package.json` and the workspace `pnpm-lock.yaml` are generated together.
  - Do not hand-edit `pnpm-lock.yaml`, invent integrity values, or add a lockfile entry without the package-manager resolution.
  - Keep all existing dependency versions and unrelated importer entries unchanged unless `pnpm` updates the required Radix peer-resolution metadata as part of the dependency sync.
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 2. Add the standard shared Radix/shadcn ScrollArea primitive
  - Create `apps/frontend/src/components/ui/scroll-area.tsx`; do not modify `apps/frontend/src/components/layout/notification-inbox.tsx`.
  - Implement `ScrollArea` as a wrapper around `@radix-ui/react-scroll-area` `Root`, forwarding supported Root props and preserving `className`, `children`, and `data-slot="scroll-area"`.
  - Render `Viewport` inside `Root`, keep `children` unchanged and in the Viewport, and apply the standard full-size/inherited-radius classes so the consumer's `h-80` continues to bound the scroll region.
  - Export `ScrollBar` as the standard helper around Radix `Scrollbar`; default its orientation to `vertical`, render a Radix `Thumb` inside it, and preserve `data-slot` markers, class merging through the existing `cn` helper, and the standard vertical/horizontal styling contract.
  - Include Radix `Corner` only when supported by the installed package and appropriate for this one-axis consumer.
  - Follow the local shadcn/Radix component conventions for React imports, `React.ComponentProps` types, named exports, client directive usage if required locally, and `data-slot` attributes.
  - Do not implement a native `div` wrapper, notification-specific scrolling logic, a second bespoke wrapper, or any consumer-specific selector.
  - _Bug_Condition: `isBugCondition(input)` from the design, where the notification inbox import cannot resolve to the missing shared module._
  - _Expected_Behavior: `expectedBehavior(result)` from the design, where the alias resolves to a Radix Root containing Viewport and vertical Scrollbar/Thumb, preserves `h-80`, and renders children unchanged._
  - _Preservation: Preserve the notification consumer, Popover composition, children, empty state, notification actions, query invalidation, and unrelated frontend behavior described in the design._
  - _Requirements: 2.1, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [-] 3. Perform implementation and scope checks by inspection
  - Confirm `apps/frontend/package.json` declares `@radix-ui/react-scroll-area` and `pnpm-lock.yaml` contains the corresponding package-manager-generated importer/package resolution; do not manually alter the lockfile during this check.
  - Confirm `scroll-area.tsx` exports `ScrollArea` and `ScrollBar`, uses the installed Radix primitives `Root`, `Viewport`, `Scrollbar`, and `Thumb`, forwards `className`/Root props, preserves `children`, applies `data-slot` markers, and keeps the consumer height on the Root.
  - Confirm `apps/frontend/src/components/layout/notification-inbox.tsx` is unchanged, including its alias import, `className="h-80"`, notification children, empty state, read actions, and Popover composition.
  - Confirm no native `div` replacement, notification-specific behavior, alias/configuration change, Tailwind/CSS change, unrelated UI change, or test file was added.
  - Confirm `docs/Frontend.md` is unchanged and record in the implementation notes that inspection of §2 found directory/layout and ownership guidance but no enumerated shared-component inventory; therefore no component-list update is required.
  - Do not run tests, `tsc`, builds, lint, or a dev server; this task is limited to source, dependency, and scope inspection.
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [~] 4. Checkpoint — confirm the final change set is within scope
  - Confirm the intended change set contains only `apps/frontend/package.json`, the package-manager-generated `pnpm-lock.yaml`, and the new `apps/frontend/src/components/ui/scroll-area.tsx`.
  - Confirm `notification-inbox.tsx`, `docs/Frontend.md`, aliases, Tailwind/CSS configuration, notification APIs, query definitions, and unrelated components were not modified.
  - Confirm no test files were created and no prohibited test, TypeScript, build, lint, or dev-server command was run.
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

## Dependency Order

`1` must complete first because the component imports `@radix-ui/react-scroll-area` and the package manager must generate the matching lockfile resolution. `2` depends on `1` and adds the shared primitive without changing its consumer. `3` depends on `1` and `2` and performs only the requested source/dependency/scope inspection. `4` depends on all preceding tasks and records final scope compliance. No task authorizes running tests, `tsc`, builds, lint, or a development server.

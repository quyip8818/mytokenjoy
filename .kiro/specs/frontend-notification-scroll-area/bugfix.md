# Bugfix Requirements Document

## Introduction

The frontend notification inbox cannot render because `apps/frontend/src/components/layout/notification-inbox.tsx` imports `@/components/ui/scroll-area`, but the corresponding shared UI module is not present in `apps/frontend/src/components/ui`. The Vite alias configuration maps `@/*` to `src/*` as expected, so the observed failure is a missing or mismatched shared primitive rather than an identified alias configuration defect. The current full frontend TypeScript build confirms the same defect with `TS2307`.

The repair scope is limited to restoring or aligning the generic scroll-area UI capability used by the notification inbox, checking every frontend consumer of that capability, and making the complete frontend TypeScript validation pass. Backend behavior and unrelated frontend features are outside this bugfix.

The bug condition is `C(X)`: `X` is a frontend compile or render path that reaches the notification inbox while the imported `@/components/ui/scroll-area` module cannot be resolved. The fix property is `P(result)`: the shared module resolves, the frontend compiles without TypeScript diagnostics, and the notification inbox renders its existing scrollable content and interactions. The preservation goal is `F(X) = F'(X)` for every input `X` where `NOT C(X)`; behavior outside the missing-module condition must remain unchanged.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN the frontend compiles or bundles `apps/frontend/src/components/layout/notification-inbox.tsx` with its `@/components/ui/scroll-area` import THEN the system fails to resolve the module and the notification UI cannot render.

1.2 WHEN the complete frontend TypeScript build runs with `npx tsc -b tsconfig.app.json tsconfig.node.json --pretty false` THEN the system reports `TS2307` for the missing scroll-area module and exits unsuccessfully.

1.3 WHEN the notification inbox is reached through the frontend header THEN the unresolved shared UI dependency prevents the header's notification experience from loading normally, regardless of whether notification data is empty or populated.

### Expected Behavior (Correct)

2.1 WHEN the frontend compiles or bundles `apps/frontend/src/components/layout/notification-inbox.tsx` with its `@/components/ui/scroll-area` import THEN the system SHALL resolve the shared UI module and allow the notification inbox and containing header to render without a Vite module-resolution error.

2.2 WHEN the complete frontend TypeScript build runs with `npx tsc -b tsconfig.app.json tsconfig.node.json --pretty false` THEN the system SHALL exit successfully with no TypeScript diagnostics, including no unresolved-module diagnostic for `scroll-area`.

2.3 WHEN the frontend source tree and Vite configuration are checked through the supported TypeScript project references THEN the system SHALL validate the full frontend scope, not only `notification-inbox.tsx`, and SHALL contain no remaining TypeScript errors caused by this bugfix.

2.4 WHEN the notification inbox is opened after the shared module resolves and notifications are populated THEN the system SHALL continue to present a fixed-height scrollable notification list, render each notification, and preserve the existing empty-state rendering when no notifications are available.

2.5 WHEN a user interacts with the repaired notification inbox THEN the system SHALL continue to mark an unread notification as read, refresh notification and unread-count queries, and offer the existing mark-all-as-read action when unread notifications exist.

### Unchanged Behavior (Regression Prevention)

3.1 WHEN the notification inbox is rendered with no notifications and the missing-module condition is not present THEN the system SHALL CONTINUE TO display the existing `暂无通知` empty state within the notification content area.

3.2 WHEN notifications are available and the missing-module condition is not present THEN the system SHALL CONTINUE TO show notification title, optional body, relative time, unread styling, and the unread-count badge with its existing `99+` cap.

3.3 WHEN a user clicks an unread notification and the missing-module condition is not present THEN the system SHALL CONTINUE TO call the existing mark-read operation and invalidate the notifications and unread-count queries; clicking an already-read notification SHALL CONTINUE not to mark it read.

3.4 WHEN the unread count is zero or positive and the missing-module condition is not present THEN the system SHALL CONTINUE to hide or show the existing mark-all-as-read control according to the count and preserve the notification inbox placement in the frontend header.

3.5 WHEN any frontend module does not import the missing scroll-area path and the missing-module condition is not present THEN the system SHALL CONTINUE to preserve its existing compile, render, and interaction behavior.

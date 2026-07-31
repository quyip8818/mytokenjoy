export const queryKeys = {
  session: { all: ['session'] as const },
  suppliers: {
    all: ['suppliers'] as const,
    list: (filter?: unknown) => ['suppliers', 'list', filter] as const,
    detail: (id: string) => ['suppliers', 'detail', id] as const,
    options: () => ['suppliers', 'options'] as const,
  },
  contracts: {
    all: ['contracts'] as const,
    list: (filter?: unknown) => ['contracts', 'list', filter] as const,
    detail: (id: string) => ['contracts', 'detail', id] as const,
  },
  orders: {
    all: ['orders'] as const,
    list: (filter?: unknown) => ['orders', 'list', filter] as const,
    detail: (id: string) => ['orders', 'detail', id] as const,
  },
  evaluations: {
    all: ['evaluations'] as const,
    list: (filter?: unknown) => ['evaluations', 'list', filter] as const,
    detail: (id: string) => ['evaluations', 'detail', id] as const,
    weights: () => ['evaluations', 'weights'] as const,
  },
  dashboard: {
    all: ['dashboard'] as const,
    cards: () => ['dashboard', 'cards'] as const,
    charts: () => ['dashboard', 'charts'] as const,
    expiring: () => ['dashboard', 'expiring'] as const,
    recentOrders: () => ['dashboard', 'recent-orders'] as const,
  },
}

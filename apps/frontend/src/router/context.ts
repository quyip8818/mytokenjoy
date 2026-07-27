// ponytail: 暂不注入 queryClient，因为当前不用 router loader。
// 升级路径：需要 loader 时，把 queryClient 提到 main.tsx 层传入。
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface RouterContext {
  // intentionally empty — reserved for future loader integration
}

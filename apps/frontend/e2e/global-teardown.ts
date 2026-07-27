import { dropDatabase } from './e2e-db'

/**
 * ponytail: drop e2e 库，下次跑重建 + 重新 seed。
 * 设 E2E_KEEP_DB=1 保留数据做 debug。
 */
export default function globalTeardown() {
  if (process.env.E2E_KEEP_DB) return
  dropDatabase()
}

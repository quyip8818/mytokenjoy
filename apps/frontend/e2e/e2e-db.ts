/**
 * ponytail: E2E 环境配置集中管理。升级路径：testcontainers。
 */
import { execSync } from 'node:child_process'

// --- Test mode ---
export type E2EMode = 'saas' | 'local'
export const E2E_MODE: E2EMode = (process.env.TEST_MODE ?? 'saas') as E2EMode

// --- Ports (偏移到 942x，避免和 dev 的 941x 冲突) ---
export const E2E_BACKEND_PORT = 9420
export const E2E_SMS_PORT = 9421
export const E2E_PREVIEW_PORT = 9422
export const E2E_HOST = '127.0.0.1'
export const E2E_BASE_URL = `http://${E2E_HOST}:${E2E_PREVIEW_PORT}`

// --- Postgres (测试专用 PG:5530) ---
export const PG_HOST = '127.0.0.1'
export const PG_PORT = '5530'
export const PG_USER = 'tokenjoy'
export const E2E_DB = `tokenjoy_e2e_${E2E_MODE}`
export const E2E_DATABASE_URL = `postgres://${PG_USER}:${PG_USER}@${PG_HOST}:${PG_PORT}/${E2E_DB}?sslmode=disable`

const ADMIN_CONN = `postgres://${PG_USER}:${PG_USER}@${PG_HOST}:${PG_PORT}/postgres`

/** 幂等创建 e2e 数据库 */
export function createDatabase() {
  try {
    execSync(`createdb -h ${PG_HOST} -p ${PG_PORT} -U ${PG_USER} ${E2E_DB}`, {
      stdio: 'pipe',
      env: { ...process.env, PGPASSWORD: PG_USER },
    })
  } catch (err) {
    // 42P04 = database already exists — safe to ignore
    if (err instanceof Error && err.message.includes('already exists')) return
    console.warn(`[e2e] createdb failed, assuming ${E2E_DB} exists.`)
  }
}

/** 强制 drop e2e 数据库 */
export function dropDatabase() {
  try {
    execSync(`psql "${ADMIN_CONN}" -c "DROP DATABASE IF EXISTS ${E2E_DB} WITH (FORCE)"`, {
      stdio: 'pipe',
      env: { ...process.env, PGPASSWORD: PG_USER },
    })
  } catch {
    // best-effort
  }
}

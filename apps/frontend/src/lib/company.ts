import type { CompanyType } from '@/api/types/common'

/**
 * Returns true for non-production account types that allow
 * test-model access, simulated consumption, and mock quota grants.
 * ponytail: single source of truth — mirrors backend company.IsTestingAccount.
 */
export function isTestingAccount(companyType: CompanyType): boolean {
  return companyType === 'trial' || companyType === 'demo' || companyType === 'testing'
}

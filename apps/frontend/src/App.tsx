import { BrowserRouter, Routes, Route, Navigate } from 'react-router'
import { AdminLayout } from '@/components/layout/admin-layout'
import { MemberLayout } from '@/components/layout/member-layout'
import DataSourcePage from '@/routes/org/data-source'
import StructurePage from '@/routes/org/structure'
import RolesPage from '@/routes/org/roles'
import CostDashboardPage from '@/routes/dashboard/cost'
import UsageDashboardPage from '@/routes/dashboard/usage'
import BudgetPage from '@/routes/budget'
import BudgetAlertsPage from '@/routes/budget/alerts'
import ProviderKeysPage from '@/routes/keys/provider'
import PlatformKeysPage from '@/routes/keys/platform'
import ApprovalPage from '@/routes/keys/approval'
import ModelListPage from '@/routes/models/list'
import ModelRoutingPage from '@/routes/models/routing'
import OperationLogsPage from '@/routes/audit/operations'
import CallLogsPage from '@/routes/audit/calls'
import MemberDashboardPage from '@/routes/member'
import MemberCallLogsPage from '@/routes/member/call-logs'
import MemberKeysPage from '@/routes/member/keys'
import WalletPage from '@/routes/wallet'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AdminLayout />}>
          <Route index element={<Navigate to="/dashboard/cost" replace />} />
          <Route path="dashboard/cost" element={<CostDashboardPage />} />
          <Route path="dashboard/usage" element={<UsageDashboardPage />} />
          <Route path="org/data-source" element={<DataSourcePage />} />
          <Route path="org/structure" element={<StructurePage />} />
          <Route path="org/roles" element={<RolesPage />} />
          <Route path="budget" element={<BudgetPage />} />
          <Route path="budget/alerts" element={<BudgetAlertsPage />} />
          <Route path="keys/provider" element={<ProviderKeysPage />} />
          <Route path="keys/platform" element={<PlatformKeysPage />} />
          <Route path="keys/approval" element={<ApprovalPage />} />
          <Route path="models/list" element={<ModelListPage />} />
          <Route path="models/routing" element={<ModelRoutingPage />} />
          <Route path="audit/operations" element={<OperationLogsPage />} />
          <Route path="audit/calls" element={<CallLogsPage />} />
          <Route path="wallet" element={<WalletPage />} />
        </Route>
        <Route element={<MemberLayout />}>
          <Route path="me" element={<MemberDashboardPage />} />
          <Route path="me/keys" element={<MemberKeysPage />} />
          <Route path="me/call-logs" element={<MemberCallLogsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

import { PageShell } from '@/components/layout/page-shell'
import type { LoginActivityPageState } from '../hooks/use-login-activity-page'
import { LoginActivityPanel } from './login-activity-panel'

export function LoginActivityPageShell(props: LoginActivityPageState) {
  return (
    <PageShell description={<h1 className="text-sm font-semibold">登录活动</h1>}>
      <div className="mx-auto w-full max-w-xl">
        <LoginActivityPanel
          data={props.data}
          loading={props.loading}
          offset={props.offset}
          onOffsetChange={props.setOffset}
        />
      </div>
    </PageShell>
  )
}

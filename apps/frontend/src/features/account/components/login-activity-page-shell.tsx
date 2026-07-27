import type { LoginActivityPageState } from '../hooks/use-login-activity-page'
import { LoginActivityPanel } from './login-activity-panel'

export function LoginActivityPageShell(props: LoginActivityPageState) {
  return (
    <div className="space-y-4">
      <div className="mx-auto w-full max-w-xl">
        <LoginActivityPanel
          data={props.data}
          loading={props.loading}
          offset={props.offset}
          onOffsetChange={props.setOffset}
        />
      </div>
    </div>
  )
}

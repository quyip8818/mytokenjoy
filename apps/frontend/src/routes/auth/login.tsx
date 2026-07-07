import { LoginForm, useLoginPage } from '@/features/auth'

export default function LoginPage() {
  return <LoginForm {...useLoginPage()} />
}

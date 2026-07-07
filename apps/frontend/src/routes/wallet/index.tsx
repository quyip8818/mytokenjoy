import { WalletPageShell, useWalletPage } from '@/features/wallet'

export default function WalletPage() {
  return <WalletPageShell {...useWalletPage()} />
}

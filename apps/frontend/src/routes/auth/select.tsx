import { Building2 } from 'lucide-react'
import { Navigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { useSelectPage } from '@/features/auth'
import { LOGIN_PATH } from '@/config/auth'

export default function SelectCompanyPage() {
  const { companies, selecting, error, handleSelect } = useSelectPage()

  if (companies.length === 0) {
    return <Navigate to={LOGIN_PATH} replace />
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <div className="flex w-full max-w-md flex-col gap-4">
        <h1 className="text-center text-lg font-semibold">选择企业</h1>

        <div className="flex flex-col gap-2">
          {companies.map((company) => (
            <Button
              key={company.companyId}
              variant="outline"
              className="flex h-auto items-center justify-start gap-3 px-4 py-3"
              disabled={selecting !== null}
              onClick={() => handleSelect(company.companyId)}
            >
              <Building2 className="h-5 w-5 shrink-0 text-muted-foreground" />
              <div className="flex flex-col items-start text-left">
                <span className="font-medium">{company.companyName}</span>
                <span className="text-xs text-muted-foreground">{company.role}</span>
              </div>
              {selecting === company.companyId ? (
                <span className="ml-auto text-xs text-muted-foreground">进入中…</span>
              ) : null}
            </Button>
          ))}
        </div>

        {error ? <p className="text-sm text-destructive">{error}</p> : null}
      </div>
    </div>
  )
}

const NEWAPI_URL = import.meta.env.VITE_NEWAPI_URL || 'http://localhost:3000'

export default function NewApiPage() {
  return (
    <iframe
      src={NEWAPI_URL}
      title="NewAPI 管理后台"
      className="h-[calc(100vh-4rem)] w-full border-0"
      allow="clipboard-read; clipboard-write"
    />
  )
}

import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useApis } from '@/api/use-apis'
import { ApiError } from '@/api/client'

export const accountKeys = {
  profile: ['account', 'profile'] as const,
}

export function useAccountPage() {
  const { accountApi, authApi } = useApis()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const profileQuery = useQuery({
    queryKey: accountKeys.profile,
    queryFn: () => accountApi.getProfile(),
  })

  // --- Change Password ---
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false)
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSaving, setPasswordSaving] = useState(false)

  const changePassword = useCallback(
    async (oldPassword: string | undefined, newPassword: string) => {
      setPasswordSaving(true)
      setPasswordError(null)
      try {
        await accountApi.changePassword({ oldPassword, newPassword })
        setPasswordDialogOpen(false)
        queryClient.invalidateQueries({ queryKey: accountKeys.profile })
        return true
      } catch (err) {
        setPasswordError(err instanceof ApiError ? err.message : '操作失败')
        return false
      } finally {
        setPasswordSaving(false)
      }
    },
    [accountApi, queryClient],
  )

  // --- Change Phone ---
  const [phoneDialogOpen, setPhoneDialogOpen] = useState(false)
  const [phoneError, setPhoneError] = useState<string | null>(null)
  const [phoneSaving, setPhoneSaving] = useState(false)

  const changePhone = useCallback(
    async (phone: string, code: string) => {
      setPhoneSaving(true)
      setPhoneError(null)
      try {
        await accountApi.changePhone(phone, code)
        setPhoneDialogOpen(false)
        queryClient.invalidateQueries({ queryKey: accountKeys.profile })
        return true
      } catch (err) {
        setPhoneError(err instanceof ApiError ? err.message : '操作失败')
        return false
      } finally {
        setPhoneSaving(false)
      }
    },
    [accountApi, queryClient],
  )

  // --- Change Email ---
  const [emailDialogOpen, setEmailDialogOpen] = useState(false)
  const [emailError, setEmailError] = useState<string | null>(null)
  const [emailSaving, setEmailSaving] = useState(false)

  const changeEmail = useCallback(
    async (email: string, code: string) => {
      setEmailSaving(true)
      setEmailError(null)
      try {
        await accountApi.changeEmail(email, code)
        setEmailDialogOpen(false)
        queryClient.invalidateQueries({ queryKey: accountKeys.profile })
        return true
      } catch (err) {
        setEmailError(err instanceof ApiError ? err.message : '操作失败')
        return false
      } finally {
        setEmailSaving(false)
      }
    },
    [accountApi, queryClient],
  )

  // --- Revoke Sessions ---
  const [revokeDialogOpen, setRevokeDialogOpen] = useState(false)
  const [revoking, setRevoking] = useState(false)

  const revokeSessions = useCallback(async () => {
    setRevoking(true)
    try {
      await accountApi.revokeSessions()
      setRevokeDialogOpen(false)
      return true
    } catch {
      return false
    } finally {
      setRevoking(false)
    }
  }, [accountApi])

  // --- Logout ---
  const logout = useCallback(async () => {
    await authApi.logout()
    navigate('/login', { replace: true })
  }, [authApi, navigate])

  return {
    profile: profileQuery.data ?? null,
    profileLoading: profileQuery.isLoading,

    passwordDialogOpen,
    setPasswordDialogOpen,
    passwordError,
    passwordSaving,
    changePassword,

    phoneDialogOpen,
    setPhoneDialogOpen,
    phoneError,
    phoneSaving,
    changePhone,

    emailDialogOpen,
    setEmailDialogOpen,
    emailError,
    emailSaving,
    changeEmail,

    revokeDialogOpen,
    setRevokeDialogOpen,
    revoking,
    revokeSessions,

    logout,
  }
}

export type AccountPageState = ReturnType<typeof useAccountPage>

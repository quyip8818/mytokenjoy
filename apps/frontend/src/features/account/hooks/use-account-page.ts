import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useApis } from '@/api/use-apis'
import { ApiError } from '@/api/client'
import { queryKeys } from '@/features/query'

export const accountKeys = {
  profile: ['account', 'profile'] as const,
}

export function useAccountPage() {
  const { meApi, authApi } = useApis()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const profileQuery = useQuery({
    queryKey: accountKeys.profile,
    queryFn: () => meApi.getProfile(),
  })

  // --- Update Profile (name / avatar) ---
  const [profileSaving, setProfileSaving] = useState(false)
  const [profileError, setProfileError] = useState<string | null>(null)

  const updateProfile = useCallback(
    async (params: { name?: string; avatar?: string; alias?: string }) => {
      setProfileSaving(true)
      setProfileError(null)
      try {
        await meApi.updateProfile(params)
        queryClient.invalidateQueries({ queryKey: accountKeys.profile })
        if (params.alias !== undefined || params.avatar !== undefined) {
          queryClient.invalidateQueries({ queryKey: queryKeys.session.all })
        }
        return true
      } catch (err) {
        setProfileError(err instanceof ApiError ? err.message : '保存失败')
        return false
      } finally {
        setProfileSaving(false)
      }
    },
    [meApi, queryClient],
  )

  // --- Change Password ---
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false)
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSaving, setPasswordSaving] = useState(false)

  const changePassword = useCallback(
    async (oldPassword: string | undefined, newPassword: string) => {
      setPasswordSaving(true)
      setPasswordError(null)
      try {
        await meApi.changePassword({ oldPassword, newPassword })
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
    [meApi, queryClient],
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
        await meApi.changePhone(phone, code)
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
    [meApi, queryClient],
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
        await meApi.changeEmail(email, code)
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
    [meApi, queryClient],
  )

  // --- Revoke Sessions ---
  const [revokeDialogOpen, setRevokeDialogOpen] = useState(false)
  const [revoking, setRevoking] = useState(false)

  const revokeSessions = useCallback(async () => {
    setRevoking(true)
    try {
      await meApi.revokeSessions()
      setRevokeDialogOpen(false)
      return true
    } catch {
      return false
    } finally {
      setRevoking(false)
    }
  }, [meApi])

  // --- Logout ---
  const logout = useCallback(async () => {
    await authApi.logout()
    navigate('/login', { replace: true })
  }, [authApi, navigate])

  return {
    profile: profileQuery.data ?? null,
    profileLoading: profileQuery.isLoading,

    profileSaving,
    profileError,
    updateProfile,

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

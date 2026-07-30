import { useCallback, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useApis } from '@/api/use-apis'
import { ApiError } from '@/api/client'
import { queryKeys } from '@/features/query'

export const accountKeys = {
  profile: ['account', 'profile'] as const,
}

function errorMessage(err: Error | null, fallback: string): string | null {
  if (!err) return null
  return err instanceof ApiError ? err.message : fallback
}

export function useAccountPage() {
  const { meApi, authApi } = useApis()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const profileQuery = useQuery({
    queryKey: accountKeys.profile,
    queryFn: () => meApi.getProfile(),
  })

  // --- Update Profile ---
  const profileMutation = useMutation({
    mutationFn: (params: { name?: string; avatar?: string; alias?: string }) =>
      meApi.updateProfile(params),
    onSuccess: (_data, params) => {
      queryClient.invalidateQueries({ queryKey: accountKeys.profile })
      if (params.alias !== undefined || params.avatar !== undefined) {
        queryClient.invalidateQueries({ queryKey: queryKeys.session.all })
      }
    },
  })

  // ponytail: ProfileEditSection 需要 boolean 返回值来控制编辑态关闭。
  // 升级路径：改 ProfileEditSection 为受控组件 + onSuccess 回调。
  const updateProfile = useCallback(
    async (params: { name?: string; avatar?: string; alias?: string }) => {
      try {
        await profileMutation.mutateAsync(params)
        return true
      } catch {
        return false
      }
    },
    [profileMutation],
  )

  // --- Change Password ---
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false)

  const passwordMutation = useMutation({
    mutationFn: ({ oldPassword, newPassword }: { oldPassword?: string; newPassword: string }) =>
      meApi.changePassword({ oldPassword, newPassword }),
    onSuccess: () => {
      setPasswordDialogOpen(false)
      queryClient.invalidateQueries({ queryKey: accountKeys.profile })
    },
  })

  // ponytail: dialog 组件 await onSubmit()，需要 promise 但不需要 return value。
  // mutateAsync throws on error — 用 catch 吞掉让 mutation.error 接管展示。
  const changePassword = useCallback(
    (oldPassword: string | undefined, newPassword: string) =>
      passwordMutation.mutateAsync({ oldPassword, newPassword }).catch(() => {}),
    [passwordMutation],
  )

  // --- Change Phone ---
  const [phoneDialogOpen, setPhoneDialogOpen] = useState(false)

  const phoneMutation = useMutation({
    mutationFn: ({ phone, code }: { phone: string; code: string }) =>
      meApi.changePhone(phone, code),
    onSuccess: () => {
      setPhoneDialogOpen(false)
      queryClient.invalidateQueries({ queryKey: accountKeys.profile })
    },
  })

  const changePhone = useCallback(
    (phone: string, code: string) => phoneMutation.mutateAsync({ phone, code }).catch(() => {}),
    [phoneMutation],
  )

  // --- Change Email ---
  const [emailDialogOpen, setEmailDialogOpen] = useState(false)

  const emailMutation = useMutation({
    mutationFn: ({ email, code }: { email: string; code: string }) =>
      meApi.changeEmail(email, code),
    onSuccess: () => {
      setEmailDialogOpen(false)
      queryClient.invalidateQueries({ queryKey: accountKeys.profile })
    },
  })

  const changeEmail = useCallback(
    (email: string, code: string) => emailMutation.mutateAsync({ email, code }).catch(() => {}),
    [emailMutation],
  )

  // --- Revoke Sessions ---
  const [revokeDialogOpen, setRevokeDialogOpen] = useState(false)

  const revokeMutation = useMutation({
    mutationFn: () => meApi.revokeSessions(),
    onSuccess: () => setRevokeDialogOpen(false),
  })

  // --- Logout ---
  const logout = useCallback(async () => {
    await authApi.logout()
    navigate({ to: '/login', replace: true })
  }, [authApi, navigate])

  return {
    profile: profileQuery.data ?? null,
    profileLoading: profileQuery.isLoading,

    profileSaving: profileMutation.isPending,
    profileError: errorMessage(profileMutation.error, '保存失败'),
    updateProfile,

    passwordDialogOpen,
    setPasswordDialogOpen,
    passwordError: errorMessage(passwordMutation.error, '操作失败'),
    passwordSaving: passwordMutation.isPending,
    changePassword,

    phoneDialogOpen,
    setPhoneDialogOpen,
    phoneError: errorMessage(phoneMutation.error, '操作失败'),
    phoneSaving: phoneMutation.isPending,
    changePhone,

    emailDialogOpen,
    setEmailDialogOpen,
    emailError: errorMessage(emailMutation.error, '操作失败'),
    emailSaving: emailMutation.isPending,
    changeEmail,

    revokeDialogOpen,
    setRevokeDialogOpen,
    revoking: revokeMutation.isPending,
    revokeSessions: () => void revokeMutation.mutate(),

    logout,
  }
}

export type AccountPageState = ReturnType<typeof useAccountPage>

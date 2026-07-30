export interface ProfileCompany {
  companyId: string
  companyName: string
  role: string
  current: boolean
}

export interface Profile {
  phone: string
  email: string
  name: string
  avatar: string
  hasPassword: boolean
  companies: ProfileCompany[]
}

export interface LoginActivityItem {
  time: string
  ip: string
  userAgent: string
  current: boolean
}

export interface LoginActivityResponse {
  items: LoginActivityItem[]
  total: number
}

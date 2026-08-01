import type { FooterContent } from '@/content/types'

export const footerContent: FooterContent = {
  description: '企业级大模型智能管理平台。聚合调度 · 统一掌控,让每一次 AI 调用都令人愉悦。',
  email: 'contact@tokenjoy.me',
  copyright: '© 2026 Tokenjoy. All rights reserved.',
  icpNumber: '浙ICP备2026056029号',
  icpUrl: 'https://beian.miit.gov.cn/',
  columns: [
    {
      title: '产品',
      links: [
        { label: '核心能力', href: '#' },
        { label: '解决方案', href: '#' },
        { label: '更新日志', href: '#' },
        { label: '路线图', href: '#' },
        { label: '价格', href: '#' },
      ],
    },
    {
      title: '资源',
      links: [
        { label: '产品文档', href: '#' },
        { label: 'API 参考', href: '#' },
        { label: '技术博客', href: '#' },
        { label: '客户案例', href: '#' },
        { label: '帮助中心', href: '#' },
      ],
    },
    {
      title: '公司',
      links: [
        { label: '关于我们', href: '#' },
        { label: '加入我们', href: '#' },
        { label: '新闻动态', href: '#' },
        { label: '合作伙伴', href: '#' },
        { label: '联系我们', href: '#' },
      ],
    },
    {
      title: '法律',
      links: [
        { label: '服务条款', href: '#' },
        { label: '隐私政策', href: '#' },
        { label: '数据安全', href: '#' },
        { label: '合规认证', href: '#' },
      ],
    },
  ],
  socialLinks: [
    { icon: 'github', label: 'GitHub', href: '#' },
    { icon: 'twitter', label: 'Twitter', href: '#' },
    { icon: 'linkedin', label: 'LinkedIn', href: '#' },
    { icon: 'wechat', label: '微信', href: '#' },
  ],
}

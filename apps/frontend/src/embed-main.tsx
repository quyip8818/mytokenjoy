import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { AuthCard } from '@/features/auth'

const params = new URLSearchParams(window.location.search)
const mode = params.get('mode') === 'register' ? 'register' : 'login'

// ponytail: embed 只在父窗口 iframe 内运行，postMessage 用 '*'。
// 安全边界在接收方（Navbar 校验 e.origin），不在发送方。
// 升级路径：如需收紧，部署时用 CSP frame-ancestors 限制谁能 iframe 本页。
function postToParent(data: object) {
  if (window.parent !== window) {
    window.parent.postMessage(data, '*')
  }
}

function handleSuccess() {
  if (window.parent !== window) {
    postToParent({ type: 'auth:success' })
  } else {
    window.location.href = '/'
  }
}

const root = document.getElementById('root')!

createRoot(root).render(
  <StrictMode>
    <AuthCard defaultMode={mode} onSuccess={handleSuccess} />
  </StrictMode>,
)

// 持续上报 #root 的内容高度给父窗口，覆盖所有操作（步骤切换、错误提示等）
if (window.parent !== window) {
  const report = () => postToParent({ type: 'auth:resize', height: root.scrollHeight })
  const observer = new MutationObserver(report)
  observer.observe(root, { childList: true, subtree: true, attributes: true })
  // ResizeObserver 兜底布局变化
  new ResizeObserver(report).observe(root)
  // 首次上报
  requestAnimationFrame(report)
}

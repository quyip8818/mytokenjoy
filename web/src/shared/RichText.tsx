import type { ReactNode } from 'react'

const BOLD_PATTERN = /\*\*(.+?)\*\*/g

interface RichTextProps {
  text: string
  className?: string
  boldClassName?: string
}

export function RichText({
  text,
  className,
  boldClassName = 'text-ink-900 font-semibold',
}: RichTextProps) {
  const nodes: ReactNode[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null
  let key = 0

  const pattern = new RegExp(BOLD_PATTERN.source, 'g')
  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) {
      nodes.push(text.slice(lastIndex, match.index))
    }
    nodes.push(
      <strong key={key} className={boldClassName}>
        {match[1]}
      </strong>,
    )
    key += 1
    lastIndex = match.index + match[0].length
  }

  if (lastIndex < text.length) {
    nodes.push(text.slice(lastIndex))
  }

  return <span className={className}>{nodes}</span>
}

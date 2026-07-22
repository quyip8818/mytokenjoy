import { Style, Avatar } from '@dicebear/core'
import adventurer from '@dicebear/styles/adventurer.json' with { type: 'json' }
import notionists from '@dicebear/styles/notionists.json' with { type: 'json' }
import bottts from '@dicebear/styles/bottts.json' with { type: 'json' }
import shapes from '@dicebear/styles/shapes.json' with { type: 'json' }
import lorelei from '@dicebear/styles/lorelei.json' with { type: 'json' }
import funEmoji from '@dicebear/styles/fun-emoji.json' with { type: 'json' }

const styleDefinitions = { adventurer, notionists, bottts, shapes, lorelei, 'fun-emoji': funEmoji }

export type AvatarStyle = keyof typeof styleDefinitions

export const AVATAR_STYLE_NAMES: AvatarStyle[] = [
  'adventurer',
  'notionists',
  'bottts',
  'shapes',
  'lorelei',
  'fun-emoji',
]

/** Render an avatar string to a data URI for display. Returns empty string for no avatar. */
export function renderAvatar(avatar: string): string {
  if (!avatar) return ''
  if (avatar.startsWith('dicebear:')) {
    const [, styleName, seed] = avatar.split(':')
    const def = styleDefinitions[styleName as AvatarStyle]
    if (!def) return ''
    const style = new Style(def)
    return new Avatar(style, { seed }).toDataUri()
  }
  return avatar // data:image/... used directly
}

/** Generate a dicebear avatar string for storage. */
export function dicebearValue(style: AvatarStyle, seed: string): string {
  return `dicebear:${style}:${seed}`
}

/** Generate a random seed string. */
export function randomSeed(): string {
  return Math.random().toString(36).slice(2, 10)
}

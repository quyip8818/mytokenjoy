import { Style, Avatar } from '@dicebear/core'
import avataaars from '@dicebear/styles/avataaars.json' with { type: 'json' }
import lorelei from '@dicebear/styles/lorelei.json' with { type: 'json' }
import micah from '@dicebear/styles/micah.json' with { type: 'json' }
import adventurer from '@dicebear/styles/adventurer.json' with { type: 'json' }
import notionists from '@dicebear/styles/notionists.json' with { type: 'json' }
import bigSmile from '@dicebear/styles/big-smile.json' with { type: 'json' }
import openPeeps from '@dicebear/styles/open-peeps.json' with { type: 'json' }
import pixelArt from '@dicebear/styles/pixel-art.json' with { type: 'json' }
// Legacy styles: not shown in picker but still renderable for existing data
import bottts from '@dicebear/styles/bottts.json' with { type: 'json' }
import shapes from '@dicebear/styles/shapes.json' with { type: 'json' }
import funEmoji from '@dicebear/styles/fun-emoji.json' with { type: 'json' }

/** All renderable styles (includes legacy ones for backward compat). */
const allStyles: Record<string, object> = {
  avataaars,
  lorelei,
  micah,
  adventurer,
  notionists,
  'big-smile': bigSmile,
  'open-peeps': openPeeps,
  'pixel-art': pixelArt,
  // legacy
  bottts,
  shapes,
  'fun-emoji': funEmoji,
}

/** Styles shown in the picker UI. */
const pickerStyles = {
  avataaars,
  lorelei,
  micah,
  adventurer,
  notionists,
  'big-smile': bigSmile,
  'open-peeps': openPeeps,
  'pixel-art': pixelArt,
}

export type AvatarStyle = keyof typeof pickerStyles

export const AVATAR_STYLE_NAMES: AvatarStyle[] = [
  'avataaars',
  'lorelei',
  'micah',
  'adventurer',
  'notionists',
  'big-smile',
  'open-peeps',
  'pixel-art',
]

/** Render an avatar string to a data URI for display. Returns empty string for no avatar. */
export function renderAvatar(avatar: string): string {
  if (!avatar) return ''
  if (avatar.startsWith('dicebear:')) {
    const [, styleName, seed] = avatar.split(':')
    const def = allStyles[styleName]
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

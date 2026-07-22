const MAX_INPUT_SIZE = 5 * 1024 * 1024 // 5MB raw file limit
const ACCEPTED_TYPES = ['image/png', 'image/jpeg', 'image/webp']

export class ImageValidationError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ImageValidationError'
  }
}

/**
 * Compress an image file to a square webp data URI suitable for avatar storage.
 * Center-crops to square, resizes to `size`×`size`, outputs webp at given quality.
 * Returns a data:image/webp;base64,... string.
 */
export async function compressImage(file: File, size = 128, quality = 0.8): Promise<string> {
  if (!ACCEPTED_TYPES.includes(file.type)) {
    throw new ImageValidationError('仅支持 PNG、JPEG、WebP 格式')
  }
  if (file.size > MAX_INPUT_SIZE) {
    throw new ImageValidationError('图片不能超过 5MB')
  }

  const bitmap = await createImageBitmap(file)
  const min = Math.min(bitmap.width, bitmap.height)
  const sx = (bitmap.width - min) / 2
  const sy = (bitmap.height - min) / 2

  const canvas = new OffscreenCanvas(size, size)
  const ctx = canvas.getContext('2d')!
  ctx.drawImage(bitmap, sx, sy, min, min, 0, 0, size, size)
  bitmap.close()

  const blob = await canvas.convertToBlob({ type: 'image/webp', quality })
  const buf = await blob.arrayBuffer()
  const bytes = new Uint8Array(buf)
  // Chunk-based btoa to avoid call stack limits on large arrays
  let binary = ''
  const chunkSize = 8192
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize))
  }
  return `data:image/webp;base64,${btoa(binary)}`
}

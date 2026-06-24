export const EMAIL_DOMAIN = 'example.com'

export const GIVEN_NAMES = [
  '伟',
  '芳',
  '娜',
  '敏',
  '静',
  '丽',
  '强',
  '磊',
  '军',
  '洋',
  '勇',
  '艳',
  '杰',
  '涛',
  '明',
  '超',
  '秀英',
  '霞',
  '平',
  '刚',
  '桂英',
  '建华',
  '文',
  '华',
  '金凤',
  '玉兰',
  '桂兰',
  '志强',
  '秀兰',
  '建国',
]

export const SURNAMES = [
  '王',
  '李',
  '张',
  '刘',
  '陈',
  '杨',
  '赵',
  '黄',
  '周',
  '吴',
  '徐',
  '孙',
  '胡',
  '朱',
  '高',
  '林',
  '何',
  '郭',
  '马',
  '罗',
  '梁',
  '宋',
  '郑',
  '谢',
  '韩',
  '唐',
  '冯',
  '于',
  '董',
  '萧',
]

export function buildChineseName(index: number): string {
  const surname = SURNAMES[index % SURNAMES.length]
  const given = GIVEN_NAMES[Math.floor(index / SURNAMES.length) % GIVEN_NAMES.length]
  return `${surname}${given}`
}

export function buildEmail(_name: string, index: number): string {
  const slug = `user${index}`
  return `${slug}@${EMAIL_DOMAIN}`
}

export function buildPhone(index: number): string {
  const suffix = String(10000000 + index).slice(-8)
  return `13${suffix}`
}

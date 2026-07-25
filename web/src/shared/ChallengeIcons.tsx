export function ChallengeBudgetIcon() {
  return (
    <svg viewBox="0 0 48 48" fill="none" className="w-10 h-10" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="icon1Grad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#A78BFA" />
          <stop offset="100%" stopColor="#7C3AED" />
        </linearGradient>
      </defs>
      <rect x="6" y="26" width="3.5" height="14" rx="1" fill="url(#icon1Grad)" />
      <rect x="14" y="18" width="3.5" height="22" rx="1" fill="url(#icon1Grad)" />
      <rect x="22" y="22" width="3.5" height="18" rx="1" fill="url(#icon1Grad)" />
      <rect x="30" y="10" width="3.5" height="30" rx="1" fill="url(#icon1Grad)" />
      <rect x="38" y="14" width="3.5" height="26" rx="1" fill="url(#icon1Grad)" />
      <path
        d="M 6 24 L 14 18 L 22 22 L 30 10 L 38 14"
        stroke="#7C3AED"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
      <circle cx="38" cy="14" r="2.5" fill="#8B5CF6" stroke="white" strokeWidth="1.5" />
    </svg>
  )
}

export function ChallengeSecurityIcon() {
  return (
    <svg viewBox="0 0 48 48" fill="none" className="w-10 h-10" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="icon2Grad" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#22D3EE" />
          <stop offset="100%" stopColor="#0891B2" />
        </linearGradient>
      </defs>
      <path
        d="M 24 6 L 38 12 L 38 24 C 38 33 32 40 24 43 C 16 40 10 33 10 24 L 10 12 Z"
        fill="url(#icon2Grad)"
      />
      <path
        d="M 18 24 L 22 28 L 30 20"
        stroke="white"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
      />
    </svg>
  )
}

export function ChallengeObservabilityIcon() {
  return (
    <svg viewBox="0 0 48 48" fill="none" className="w-10 h-10" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="icon3Grad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#FB923C" />
          <stop offset="100%" stopColor="#EA580C" />
        </linearGradient>
      </defs>
      <path
        d="M 6 24 C 6 16 14 8 24 8 C 32 8 38 12 42 18 C 38 16 32 16 28 18 C 22 22 22 28 24 32 C 26 36 30 38 36 38 C 32 42 28 44 24 44 C 14 44 6 36 6 28 Z"
        fill="url(#icon3Grad)"
      />
      <circle cx="24" cy="24" r="6" fill="white" />
      <circle cx="24" cy="24" r="3" fill="url(#icon3Grad)" />
      <line x1="8" y1="40" x2="40" y2="8" stroke="white" strokeWidth="3" strokeLinecap="round" />
    </svg>
  )
}

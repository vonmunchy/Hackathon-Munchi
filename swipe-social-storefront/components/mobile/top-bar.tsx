'use client'

import { useRouter } from 'next/navigation'
import { type ReactNode } from 'react'

interface TopBarProps {
  title: string
  showBack?: boolean
  rightAction?: ReactNode
}

export function TopBar({ title, showBack = true, rightAction }: TopBarProps) {
  const router = useRouter()

  return (
    <header
      className="flex shrink-0 items-center bg-white/80 backdrop-blur-xl border-b border-slate-100 px-4"
      style={{
        height: `calc(52px + env(safe-area-inset-top, 0px))`,
        paddingTop: `env(safe-area-inset-top, 0px)`,
      }}
    >
      <div className="flex w-full items-center justify-between">
        <div className="flex items-center gap-3">
          {showBack && (
            <button
              onClick={() => router.back()}
              className="flex h-8 w-8 items-center justify-center rounded-lg text-slate-700 transition-colors hover:bg-slate-100"
              aria-label="Go back"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="h-5 w-5"
              >
                <polyline points="15 18 9 12 15 6" />
              </svg>
            </button>
          )}
          <h1 className="font-display text-[15px] font-semibold tracking-tight text-slate-900">
            {title}
          </h1>
        </div>
        {rightAction && <div>{rightAction}</div>}
      </div>
    </header>
  )
}

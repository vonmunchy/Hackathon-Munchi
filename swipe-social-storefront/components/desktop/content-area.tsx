'use client'

import { type ReactNode } from 'react'

export function ContentArea({ children }: { children: ReactNode }) {
  return (
    <div className="mx-auto max-w-[1200px] px-6 py-6">
      {children}
    </div>
  )
}

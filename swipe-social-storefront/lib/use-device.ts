'use client'
import { useState, useEffect } from 'react'

export function useIsMobile(): boolean | null {
  const [isMobile, setIsMobile] = useState<boolean | null>(null)
  useEffect(() => {
    setIsMobile(window.innerWidth < 768)
    // Do NOT add resize listener — layout stays fixed for the session
  }, [])
  return isMobile
}

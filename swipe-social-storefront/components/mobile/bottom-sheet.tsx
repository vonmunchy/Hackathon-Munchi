'use client'

import { useRef, type ReactNode } from 'react'
import { motion, AnimatePresence, type PanInfo } from 'framer-motion'

interface BottomSheetProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: ReactNode
}

export function BottomSheet({ isOpen, onClose, title, children }: BottomSheetProps) {
  const sheetRef = useRef<HTMLDivElement>(null)

  function handleDragEnd(_: unknown, info: PanInfo) {
    const sheetHeight = sheetRef.current?.offsetHeight ?? 0
    if (info.offset.y > sheetHeight * 0.3) {
      onClose()
    }
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50">
          {/* Backdrop */}
          <motion.div
            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />

          {/* Sheet */}
          <motion.div
            ref={sheetRef}
            className="absolute inset-x-0 bottom-0 max-h-[90vh] overflow-y-auto rounded-t-xl bg-white shadow-sheet"
            initial={{ y: '100%' }}
            animate={{ y: 0 }}
            exit={{ y: '100%' }}
            transition={{ type: 'tween', ease: [0.16, 1, 0.3, 1], duration: 0.35 }}
            drag="y"
            dragConstraints={{ top: 0 }}
            dragElastic={0.2}
            onDragEnd={handleDragEnd}
          >
            {/* Drag handle */}
            <div className="flex justify-center pt-3 pb-2">
              <div className="h-1 w-8 rounded-full bg-slate-300" />
            </div>

            {title && (
              <div className="px-5 pb-3">
                <h2 className="font-display text-lg font-semibold text-slate-900">
                  {title}
                </h2>
              </div>
            )}

            <div className="px-5 pb-6">{children}</div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}

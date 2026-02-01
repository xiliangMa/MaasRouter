'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function Navbar() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center justify-between">
        {/* Brand and mobile menu button */}
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-4">
            <button
              type="button"
              className="md:hidden -ml-2 p-2"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                {mobileMenuOpen ? (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                ) : (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                )}
              </svg>
            </button>
            <Link href="/" className="flex items-center gap-2">
              <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center">
                <span className="font-bold text-primary-foreground">MR</span>
              </div>
              <span className="text-xl font-bold">MassRouter</span>
            </Link>
          </div>
          <nav className="hidden md:flex items-center gap-6 ml-10">
            <Link href="/models" className="text-sm font-medium hover:text-primary transition-colors">
              Models
            </Link>
            <Link href="/pricing" className="text-sm font-medium hover:text-primary transition-colors">
              Pricing
            </Link>
            <Link href="/documentation" className="text-sm font-medium hover:text-primary transition-colors">
              Documentation
            </Link>
            <Link href="/blog" className="text-sm font-medium hover:text-primary transition-colors">
              Blog
            </Link>
          </nav>
        </div>

        {/* Right side actions */}
        <div className="flex items-center gap-4">
          {/* Language selector - hidden on mobile */}
          <div className="hidden sm:block">
            <Select defaultValue="en">
              <SelectTrigger className="w-[120px]">
                <SelectValue placeholder="Language" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="en">English</SelectItem>
                <SelectItem value="zh">中文</SelectItem>
                <SelectItem value="ja">日本語</SelectItem>
                <SelectItem value="ko">한국어</SelectItem>
                <SelectItem value="es">Español</SelectItem>
                <SelectItem value="fr">Français</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Auth buttons */}
          <div className="hidden sm:flex items-center gap-2">
            <Button variant="ghost" asChild>
              <Link href="/login">Login</Link>
            </Button>
            <Button asChild>
              <Link href="/register">Sign Up</Link>
            </Button>
          </div>

          {/* Mobile auth buttons - shown only when menu is open */}
          {mobileMenuOpen && (
            <div className="sm:hidden absolute top-16 left-0 right-0 border-t bg-background p-4 shadow-lg">
              <div className="container">
                <div className="space-y-4">
                  <div className="grid gap-2">
                    <Link href="/models" className="py-2 text-sm font-medium hover:text-primary transition-colors">
                      Models
                    </Link>
                    <Link href="/pricing" className="py-2 text-sm font-medium hover:text-primary transition-colors">
                      Pricing
                    </Link>
                    <Link href="/documentation" className="py-2 text-sm font-medium hover:text-primary transition-colors">
                      Documentation
                    </Link>
                    <Link href="/blog" className="py-2 text-sm font-medium hover:text-primary transition-colors">
                      Blog
                    </Link>
                  </div>
                  <div className="pt-4 border-t">
                    <Select defaultValue="en">
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Language" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="en">English</SelectItem>
                        <SelectItem value="zh">中文</SelectItem>
                        <SelectItem value="ja">日本語</SelectItem>
                        <SelectItem value="ko">한국어</SelectItem>
                        <SelectItem value="es">Español</SelectItem>
                        <SelectItem value="fr">Français</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex flex-col gap-2 pt-4">
                    <Button variant="outline" asChild className="w-full">
                      <Link href="/login">Login</Link>
                    </Button>
                    <Button asChild className="w-full">
                      <Link href="/register">Sign Up</Link>
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
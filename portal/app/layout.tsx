import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { headers } from 'next/headers'
import './globals.css'
import { ThemeProvider } from '@/components/theme-provider'
import { Providers } from './providers'
import Navbar from '@/components/navbar'
import Footer from '@/components/footer'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'MassRouter SaaS - Portal',
  description: 'MassRouter SaaS Platform User Portal',
}

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const headersList = await headers()
  const pathname = headersList.get('x-invoke-path') || '/'
  const isHomePage = pathname === '/'
  console.log('RootLayout pathname:', pathname, 'isHomePage:', isHomePage)
  
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <Providers>
            <div className="min-h-screen flex flex-col">
              {!isHomePage && <Navbar />}
              <main className="flex-1">
                {children}
              </main>
              {!isHomePage && <Footer />}
            </div>
          </Providers>
        </ThemeProvider>
      </body>
    </html>
  )
}
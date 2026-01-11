import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'OAuth Template',
  description: 'Go + Next.js OAuth 2.1 Template',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ja">
      <body className="antialiased">{children}</body>
    </html>
  )
}

import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Ikonex Academy Institutional Portal',
  description: 'High-security operational B2B EdTech SaaS for schools',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-[#edf7f5] text-slate-800 antialiased selection:bg-[#008272] selection:text-white">
        {children}
      </body>
    </html>
  );
}
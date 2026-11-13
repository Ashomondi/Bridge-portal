import './globals.css'; // <-- CRITICAL: This imports your Tailwind CSS styles
import type { Metadata } from 'node_modules/next';

export const metadata: Metadata = {
  title: 'Ikonex Academy Institutional Portal',
  description: 'High-security operational B2B EdTech SaaS for schools',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-[#0d1b1e] text-slate-100 antialiased selection:bg-[#008272] selection:text-white">
        {children}
      </body>
    </html>
  );
}
import type { Metadata, Viewport } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Bridge Portal | School Portal",
  description:
    "Multi-school portal for Kenyan secondary schools, with role-based access for staff, teachers and parents.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  themeColor: "#007A78",
};

/** Renders the document shell shared by every portal route. */
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

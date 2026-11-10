import type { Metadata, Viewport } from "next";
import "./globals.css";
import AssistantWidget from "@/components/assistant/AssistantWidget";

export const metadata: Metadata = {
  title: "Ikonex Academy | School Portal",
  description:
    "Multi-school portal for Kenyan secondary schools, with role-based access for staff, teachers and parents.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  themeColor: "#007A78",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        {children}
        <AssistantWidget />
      </body>
    </html>
  );
}

import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#E8F5F3",
          100: "#D3ECE8",
          200: "#A7DAD4",
          300: "#74C2BB",
          400: "#3BA29C",
          500: "#007A78",
          600: "#006B69",
          700: "#005855",
          800: "#064542",
          900: "#0A3633",
        },
      },
      fontFamily: {
        sans: [
          "Inter",
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "sans-serif",
        ],
      },
      boxShadow: {
        card: "0 1px 2px rgba(15, 23, 42, 0.03), 0 12px 30px -22px rgba(15, 23, 42, 0.24)",
        "card-hover":
          "0 2px 4px rgba(15, 23, 42, 0.04), 0 18px 36px -24px rgba(15, 23, 42, 0.28)",
      },
      keyframes: {
        "drawer-in": {
          "0%": { opacity: "0", transform: "translateX(-100%)" },
          "100%": { opacity: "1", transform: "translateX(0)" },
        },
        "fade-in": {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
      },
      animation: {
        "drawer-in": "drawer-in 220ms ease-out",
        "fade-in": "fade-in 180ms ease-out",
      },
    },
  },
  plugins: [],
};

export default config;

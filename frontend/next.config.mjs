/** @type {import('next').NextConfig} */
function backendOrigin(value) {
  try {
    return new URL(value).origin;
  } catch {
    return 'http://localhost:8080';
  }
}

const backendUrl = backendOrigin(process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080');

const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;

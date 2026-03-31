const basePath = process.env.BASE_PATH || (process.env.NODE_ENV === 'production' ? '/localaihub-admin' : '')

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  basePath,
}

export default nextConfig

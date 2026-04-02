import pkg from './package.json' with { type: 'json' }

const isDev = process.env.NODE_ENV === 'development'
const basePath = isDev ? '' : (process.env.BASE_PATH || '/localaihub-admin')

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'export',
  ...(basePath ? { basePath } : {}),
  images: { unoptimized: true },
  trailingSlash: true,
}

export default nextConfig

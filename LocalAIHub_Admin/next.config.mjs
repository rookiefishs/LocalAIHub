import pkg from './package.json' with { type: 'json' }

const basePath = process.env.BASE_PATH || '/localaihub-admin'

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'export',
  basePath,
  images: { unoptimized: true },
  trailingSlash: true,
}

export default nextConfig
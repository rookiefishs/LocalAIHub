/*
 * @Author: WangZhiYu <w19165802736@163.com>
 * @Date: 2026-03-30 18:06:26
 * @LastEditTime: 2026-04-01 00:27:27
 * @LastEditors: WangZhiYu <w19165802736@163.com>
 * @Descripttion: 
 */
const basePath = process.env.BASE_PATH || (process.env.NODE_ENV === 'production' ? '/localaihub-admin' : '')

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  basePath,
  async rewrites() {
    return {
      beforeFiles: [
        {
          source: '/proxy/:path*',
          destination: 'http://localhost:3334/proxy/:path*',
        },
      ],
    }
  },
}

export default nextConfig

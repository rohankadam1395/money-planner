/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  typescript: {
    strict: true,
  },
  eslint: {
    dirs: ['src'],
  },
};

module.exports = nextConfig;

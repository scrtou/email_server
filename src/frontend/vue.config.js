const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    proxy: {
      '/api/v1': {
        target: 'http://localhost:5555', // Your backend API server
        changeOrigin: true,
        // pathRewrite: { '^/api': '' }, // Uncomment if your backend API doesn't have /api prefix
      },
    },
  },
})

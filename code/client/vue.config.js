
const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
    client: {
      overlay: {
       runtimeErrors: error => {
      const ignoreErrors = [
        'ResizeObserver loop completed with undelivered notifications.'
      ]
      return !ignoreErrors.includes(error.message)
    }
      }
    }
  },
  
})

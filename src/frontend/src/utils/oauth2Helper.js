// OAuth2令牌管理辅助工具
import { checkOAuth2TokenStatus } from './api'
import { ElMessage, ElMessageBox } from 'element-plus'

/**
 * 检查OAuth2令牌状态并处理相应的用户交互
 * @param {number} accountId - 邮箱账户ID
 * @param {boolean} showSuccessMessage - 是否显示成功消息
 * @returns {Promise<{status: string, needsReauth: boolean}>}
 */
export async function checkAndHandleOAuth2Status(accountId, showSuccessMessage = false) {
  try {
    const response = await checkOAuth2TokenStatus(accountId)
    // 令牌状态检查API直接返回状态数据，不使用标准响应格式
    const { status, message, provider, reauth_required } = response

    console.log('🔐 OAuth2令牌状态检查结果:', response)

    switch (status) {
      case 'valid':
        if (showSuccessMessage) {
          ElMessage.success(message || '账户连接正常')
        }
        return { status: 'valid', needsReauth: false }

      case 'expired':
      case 'error':
        if (reauth_required) {
          await showReauthDialog(provider, message)
          return { status, needsReauth: true }
        } else {
          ElMessage.warning(message || '令牌验证失败')
          return { status, needsReauth: false }
        }

      default:
        ElMessage.error('未知的令牌状态')
        return { status: 'unknown', needsReauth: false }
    }
  } catch (error) {
    console.error('🔐 检查OAuth2令牌状态失败:', error)
    ElMessage.error('无法检查账户连接状态，请稍后重试')
    return { status: 'error', needsReauth: false }
  }
}

/**
 * 显示重新授权对话框
 * @param {string} provider - OAuth2提供商名称
 * @param {string} message - 错误消息
 */
async function showReauthDialog(provider, message) {
  const providerNames = {
    google: 'Google',
    microsoft: 'Microsoft',
    outlook: 'Outlook'
  }

  const providerDisplayName = providerNames[provider] || provider

  try {
    await ElMessageBox.confirm(
      `${message || '账户授权已过期'}\n\n是否立即重新连接您的${providerDisplayName}账户？`,
      '需要重新授权',
      {
        confirmButtonText: '重新连接',
        cancelButtonText: '稍后处理',
        type: 'warning',
        center: true
      }
    )

    // 用户确认重新授权，跳转到OAuth2授权页面
    redirectToOAuth2Authorization(provider)
  } catch (error) {
    // 用户取消了重新授权
    console.log('用户取消了重新授权')
  }
}

/**
 * 重定向到OAuth2授权页面
 * @param {string} provider - OAuth2提供商名称
 */
function redirectToOAuth2Authorization(provider) {
  // 构建OAuth2授权URL
  const baseUrl = process.env.VUE_APP_API_BASE_URL || 'http://localhost:5555/api/v1'
  const authUrl = `${baseUrl}/oauth2/connect/${provider}`
  
  // 在新窗口中打开授权页面
  const authWindow = window.open(
    authUrl,
    'oauth2_auth',
    'width=600,height=700,scrollbars=yes,resizable=yes'
  )

  // 监听授权完成
  const checkClosed = setInterval(() => {
    if (authWindow.closed) {
      clearInterval(checkClosed)
      ElMessage.info('授权窗口已关闭，请刷新页面查看最新状态')
      // 可以在这里触发页面刷新或重新检查令牌状态
      setTimeout(() => {
        window.location.reload()
      }, 2000)
    }
  }, 1000)
}

/**
 * 处理邮件API错误，特别是OAuth2相关错误
 * @param {Error} error - API错误对象
 * @param {number} accountId - 邮箱账户ID
 * @returns {Promise<boolean>} - 是否已处理错误
 */
export async function handleEmailAPIError(error, accountId) {
  // 处理OAuth2重新授权需求错误
  if (error.isReauthRequired) {
    console.log('🔐 检测到OAuth2重新授权需求:', error)

    await showReauthDialog(error.provider, error.message)
    return true // 已处理错误
  }

  if (!error.response) {
    return false
  }

  const { status, data } = error.response

  // 处理401认证错误
  if (status === 401) {
    const message = data?.message || ''

    // 检查是否是OAuth2相关错误
    if (message.includes('Google账户') || message.includes('Microsoft账户') ||
        message.includes('授权已过期') || message.includes('重新连接')) {

      console.log('🔐 检测到OAuth2认证错误，检查令牌状态...')

      // 自动检查令牌状态并处理
      const result = await checkAndHandleOAuth2Status(accountId)

      if (result.needsReauth) {
        ElMessage.warning('请完成账户重新授权后再试')
        return true // 已处理错误
      }
    }
  }

  return false // 未处理错误，让调用方继续处理
}

/**
 * 批量检查多个账户的OAuth2令牌状态
 * @param {Array<number>} accountIds - 邮箱账户ID数组
 * @returns {Promise<Array<{accountId: number, status: string, needsReauth: boolean}>>}
 */
export async function batchCheckOAuth2Status(accountIds) {
  const results = []
  
  for (const accountId of accountIds) {
    try {
      const result = await checkAndHandleOAuth2Status(accountId, false)
      results.push({
        accountId,
        ...result
      })
    } catch (error) {
      console.error(`检查账户 ${accountId} 的OAuth2状态失败:`, error)
      results.push({
        accountId,
        status: 'error',
        needsReauth: false
      })
    }
  }
  
  return results
}

/**
 * 定期检查OAuth2令牌状态（用于后台监控）
 * @param {Array<number>} accountIds - 需要监控的账户ID数组
 * @param {number} intervalMinutes - 检查间隔（分钟）
 * @returns {Function} - 停止监控的函数
 */
export function startOAuth2StatusMonitoring(accountIds, intervalMinutes = 30) {
  console.log(`🔐 开始OAuth2状态监控，检查间隔: ${intervalMinutes}分钟`)
  
  const intervalId = setInterval(async () => {
    console.log('🔐 执行定期OAuth2状态检查...')
    const results = await batchCheckOAuth2Status(accountIds)
    
    const expiredAccounts = results.filter(r => r.needsReauth)
    if (expiredAccounts.length > 0) {
      console.warn('🔐 发现过期的OAuth2令牌:', expiredAccounts)
      ElMessage.warning(`发现 ${expiredAccounts.length} 个账户需要重新授权`)
    }
  }, intervalMinutes * 60 * 1000)
  
  // 返回停止监控的函数
  return () => {
    clearInterval(intervalId)
    console.log('🔐 OAuth2状态监控已停止')
  }
}

export default {
  checkAndHandleOAuth2Status,
  handleEmailAPIError,
  batchCheckOAuth2Status,
  startOAuth2StatusMonitoring
}

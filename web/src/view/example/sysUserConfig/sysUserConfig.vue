<template>
  <div class="config-page">
    <el-tabs v-model="activeTab" type="card" class="config-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="基础配置" name="site">
        <el-card class="config-card">
          <el-form :model="formState" ref="elFormRef" :rules="formRules" label-position="top" class="config-form">
            <el-form-item prop="allow_request_url" :required="true">
              <template #label>
                <span>允许请求地址</span>
                <el-tooltip placement="top" content="若不设置则默认允许所有请求; 支持填入 IP 或域名其中一种; 多个地址用英文分号';'分隔">
                  <el-icon class="label-tip-icon"><QuestionFilled /></el-icon>
                </el-tooltip>
              </template>
              <el-input v-model="formState.allow_request_url" 
                type="textarea"
                rows="2"
                placeholder="支持填入 IP 或域名其中一种(有多个地址时,使用';'隔开)" />
              <div class="form-tip">示例：example.com;www.example.com</div>
            </el-form-item>
            <div class="config-footer">
              <el-button type="primary" :loading="btnLoading" @click="saveConfig">保 存</el-button>
            </div>
          </el-form>
        </el-card>
      </el-tab-pane>
      
      <el-tab-pane label="Api请求Token管理" name="token">
        <el-card class="config-card">
          <div class="token-header">
            <el-button type="primary" @click="generateToken" :loading="generateLoading">
              生成新Token
            </el-button>
            <el-button @click="loadTokens" :loading="loadTokensLoading">
              刷新列表
            </el-button>
          </div>
          
          <el-table :data="tokenList" v-loading="loadTokensLoading" class="token-table">
            <el-table-column label="Token" min-width="200">
              <template #default="{ row }">
                <div class="token-cell">
                  <span v-if="row.showToken">{{ row.token }}</span>
                  <span v-else>{{ maskToken(row.token) }}</span>
                  <el-button 
                    link 
                    type="primary" 
                    @click="toggleTokenVisibility(row)"
                    class="token-toggle"
                  >
                    {{ row.showToken ? '隐藏' : '显示' }}
                  </el-button>
                  <el-button 
                    link 
                    type="primary" 
                    @click="copyToken(row.token)"
                    class="token-copy"
                  >
                    复制
                  </el-button>
                </div>
              </template>
            </el-table-column>
            
            <el-table-column label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatTimestamp(row.created_at) }}
              </template>
            </el-table-column>
            
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.is_active ? 'success' : 'danger'">
                  {{ row.is_active ? '激活' : '已撤销' }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button 
                  v-if="row.is_active"
                  type="danger" 
                  size="small" 
                  @click="revokeToken(row)"
                  :loading="row.revoking"
                >
                  撤销
                </el-button>
                <span v-else class="disabled-text">已撤销</span>
              </template>
            </el-table-column>
          </el-table>
          
          <el-empty v-if="!loadTokensLoading && tokenList.length === 0" description="暂无永久Token" />
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
  
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { updateSysUserConfig, getSysUserConfigList } from '@/api/example/sysUserConfig'
import { generatePermanentToken, getPermanentTokens, revokePermanentToken } from '@/api/example/merUser'
import { QuestionFilled } from '@element-plus/icons-vue'

defineOptions({ name: 'SysUserConfig' })

const activeTab = ref('site')
const elFormRef = ref()
const btnLoading = ref(false)

// 根据用户要求：本页 formId 默认设为 1，可修改
const formId = ref(1)

// 永久token管理相关状态
const tokenList = ref([])
const loadTokensLoading = ref(false)
const generateLoading = ref(false)

// 双字段配置状态，映射到返回的 data.list
const formState = reactive({
  allow_request_url: '',
  encrypt_key: ''
})

// IP地址格式验证函数
const isValidIP = (ip) => {
  // 严格的IPv4地址验证
  const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
  const trimmedIp = ip.trim()
  
  // 首先检查是否包含非法字符（如端口号、特殊字符等）
  if (!/^[\d.]+$/.test(trimmedIp)) {
    return false
  }
  
  // 检查是否符合IP格式
  if (!ipRegex.test(trimmedIp)) {
    return false
  }
  
  // 额外验证：确保没有前导零（除了0本身）
  const parts = trimmedIp.split('.')
  for (const part of parts) {
    if (part.length > 1 && part.startsWith('0')) {
      return false
    }
  }
  
  return true
}

// 域名格式验证函数
const isValidDomain = (domain) => {
  const trimmedDomain = domain.trim()
  
  // 检查长度限制
  if (trimmedDomain.length === 0 || trimmedDomain.length > 253) {
    return false
  }
  
  // 检查是否包含非法字符（端口号、特殊字符等）
  if (!/^[a-zA-Z0-9.-]+$/.test(trimmedDomain)) {
    return false
  }
  
  // 不能以点开头或结尾
  if (trimmedDomain.startsWith('.') || trimmedDomain.endsWith('.')) {
    return false
  }
  
  // 不能包含连续的点
  if (trimmedDomain.includes('..')) {
    return false
  }
  
  // 分割域名各部分进行验证
  const labels = trimmedDomain.split('.')
  
  // 至少要有一个标签
  if (labels.length === 0) {
    return false
  }
  
  // 验证每个标签
  for (const label of labels) {
    // 标签长度限制
    if (label.length === 0 || label.length > 63) {
      return false
    }
    
    // 标签不能以连字符开头或结尾
    if (label.startsWith('-') || label.endsWith('-')) {
      return false
    }
    
    // 标签必须包含字母或数字
    if (!/^[a-zA-Z0-9-]+$/.test(label)) {
      return false
    }
  }
  
  // 顶级域名必须包含至少一个字母
  const tld = labels[labels.length - 1]
  if (!/[a-zA-Z]/.test(tld)) {
    return false
  }
  
  return true
}

// 检测地址类型并返回详细错误信息
const getAddressValidationError = (address) => {
  const trimmed = address.trim()
  
  if (trimmed === '') {
    return '地址不能为空'
  }
  
  // 检查是否包含端口号
  if (/:\d+/.test(trimmed)) {
    return '地址不应包含端口号，请只输入IP地址或域名'
  }
  
  // 检查是否包含协议
  if (/^https?:\/\//.test(trimmed)) {
    return '地址不应包含协议（http://或https://），请只输入IP地址或域名'
  }
  
  // 检查是否包含路径
  if (trimmed.includes('/')) {
    return '地址不应包含路径，请只输入IP地址或域名'
  }
  
  // 检查是否可能是IP地址
  if (/^\d+\.\d+\.\d+\.\d+/.test(trimmed)) {
    if (!isValidIP(trimmed)) {
      return 'IP地址格式不正确，请输入有效的IPv4地址（如：192.168.1.1）'
    }
  } else {
    // 检查域名
    if (!isValidDomain(trimmed)) {
      return '域名格式不正确，请输入有效的域名（如：example.com）'
    }
  }
  
  return null
}

// 允许请求地址验证函数
const validateAllowRequestUrl = (rule, value, callback) => {
  if (!value || value.trim() === '') {
    // 允许为空，表示允许所有请求
    callback()
    return
  }

  const addresses = value.split(';').map(addr => addr.trim()).filter(addr => addr !== '')
  
  for (const address of addresses) {
    const error = getAddressValidationError(address)
    if (error) {
      callback(new Error(`地址 "${address}" ${error}`))
      return
    }
  }
  
  callback()
}

// 表单验证规则
const formRules = reactive({
  allow_request_url: [
    { validator: validateAllowRequestUrl, trigger: 'blur' }
  ],
  encrypt_key: [
    { required: true, message: '请输入加密密钥', trigger: 'blur' },
    { min: 1, message: '加密密钥不能为空', trigger: 'blur' }
  ]
})

// 将后端返回的 list 映射到本地状态
const mapListToState = (list = []) => {
  list.forEach(item => {
    if (!item || typeof item !== 'object') return
    if (item.name === 'allow_request_url') {
      formState.allow_request_url = item.value ?? ''
    } else if (item.name === 'encrypt_key') {
      formState.encrypt_key = item.value ?? ''
    }
  })
}

// 加载配置：使用项目内置接口（分页列表），按 formId 过滤
const loadConfig = async () => {
  try {
    const { code, data, msg } = await getSysUserConfigList({ page: 1, pageSize: 100, formId: formId.value })
    if (code === 0 && data && Array.isArray(data.list)) {
      mapListToState(data.list)
      // ElMessage.success('加载成功')
    } else {
      ElMessage.error(msg || '加载失败')
    }
  } catch (e) {
    ElMessage.error('加载异常：' + (e?.message || e))
  }
}

// 保存配置：使用项目内置接口，逐项创建或更新
const saveConfig = async () => {
  // 先进行表单验证
  if (!elFormRef.value) return
  
  try {
    await elFormRef.value.validate()
  } catch (error) {
    ElMessage.error('请检查输入格式')
    return
  }

  btnLoading.value = true
  try {
    // 读取当前 formId 下的记录，以决定是更新还是创建
    const { code, data } = await getSysUserConfigList({ page: 1, pageSize: 100, formId: formId.value })
    const existing = code === 0 && data && Array.isArray(data.list) ? data.list : []
    const byName = Object.create(null)
    existing.forEach(item => {
      if (item && item.name) byName[item.name] = item
    })

    const tasks = []
    const items = [
      { name: 'allow_request_url', title: 'allow_request_url', formId: formId.value, value: formState.allow_request_url },
      { name: 'encrypt_key', title: 'encrypt_key', formId: formId.value, value: formState.encrypt_key }
    ]

    const missingNames = []
    items.forEach(it => {
              tasks.push(updateSysUserConfig({...it }))

    })
    console.log(missingNames);
    
    if (missingNames.length) {
      ElMessage.warning(`缺少配置项：${missingNames.join(', ')}，已跳过缺失项，仅更新已存在项`)
    }

    if (tasks.length === 0) {
      // 全部缺失，无可更新项
      return
    }

    const results = await Promise.all(tasks)
    const ok = results.every(r => r && r.code === 0)
    if (ok) {
      ElMessage.success('保存成功')
      await loadConfig()
    } else {
      ElMessage.error('部分保存失败')
    }
  } catch (e) {
    ElMessage.error('保存异常：' + (e?.message || e))
  } finally {
    btnLoading.value = false
  }
}

// 永久token管理功能
const generateToken = async () => {
  try {
    generateLoading.value = true
    const res = await generatePermanentToken()
    if (res.code === 0) {
      ElMessage.success('永久token生成成功')
      await loadTokens()
    } else {
      ElMessage.error(res.msg || '生成token失败')
    }
  } catch (error) {
    console.error('生成token失败:', error)
    ElMessage.error('生成token失败')
  } finally {
    generateLoading.value = false
  }
}

const loadTokens = async () => {
  try {
    loadTokensLoading.value = true
    const res = await getPermanentTokens()
    if (res.code === 0) {
      tokenList.value = (res.data.tokens || []).map(token => ({
        ...token,
        showToken: false,
        revoking: false
      }))
    } else {
      ElMessage.error(res.msg || '获取token列表失败')
    }
  } catch (error) {
    console.error('获取token列表失败:', error)
    ElMessage.error('获取token列表失败')
  } finally {
    loadTokensLoading.value = false
  }
}

const revokeToken = async (tokenRow) => {
  try {
    await ElMessageBox.confirm('确定要撤销此token吗？撤销后将无法恢复。', '确认撤销', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    tokenRow.revoking = true
    const res = await revokePermanentToken({ token: tokenRow.token })
    if (res.code === 0) {
      ElMessage.success('token撤销成功')
      await loadTokens()
    } else {
      ElMessage.error(res.msg || '撤销token失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('撤销token失败:', error)
      ElMessage.error('撤销token失败')
    }
  } finally {
    tokenRow.revoking = false
  }
}

const copyToken = (token) => {
  navigator.clipboard.writeText(token).then(() => {
    ElMessage.success('token已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const toggleTokenVisibility = (tokenRow) => {
  tokenRow.showToken = !tokenRow.showToken
}

const maskToken = (token) => {
  if (!token || token.length <= 8) return '****'
  return token.substring(0, 4) + '****' + token.substring(token.length - 4)
}

const formatTimestamp = (timestamp) => {
  return new Date(timestamp * 1000).toLocaleString()
}

// 监听tab切换
const handleTabChange = (tabName) => {
  if (tabName === 'token') {
    loadTokens()
  }
}

onMounted(() => {
  loadConfig()
  if (activeTab.value === 'token') {
    loadTokens()
  }
})

</script>

<style>
.config-page { padding: 16px; }
.config-tabs { background: var(--el-bg-color-overlay); }
.config-card { background: var(--el-bg-color); }
.config-header { display: flex; justify-content: flex-end; margin-bottom: 12px; }
.config-actions { display: flex; align-items: center; }
.config-form { padding: 12px; }
.config-footer { margin-top: 16px; }
.label-tip-icon { margin-left: 6px; cursor: help; color: var(--el-text-color-secondary); }
.form-tip { margin-top: 6px; font-size: 12px; color: var(--el-text-color-secondary); }
</style>
<template>
  <div class="config-page">
    <el-tabs v-model="activeTab" type="card" class="config-tabs">
      <el-tab-pane label="基础配置" name="site">
        <el-card class="config-card">
          <!-- <div class="config-header">
            <div class="config-actions">
              <el-input v-model="token" placeholder="可选：鉴权 Token" clearable style="max-width: 280px" />
              <el-input-number v-model="formId" :min="1" :max="9999" controls-position="right" style="margin-left: 12px" />
              <el-button style="margin-left: 12px" icon="refresh" @click="loadConfig">加载</el-button>
            </div>
          </div> -->
          <el-form :model="formState" ref="elFormRef" label-position="top" class="config-form">
            <el-form-item :required="true">
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
            <el-form-item label="加密密钥" :required="true">
              <el-input v-model="formState.encrypt_key" rows="2" type="textarea" placeholder="请输入加密密钥" />
            </el-form-item>
            <div class="config-footer">
              <el-button type="primary" :loading="btnLoading" @click="saveConfig">保 存</el-button>
            </div>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
  
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { updateSysUserConfig, getSysUserConfigList } from '@/api/example/sysUserConfig'
import { QuestionFilled } from '@element-plus/icons-vue'

defineOptions({ name: 'SysUserConfig' })

const activeTab = ref('site')
const elFormRef = ref()
const btnLoading = ref(false)

// 根据用户要求：本页 formId 默认设为 1，可修改
const formId = ref(1)

// 双字段配置状态，映射到返回的 data.list
const formState = reactive({
  allow_request_url: '',
  encrypt_key: ''
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
  // 轻量校验：两项必填

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

onMounted(() => {
  loadConfig()
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
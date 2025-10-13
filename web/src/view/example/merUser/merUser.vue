
<template>
  <div>
    <div class="gva-search-box">
      <el-form ref="elSearchFormRef" :inline="true" :model="searchInfo" class="demo-form-inline" @keyup.enter="onSubmit">

        <template v-if="showAllQuery">
          <!-- 将需要控制显示状态的查询条件添加到此范围内 -->
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
          <el-button link type="primary" icon="arrow-down" @click="showAllQuery=true" v-if="!showAllQuery">展开</el-button>
          <el-button link type="primary" icon="arrow-up" @click="showAllQuery=false" v-else>收起</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
        <div class="gva-btn-list">
            <el-button v-auth="btnAuth.add" type="primary" icon="plus" @click="openDialog()">新增</el-button>
            <el-button v-auth="btnAuth.batchDelete" icon="delete" style="margin-left: 10px;" :disabled="!multipleSelection.length" @click="onDelete">删除</el-button>
            
        </div>
        <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
        >
        <el-table-column type="selection" width="55" />
        
            <el-table-column align="left" label="id字段" prop="id" width="120" />

            <el-table-column align="left" label="接入类型" prop="merType" width="120">
              <template #default="scope">
                {{ merTypeLabel(scope.row.merType) }}
              </template>
            </el-table-column>

            <el-table-column align="left" label="账号" prop="userName" width="120" />

            <el-table-column align="left" label="密码" prop="password" width="160">
              <template #default="scope">
                <span>{{ isRowPwdVisible(scope.row) ? scope.row.password : '••••••' }}</span>
                <el-button link type="primary" size="small" @click="toggleRowPwd(scope.row)">
                  {{ isRowPwdVisible(scope.row) ? '隐藏' : '显示' }}
                </el-button>
              </template>
            </el-table-column>

            
            <el-table-column align="left" label="收款码" prop="qrCode" width="140">
              <template #default="scope">
                <el-image
                  v-if="scope.row.qrCode"
                  :src="getQrSrc(scope.row.qrCode)"
                  style="width: 80px; height: 80px"
                  fit="contain"
                  
                />
                <span v-else class="text-gray-400">-</span>
              </template>
            </el-table-column>

<!--            <el-table-column align="left" label="请求密钥" prop="key" width="120" />-->

            <!-- <el-table-column align="left" label="是否删除(1: 删除 0:未删除)" prop="isDel" width="120" /> -->

            <el-table-column align="left" label="创建时间" prop="createTime" width="180">
   <template #default="scope">{{ formatDate(scope.row.createTime) }}</template>
</el-table-column>
            <el-table-column align="left" label="更新时间" prop="updateTime" width="180">
   <template #default="scope">{{ formatDate(scope.row.updateTime) }}</template>
</el-table-column>
            <el-table-column align="left" label="备注" prop="remarks" width="120" />
        <!-- 末尾添加“是否启用”开关列，无需打开详情即可切换 -->
        <el-table-column align="center" label="是否启用" fixed="right" width="120">
          <template #default="scope">
            <el-switch
              :loading="togglingId === scope.row.id"
              v-model="scope.row.state"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
              @change="onToggleState(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" fixed="right" :min-width="appStore.operateMinWith">
            <template #default="scope">
            <el-button v-auth="btnAuth.info" type="primary" link class="table-button" @click="getDetails(scope.row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看</el-button>
            <el-button v-auth="btnAuth.edit" type="primary" link icon="edit" class="table-button" @click="updateMerUserFunc(scope.row)">编辑</el-button>
            <el-button  v-auth="btnAuth.delete" type="primary" link icon="delete" @click="deleteRow(scope.row)">删除</el-button>
            </template>
        </el-table-column>
        
        </el-table>
        <div class="gva-pagination">
            <el-pagination
            layout="total, sizes, prev, pager, next, jumper"
            :current-page="page"
            :page-size="pageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="total"
            @current-change="handleCurrentChange"
            @size-change="handleSizeChange"
            />
        </div>
    </div>
    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="dialogFormVisible" :show-close="false" :before-close="closeDialog">
       <template #header>
              <div class="flex justify-between items-center">
                <span class="text-lg">{{type==='create'?'新增':'编辑'}}</span>
                <div>
                  <el-button :loading="btnLoading" type="primary" @click="enterDialog">确 定</el-button>
                  <el-button @click="closeDialog">取 消</el-button>
                </div>
              </div>
            </template>

          <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
            <el-form-item label="接入类型:" prop="merType">
              <el-select v-model="formData.merType" clearable placeholder="请选择接入类型">
                <el-option
                  v-for="opt in merTypeOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="账号:" prop="userName">
    <el-input v-model="formData.userName" :clearable="true" placeholder="请输入账号" />
</el-form-item>
            <el-form-item label="密码:" prop="password">
    <el-input v-model="formData.password" :clearable="true" placeholder="请输入密码" type="password" show-password />
</el-form-item>
            <el-form-item label="是否启用:" prop="state">
    <el-switch v-model="formData.state" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
            <el-form-item label="收款码:" prop="qrCode">
              <div class="flex items-center gap-3">
                <el-upload
                  drag
                  :show-file-list="false"
                  :auto-upload="false"
                  accept="image/*"
                  :on-change="onQrFileChange"
                  class="qr-upload-area"
                >
                  <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
                  <div class="el-upload__text">拖拽或<em>点击上传</em></div>
                  <template #tip>
                    <!-- <div class="el-upload__tip">支持图片文件，自动转换为Base64</div> -->
                  </template>
                </el-upload>
                <el-image
                  v-if="formData.qrCode"
                  :src="getQrSrc(formData.qrCode)"
                  style="width: 80px; height: 80px"
                  fit="contain"
                
                />
                <el-button v-if="formData.qrCode" link type="danger" @click="clearQrCode">清除</el-button>
              </div>
            </el-form-item>
<!--            <el-form-item label="请求密钥:" prop="key">-->
<!--    <el-input v-model="formData.key" :clearable="true" placeholder="请输入请求密钥" />-->
<!--</el-form-item>-->
            <el-form-item label="备注:" prop="remarks">
    <el-input v-model="formData.remarks" :clearable="true" placeholder="请输入备注" />
</el-form-item>
          </el-form>
    </el-drawer>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="detailShow" :show-close="true" :before-close="closeDetailShow" title="查看">
            <el-descriptions :column="1" border>
                    <el-descriptions-item label="id字段">
    {{ detailForm.id }}
</el-descriptions-item>
                    <el-descriptions-item label="接入类型">
    {{ merTypeLabel(detailForm.merType) }}
</el-descriptions-item>
                    <el-descriptions-item label="账号">
    {{ detailForm.userName }}
</el-descriptions-item>
                    <el-descriptions-item label="密码">
                      <span>{{ showDetailPwd ? detailForm.password : '••••••' }}</span>
                      <el-button link type="primary" size="small" @click="showDetailPwd = !showDetailPwd">
                        {{ showDetailPwd ? '隐藏' : '显示' }}
                      </el-button>
                    </el-descriptions-item>
                    <el-descriptions-item label="是否启用">
    {{ detailForm.state }}
</el-descriptions-item>
                    <el-descriptions-item label="收款码">
                      <el-image
                        v-if="detailForm.qrCode"
                        :src="getQrSrc(detailForm.qrCode)"
                        style="width: 120px; height: 120px"
                        fit="contain"
              
                      />
                      <span v-else class="text-gray-400">-</span>
                    </el-descriptions-item>
<!--                    <el-descriptions-item label="请求密钥">-->
<!--    {{ detailForm.key }}-->
<!--</el-descriptions-item>-->
<!--                    <el-descriptions-item label="是否删除(1: 删除 0:未删除)">-->
<!--    {{ detailForm.isDel }}-->
<!--</el-descriptions-item>-->
                    <el-descriptions-item label="创建时间">
    {{ detailForm.createTime }}
</el-descriptions-item>
                    <el-descriptions-item label="更新时间">
    {{ detailForm.updateTime }}
</el-descriptions-item>
                    <el-descriptions-item label="备注">
    {{ detailForm.remarks }}
</el-descriptions-item>
            </el-descriptions>
        </el-drawer>

  </div>
</template>

<script setup>
import {
  createMerUser,
  deleteMerUser,
  deleteMerUserByIds,
  updateMerUser,
  findMerUser,
  getMerUserList
} from '@/api/example/merUser'

// 全量引入格式化工具 请按需保留
import { getDictFunc, formatDate, formatBoolean, filterDict ,filterDataSource, returnArrImg, onDownloadFile } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'
// 引入按钮权限标识
import { useBtnAuth } from '@/utils/btnAuth'
import { useAppStore } from "@/pinia"
import { UploadFilled } from '@element-plus/icons-vue'




defineOptions({
    name: 'MerUser'
})
// 按钮权限实例化
    const btnAuth = useBtnAuth()

// 提交按钮loading
const btnLoading = ref(false)
const appStore = useAppStore()

// 控制更多查询条件显示/隐藏状态
const showAllQuery = ref(false)

// 自动化生成的字典（可能为空）以及字段
const formData = ref({
            merType: '',
            userName: '',
            password: '',
            state: false,
            qrCode: '',
            key: '',
            remarks: '',
        })

// 接入类型选项与映射（0: 星驿，1: 富掌柜）
const merTypeOptions = [
  { label: '星驿', value: '0' },
  { label: '富掌柜', value: '1' }
]
const merTypeLabel = (val) => {
  const code = val == null ? '' : String(val)
  const found = merTypeOptions.find(o => o.value === code)
  return found ? found.label : code
}



// 验证规则
const rule = reactive({
})

const elFormRef = ref()
const elSearchFormRef = ref()

// =========== 表格控制部分 ===========
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})
// 重置
const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

// 搜索
const onSubmit = () => {
  elSearchFormRef.value?.validate(async(valid) => {
    if (!valid) return
    page.value = 1
    if (searchInfo.value.state === ""){
        searchInfo.value.state=null
    }
    getTableData()
  })
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

// 修改页面容量
const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

// 查询
const getTableData = async() => {
  const table = await getMerUserList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
    // 刷新列表时，重置表格中的密码显示状态为隐藏
    showPwdMap.value = {}
  }
}

getTableData()

// ============== 表格控制部分结束 ===============

// 获取需要的字典 可能为空 按需保留
const setOptions = async () =>{
}

// 获取需要的字典 可能为空 按需保留
setOptions()


// 多选数据
const multipleSelection = ref([])
// 多选
const handleSelectionChange = (val) => {
    multipleSelection.value = val
}

// 删除行
const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
            deleteMerUserFunc(row)
        })
    }

// 多选删除
const onDelete = async() => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async() => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map(item => {
          ids.push(item.id)
        })
      const res = await deleteMerUserByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
      })
    }

// 行为控制标记（弹窗内部需要增还是改）
const type = ref('')

// 更新行
const updateMerUserFunc = async(row) => {
    const res = await findMerUser({ id: row.id })
    type.value = 'update'
    if (res.code === 0) {
        // 将后端返回的类型值统一为字符串（"0"/"1"）以适配下拉框
        const data = { ...res.data }
        data.merType = data.merType == null ? '' : String(data.merType)
        formData.value = data
        dialogFormVisible.value = true
    }
}


// 删除行
const deleteMerUserFunc = async (row) => {
    const res = await deleteMerUser({ id: row.id })
    if (res.code === 0) {
        ElMessage({
                type: 'success',
                message: '删除成功'
            })
            if (tableData.value.length === 1 && page.value > 1) {
            page.value--
        }
        getTableData()
    }
}

// 弹窗控制标记
const dialogFormVisible = ref(false)

// 打开弹窗
const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
}

// 关闭弹窗
const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
        merType: '',
        userName: '',
        password: '',
        state: false,
        qrCode: '',
        key: '',
        remarks: '',
        }
}
// 弹窗确定
const enterDialog = async () => {
     btnLoading.value = true
     elFormRef.value?.validate( async (valid) => {
             if (!valid) return btnLoading.value = false
              let res
              // 仅向后端发送纯 base64 字符串
              const payload = {
                ...formData.value,
                qrCode: toBase64Raw(formData.value.qrCode)
              }
              // 保证提交的接入类型为代码字符串 "0"/"1"
              if (payload.merType != null) {
                payload.merType = String(payload.merType)
              }
              switch (type.value) {
                case 'create':
                  res = await createMerUser(payload)
                  break
                case 'update':
                  res = await updateMerUser(payload)
                  break
                default:
                  res = await createMerUser(payload)
                  break
              }
              btnLoading.value = false
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: '创建/更改成功'
                })
                closeDialog()
                getTableData()
              }
      })
}

const detailForm = ref({})

// 查看详情控制标记
const detailShow = ref(false)


// 打开详情弹窗
const openDetailShow = () => {
  detailShow.value = true
}


// 打开详情
const getDetails = async (row) => {
  // 打开弹窗
  const res = await findMerUser({ id: row.id })
  if (res.code === 0) {
    detailForm.value = res.data
    openDetailShow()
  }
}


// 关闭详情弹窗
const closeDetailShow = () => {
  detailShow.value = false
  detailForm.value = {}
  // 关闭详情时自动隐藏完整密码
  showDetailPwd.value = false
}

// ============== 收款码上传与显示 ===============
// 显示：将字符串标准化为 data URL
const getQrSrc = (val) => {
  if (!val || typeof val !== 'string') return ''
  return val.startsWith('data:') ? val : `data:image/png;base64,${val}`
}

// 提交：仅保留纯 base64 内容
const toBase64Raw = (val) => {
  if (!val || typeof val !== 'string') return ''
  if (val.startsWith('data:')) {
    const idx = val.indexOf(',')
    return idx > -1 ? val.slice(idx + 1) : val
  }
  return val
}

// 处理上传图片 -> base64
const onQrFileChange = (uploadFile) => {
  const file = uploadFile?.raw || uploadFile
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    formData.value.qrCode = e.target.result
  }
  reader.readAsDataURL(file)
}

const clearQrCode = () => {
  formData.value.qrCode = ''
}

// ============== 行内切换“是否启用” ===============
const togglingId = ref(null)
const onToggleState = async (row) => {
  const original = !row.state // 因为 v-model 已经改为当前值
  try {
    togglingId.value = row.id
    const res = await updateMerUser({
      id: row.id,
      merType: row.merType,
      userName: row.userName,
      password: row.password,
      state: row.state,
      qrCode: row.qrCode,
      key: row.key,
      remarks: row.remarks,
    })
    if (res.code === 0) {
      ElMessage.success(row.state ? '已启用' : '已禁用')
    } else {
      row.state = original
      ElMessage.error(res.msg || '更新失败')
    }
  } catch (e) {
    row.state = original
    ElMessage.error(e.message || '更新失败')
  } finally {
    togglingId.value = null
  }
}

// ============== 密码显示/隐藏控制 ===============
// 表格每行密码显示状态，默认隐藏
const showPwdMap = ref({})
const toggleRowPwd = (row) => {
  const id = row.id
  showPwdMap.value[id] = !showPwdMap.value[id]
}
// 详情密码显示状态，默认隐藏
const showDetailPwd = ref(false)

// 是否显示某一行的密码（模板安全取值）
const isRowPwdVisible = (row) => {
  const id = row?.id
  return !!(id != null && showPwdMap.value && showPwdMap.value[id])
}
</script>

<style>

</style>

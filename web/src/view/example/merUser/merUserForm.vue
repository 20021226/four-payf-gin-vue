
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="接入类型:" prop="merType">
    <el-input v-model="formData.merType" :clearable="true" placeholder="请输入接入类型" />
</el-form-item>
        <el-form-item label="账号:" prop="userName">
    <el-input v-model="formData.userName" :clearable="true" placeholder="请输入账号" />
</el-form-item>
        <el-form-item label="密码:" prop="password">
    <el-input v-model="formData.password" :clearable="true" placeholder="请输入密码" />
</el-form-item>
        <el-form-item label="是否启用:" prop="state">
    <el-switch v-model="formData.state" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item label="收款码:" prop="qrCode">
    <el-input v-model="formData.qrCode" :clearable="true" placeholder="请输入收款码" />
</el-form-item>
        <el-form-item label="请求密钥:" prop="key">
    <el-input v-model="formData.key" :clearable="true" placeholder="请输入请求密钥" />
</el-form-item>
        <el-form-item label="备注:" prop="remarks">
    <el-input v-model="formData.remarks" :clearable="true" placeholder="请输入备注" />
</el-form-item>
        <el-form-item>
          <el-button :loading="btnLoading" type="primary" @click="save">保存</el-button>
          <el-button type="primary" @click="back">返回</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import {
  createMerUser,
  updateMerUser,
  findMerUser
} from '@/api/example/merUser'

defineOptions({
    name: 'MerUserForm'
})

// 自动获取字典
import { getDictFunc } from '@/utils/format'
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'


const route = useRoute()
const router = useRouter()

// 提交按钮loading
const btnLoading = ref(false)

const type = ref('')
const formData = ref({
            merType: '',
            userName: '',
            password: '',
            state: false,
            qrCode: '',
            key: '',
            remarks: '',
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findMerUser({ ID: route.query.id })
      if (res.code === 0) {
        formData.value = res.data
        type.value = 'update'
      }
    } else {
      type.value = 'create'
    }
}

init()
// 保存按钮
const save = async() => {
      btnLoading.value = true
      elFormRef.value?.validate( async (valid) => {
         if (!valid) return btnLoading.value = false
            let res
           switch (type.value) {
             case 'create':
               res = await createMerUser(formData.value)
               break
             case 'update':
               res = await updateMerUser(formData.value)
               break
             default:
               res = await createMerUser(formData.value)
               break
           }
           btnLoading.value = false
           if (res.code === 0) {
             ElMessage({
               type: 'success',
               message: '创建/更改成功'
             })
           }
       })
}

// 返回按钮
const back = () => {
    router.go(-1)
}

</script>

<style>
</style>

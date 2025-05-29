<template>
  <div class="email-list">
    <div class="page-header">

    </div>

    <!-- 搜索过滤 -->
    <el-card style="margin-bottom: 5px">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="邮箱地址">
          <el-input v-model="searchForm.email" placeholder="请输入邮箱地址" clearable />
        </el-form-item>
        <el-form-item label="邮箱提供商">
          <el-select v-model="searchForm.provider" style="width: 150px" placeholder="请选择提供商" clearable>
            <el-option label="Gmail" value="gmail" />
            <el-option label="Outlook" value="outlook" />
            <el-option label="QQ邮箱" value="qq" />
            <el-option label="163邮箱" value="163" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadEmails">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="primary" @click="showCreateDialog = true">
              <el-icon><Plus /></el-icon>
              添加邮箱
            </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 邮箱列表 -->
    <el-card>
      <el-table :data="emails" style="width: 100%" v-loading="loading">
        <el-table-column prop="email" label="邮箱地址" width="150" />
        <el-table-column prop="display_name" label="显示名称" width="150" />
        <el-table-column prop="provider" label="提供商" width="80" />
        <el-table-column prop="phone" label="手机号" width="150" />
        <el-table-column prop="service_count" label="服务数量" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="150">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button size="small" @click="editEmail(scope.row)">编辑</el-button>
            <el-button size="small" type="info" @click="viewServices(scope.row)">服务</el-button>
            <el-button size="small" type="danger" @click="deleteEmail(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <!-- 分页 -->
        <el-pagination
          v-if="total > 0"
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          style="margin-top: 20px; text-align: right"
        />
    </el-card>

    <!-- 创建/编辑邮箱对话框 -->
    <el-dialog 
      v-model="showCreateDialog" 
      :title="isEdit ? '编辑邮箱' : '添加邮箱'"
      width="500px"
    >
      <el-form :model="emailForm" :rules="emailRules" ref="emailFormRef" label-width="100px">
        <el-form-item label="邮箱地址" prop="email">
          <el-input v-model="emailForm.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="emailForm.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="emailForm.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item label="邮箱提供商" prop="provider">
          <el-select v-model="emailForm.provider" placeholder="请选择提供商">
            <el-option label="Gmail" value="gmail" />
            <el-option label="Outlook" value="outlook" />
            <el-option label="QQ邮箱" value="qq" />
            <el-option label="163邮箱" value="163" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="emailForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="备用邮箱" prop="backup_email">
          <el-input v-model="emailForm.backup_email" placeholder="请输入备用邮箱" />
        </el-form-item>
        <el-form-item label="安全问题" prop="security_question">
          <el-input v-model="emailForm.security_question" placeholder="请输入安全问题" />
        </el-form-item>
        <el-form-item label="安全答案" prop="security_answer">
          <el-input v-model="emailForm.security_answer" placeholder="请输入安全答案" />
        </el-form-item>
        <el-form-item label="备注" prop="notes">
          <el-input v-model="emailForm.notes" type="textarea" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEmail">确定</el-button>
      </template>
    </el-dialog>

     <!-- 查看服务对话框 -->
     <el-dialog 
      v-model="showServerDialog" 
      title="服务列表"
      width="500"
    >
    <el-card>
      <el-table :data="servers" style="width: 100%" v-loading="serverLoading">
        <el-table-column prop="service_name" label="服务名称" width="150" />
        <el-table-column label="网站" width="350">
          <template #default="scope">
            <a :href="formatLink(scope.row.service_website)" target="_blank" rel="noopener noreferrer">
              {{ scope.row.service_website }}
            </a>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <template #footer>
        <el-button @click="showServerDialog = false">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { emailAPI } from '@/utils/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

export default {
name: 'EmailListPage',
components: {
  Plus
},
setup() {
  const emails = ref([])
  const servers = ref([])

  const loading = ref(false)
  const serverLoading = ref(false)
  const showCreateDialog = ref(false)
  const showServerDialog = ref(false) // 修正变量名

  const isEdit = ref(false)
  const emailFormRef = ref(null)

  // 添加分页相关的响应式数据
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(10)

  const searchForm = reactive({
    email: '',
    provider: ''
  })

  const emailForm = reactive({
    id: null,
    email: '',
    password: '',
    display_name: '',
    provider: '',
    phone: '',
    backup_email: '',
    security_question: '',
    security_answer: '',
    notes: ''
  })

  const emailRules = {
    email: [
      { required: true, message: '请输入邮箱地址', trigger: 'blur' },
      { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
    ],
    display_name: [
      { required: true, message: '请输入显示名称', trigger: 'blur' }
    ],
    provider: [
      { required: true, message: '请选择邮箱提供商', trigger: 'change' }
    ]
  }

  // 修改加载邮箱方法，支持分页
  const loadEmails = async () => {
    loading.value = true
    try {
      // 构建查询参数，包含分页信息
      const params = {
        ...searchForm,
        page: currentPage.value,
        pageSize: pageSize.value
      }
      
      const response = await emailAPI.getEmails(params)
      
      // 假设API返回格式为 { data: [...], total: number }
      if (response.data) {
        emails.value = response.data
        total.value = response.total || 0
      } else {
        // 如果API直接返回数组，则需要处理
        emails.value = response
        total.value = response.length
      }
    } catch (error) {
      ElMessage.error('加载邮箱列表失败')
      emails.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  // 添加分页处理方法
  const handleSizeChange = (newPageSize) => {
    pageSize.value = newPageSize
    currentPage.value = 1 // 重置到第一页
    loadEmails()
  }

  const handleCurrentChange = (newPage) => {
    currentPage.value = newPage
    loadEmails()
  }

  const resetSearch = () => {
    searchForm.email = ''
    searchForm.provider = ''
    currentPage.value = 1 // 重置页码
    loadEmails()
  }

  const resetForm = () => {
    Object.keys(emailForm).forEach(key => {
      emailForm[key] = key === 'id' ? null : ''
    })
  }

  const editEmail = (row) => {
    isEdit.value = true
    Object.keys(emailForm).forEach(key => {
      emailForm[key] = row[key] || ''
    })
    showCreateDialog.value = true
  }

  const saveEmail = async () => {
    if (!emailFormRef.value) return
    
    await emailFormRef.value.validate(async (valid) => {
      if (valid) {
        try {
          if (isEdit.value) {
            await emailAPI.updateEmail(emailForm.id, emailForm)
            ElMessage.success('更新邮箱成功')
          } else {
            await emailAPI.createEmail(emailForm)
            ElMessage.success('创建邮箱成功')
          }
          showCreateDialog.value = false
          resetForm()
          loadEmails()
        } catch (error) {
          ElMessage.error(isEdit.value ? '更新邮箱失败' : '创建邮箱失败')
        }
      }
    })
  }

  const deleteEmail = async (row) => {
    try {
      await ElMessageBox.confirm('此操作将删除该邮箱，是否继续？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      await emailAPI.deleteEmail(row.id)
      ElMessage.success('删除成功')
      loadEmails()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error('删除失败')
      }
    }
  }

  const viewServices = async (row) => {
    if (row.service_count == 0) {
      ElMessage.info(`没有绑定的服务`)
      return
    }
    
    serverLoading.value = true
    showServerDialog.value = true // 修正：正确使用.value
    
    try {
      // 修正：应该调用emailAPI或其他服务API，而不是servers
      const response = await emailAPI.getEmailServices(row.id) // 假设这是正确的API调用
      servers.value = response.data || response
    } catch (error) {
      ElMessage.error('获取服务数据失败')
      servers.value = []
    } finally {
      serverLoading.value = false
    }
  }

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString()
  }

  const formatLink = (url) => {
    if (url && !url.startsWith('http://') && !url.startsWith('https://')) {
      return 'http://' + url;
    }
    return url;
  };

  onMounted(() => {
    loadEmails()
  })

  return {
    emails,
    servers,
    loading,
    serverLoading,
    showCreateDialog,
    showServerDialog, // 修正变量名
    isEdit,
    emailFormRef,
    searchForm,
    emailForm,
    emailRules,
    // 添加分页相关的返回值
    total,
    currentPage,
    pageSize,
    loadEmails,
    resetSearch,
    editEmail,
    saveEmail,
    deleteEmail,
    viewServices,
    formatDate,
    formatLink, // 导出 formatLink 方法
    // 添加分页处理方法
    handleSizeChange,
    handleCurrentChange
  }
}
}
</script>
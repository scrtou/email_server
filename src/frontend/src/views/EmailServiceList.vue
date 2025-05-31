<template>
  <div class="email-service-list">
    <div class="page-header">

    </div>

    <!-- 搜索过滤 -->
    <el-card style="margin-bottom: 20px">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="邮箱">
          <el-select v-model="searchForm.email_id" placeholder="请选择邮箱" clearable filterable>
            <el-option 
              v-for="email in emails" 
              :key="email.id" 
              :label="email.email" 
              :value="email.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="服务">
          <el-select v-model="searchForm.service_id" placeholder="请选择服务" clearable filterable>
            <el-option 
              v-for="service in services" 
              :key="service.id" 
              :label="service.name" 
              :value="service.id" 
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            添加关联
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 关联列表 -->
    <el-card style="display: flex; flex-direction: column; flex-grow: 1; overflow: hidden;">
      <div class="table-container" style="flex-grow: 1; overflow-y: auto;">
        <el-table
          :data="emailServices"
          style="width: 100%"
          height="100%"
          v-loading="loading"
          empty-text="暂无关联数据"
        >
        <el-table-column prop="email_addr" label="邮箱地址" min-width="200" />
        <el-table-column prop="service_name" label="服务名称" min-width="150" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="subscription_type" label="订阅类型" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.subscription_type" type="success">
              {{ scope.row.subscription_type }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="registration_date" label="注册日期" width="120">
          <template #default="scope">
            {{ scope.row.registration_date ? formatDate(scope.row.registration_date, 'date') : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="viewDetail(scope.row)">查看</el-button>
            <el-button size="small" type="primary" @click="editEmailService(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteEmailService(scope.row)">删除</el-button>
          </template>
        </el-table-column>
        </el-table>
      </div>

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
      />
    </el-card>

    <!-- 创建/编辑关联对话框 -->
    <el-dialog 
      :title="isEdit ? '编辑关联' : '添加关联'"
      v-model="showCreateDialog"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form 
        ref="emailServiceFormRef"
        :model="emailServiceForm"
        :rules="emailServiceRules"
        label-width="100px"
      >
        <el-form-item label="邮箱" prop="email_id">
          <el-select 
            v-model="emailServiceForm.email_id" 
            placeholder="请选择邮箱"
            filterable
            style="width: 100%"
          >
            <el-option 
              v-for="email in emails" 
              :key="email.id" 
              :label="email.email" 
              :value="email.id" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="服务" prop="service_id">
          <el-select 
            v-model="emailServiceForm.service_id" 
            placeholder="请选择服务"
            filterable
            style="width: 100%"
          >
            <el-option 
              v-for="service in services" 
              :key="service.id" 
              :label="service.name" 
              :value="service.id" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="用户名" prop="username">
          <el-input v-model="emailServiceForm.username" placeholder="请输入用户名" />
        </el-form-item>
        
        <el-form-item label="密码" prop="password">
          <el-input 
            v-model="emailServiceForm.password" 
            type="password" 
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="emailServiceForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        
        <el-form-item label="注册日期" prop="registration_date">
          <el-date-picker
            v-model="emailServiceForm.registration_date"
            type="date"
            placeholder="请选择注册日期"
            style="width: 100%"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        
        <el-form-item label="订阅类型" prop="subscription_type">
          <el-select v-model="emailServiceForm.subscription_type" placeholder="请选择订阅类型">
            <el-option label="免费版" value="free" />
            <el-option label="基础版" value="basic" />
            <el-option label="专业版" value="pro" />
            <el-option label="企业版" value="enterprise" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="到期日期" prop="subscription_expires">
          <el-date-picker
            v-model="emailServiceForm.subscription_expires"
            type="date"
            placeholder="请选择到期日期"
            style="width: 100%"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        
        <el-form-item label="备注" prop="notes">
          <el-input 
            v-model="emailServiceForm.notes" 
            type="textarea" 
            placeholder="请输入备注信息"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveEmailService" :loading="saving">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { emailServiceAPI, emailAPI, serviceAPI } from '@/utils/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'

export default {
  name: 'EmailServiceListPage',
  components: {
    Plus,
    Search
  },
  setup() {
    // 响应式数据
    const emailServices = ref([])
    const emails = ref([])
    const services = ref([])
    const loading = ref(true)
    const saving = ref(false)
    const showCreateDialog = ref(false)
    const isEdit = ref(false)
    const emailServiceFormRef = ref(null)
    
    // 分页数据
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)

    // 搜索表单
    const searchForm = reactive({
      email_id: '',
      service_id: ''
    })

    // 邮箱服务关联表单
    const emailServiceForm = reactive({
      id: null,
      email_id: '',
      service_id: '',
      username: '',
      password: '',
      phone: '',
      registration_date: '',
      subscription_type: '',
      subscription_expires: '',
      notes: ''
    })

    // 表单验证规则
    const emailServiceRules = {
      email_id: [
        { required: true, message: '请选择邮箱', trigger: 'change' }
      ],
      service_id: [
        { required: true, message: '请选择服务', trigger: 'change' }
      ],
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' }
      ]
    }

    // 加载邮箱服务关联数据
    const loadData = async () => {
      loading.value = true
      
      try {
        console.log('开始加载邮箱服务关联数据...')
        
        const params = {
          ...searchForm,
          page: currentPage.value,
          page_size: pageSize.value
        }
        
        const startTime = Date.now()
        const response = await emailServiceAPI.getAllEmailServices(params)
        const endTime = Date.now()
        
        console.log(`API请求耗时: ${endTime - startTime}ms`)
        
        // 处理返回的分页数据
        if (response && response.data) {
          emailServices.value = response.data
          total.value = response.total || 0
        } else if (Array.isArray(response)) {
          emailServices.value = response
          total.value = response.length
        } else {
          emailServices.value = []
          total.value = 0
        }
        
        console.log('数据加载完成:', emailServices.value.length, '条记录')
        
      } catch (error) {
        console.error('加载邮箱服务关联失败:', error)
        ElMessage.error('加载数据失败')
        emailServices.value = []
        total.value = 0
      } finally {
        loading.value = false
      }
    }

    // 加载邮箱列表
    const loadEmails = async () => {
      try {
        const response = await emailAPI.getEmails({ page: 'all' })
        emails.value = Array.isArray(response) ? response : (response.data || [])
      } catch (error) {
        console.error('加载邮箱列表失败:', error)
        ElMessage.error('加载邮箱列表失败')
      }
    }

    // 加载服务列表
    const loadServices = async () => {
      try {
        const response = await serviceAPI.getServices({ page: 'all' })
        services.value = Array.isArray(response) ? response : (response.data || [])
      } catch (error) {
        console.error('加载服务列表失败:', error)
        ElMessage.error('加载服务列表失败')
      }
    }

    // 分页处理方法
    const handleSizeChange = (newPageSize) => {
      pageSize.value = newPageSize
      currentPage.value = 1
      loadData()
    }

    const handleCurrentChange = (newPage) => {
      currentPage.value = newPage
      loadData()
    }

    // 重置搜索
    const resetSearch = () => {
      searchForm.email_id = ''
      searchForm.service_id = ''
      currentPage.value = 1
      loadData()
    }

    // 重置表单
    const resetForm = () => {
      Object.keys(emailServiceForm).forEach(key => {
        emailServiceForm[key] = key === 'id' ? null : ''
      })
      isEdit.value = false
    }

    // 打开创建对话框
    const openCreateDialog = () => {
      resetForm()
      showCreateDialog.value = true
    }

    // 编辑邮箱服务关联
    const editEmailService = (row) => {
      isEdit.value = true
      Object.keys(emailServiceForm).forEach(key => {
        emailServiceForm[key] = row[key] || ''
      })
      showCreateDialog.value = true
    }

    // 保存邮箱服务关联
    const saveEmailService = async () => {
      if (!emailServiceFormRef.value) return
      
      await emailServiceFormRef.value.validate(async (valid) => {
        if (valid) {
          saving.value = true
          try {
            if (isEdit.value) {
              await emailServiceAPI.updateEmailService(emailServiceForm.id, emailServiceForm)
              ElMessage.success('更新关联成功')
            } else {
              await emailServiceAPI.createEmailService(emailServiceForm)
              ElMessage.success('创建关联成功')
            }
            showCreateDialog.value = false
            resetForm()
            loadData()
          } catch (error) {
            console.error('保存邮箱服务关联失败:', error)
            ElMessage.error(isEdit.value ? '更新关联失败' : '创建关联失败')
          } finally {
            saving.value = false
          }
        }
      })
    }

    // 删除邮箱服务关联
    const deleteEmailService = async (row) => {
      try {
        await ElMessageBox.confirm('此操作将删除该关联，是否继续？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        await emailServiceAPI.deleteEmailService(row.id)
        ElMessage.success('删除成功')
        loadData()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除邮箱服务关联失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }

    // 查看详情
    const viewDetail = (row) => {
      ElMessage.info(`查看关联详情功能待实现`+row)
    }

    // 对话框关闭处理
    const handleDialogClose = () => {
      resetForm()
      if (emailServiceFormRef.value) {
        emailServiceFormRef.value.resetFields()
      }
    }

    // 格式化日期
    const formatDate = (dateString, type = 'datetime') => {
      if (!dateString) return '-'
      const date = new Date(dateString)
      if (type === 'date') {
        return date.toLocaleDateString()
      }
      return date.toLocaleString()
    }

    // 组件挂载时加载数据
    onMounted(async () => {
      console.log('组件已挂载，开始加载数据...')
      // 并行加载所有数据
      await Promise.all([
        loadData(),
        loadEmails(),
        loadServices()
      ])
    })

    return {
      // 数据
      emailServices,
      emails,
      services,
      loading,
      saving,
      showCreateDialog,
      isEdit,
      emailServiceFormRef,
      searchForm,
      emailServiceForm,
      emailServiceRules,
      total,
      currentPage,
      pageSize,
      
      // 方法
      loadData,
      resetSearch,
      openCreateDialog,
      editEmailService,
      saveEmailService,
      deleteEmailService,
      viewDetail,
      handleDialogClose,
      formatDate,
      handleSizeChange,
      handleCurrentChange
    }
  }
}
</script>

<style scoped>
.email-service-list {
  padding: 20px;
  display: flex;
  flex-direction: column;
  height: 100%;
  box-sizing: border-box;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
}

/* 表格内容对齐 */
.el-table .cell {
  word-break: break-word;
}

/* 操作按钮间距 */
.el-table .el-button + .el-button {
  margin-left: 5px;
}

/* Pagination styles moved to utilities.css */
</style>
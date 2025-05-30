<template>
  <div class="service-list">
    <div class="page-header">
      
    </div>

    <!-- 搜索过滤 -->
    <el-card style="margin-bottom: 5px">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="服务名称">
          <el-input v-model="searchForm.name" placeholder="请输入服务名称" clearable />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="searchForm.category" style="width:150px" placeholder="请选择分类" clearable>
            <el-option label="社交媒体" value="social" />
            <el-option label="电商购物" value="shopping" />
            <el-option label="金融服务" value="finance" />
            <el-option label="娱乐游戏" value="entertainment" />
            <el-option label="工具软件" value="tools" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadServices">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            添加服务
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 服务列表 -->
    <el-card>
      <el-table :data="services" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="服务名称" width="200" />
        <el-table-column label="网站" width="250">
          <template #default="scope">
            <a :href="formatLink(scope.row.website)" target="_blank" rel="noopener noreferrer">
              {{ scope.row.website }}
            </a>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="description" label="描述" width="300" />
        <el-table-column prop="email_count" label="邮箱数量" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button size="small" @click="editService(scope.row)">编辑</el-button>
            <el-button size="small" type="info" @click="viewEmails(scope.row)">邮箱</el-button>
            <el-button size="small" type="danger" @click="deleteService(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 添加分页组件 -->
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

    <!-- 创建/编辑服务对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="isEdit ? '编辑服务' : '添加服务'"
      width="500px"
    >
      <el-form :model="serviceForm" :rules="serviceRules" ref="serviceFormRef" label-width="100px">
        <el-form-item label="服务名称" prop="name">
          <el-input v-model="serviceForm.name" placeholder="请输入服务名称" />
        </el-form-item>
        <el-form-item label="网站地址" prop="website">
          <el-input v-model="serviceForm.website" placeholder="请输入网站地址" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="serviceForm.category" placeholder="请选择分类">
            <el-option label="社交媒体" value="social" />
            <el-option label="电商购物" value="shopping" />
            <el-option label="金融服务" value="finance" />
            <el-option label="娱乐游戏" value="entertainment" />
            <el-option label="工具软件" value="tools" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="serviceForm.description" type="textarea" placeholder="请输入服务描述" />
        </el-form-item>
        <el-form-item label="Logo地址" prop="logo_url">
          <el-input v-model="serviceForm.logo_url" placeholder="请输入Logo图片地址" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveService">确定</el-button>
      </template>
    </el-dialog>

    <!-- 查看邮箱对话框 -->
    <el-dialog
      v-model="showEmailDialog"
      title="邮箱列表"
      width="600px"
    >
      <el-card>
        <el-table :data="emails" style="width: 100%" v-loading="emailLoading">
          <el-table-column prop="email_addr" label="邮箱地址" width="200" />
          <el-table-column prop="display_name" label="显示名称" width="150" />
          <el-table-column prop="provider" label="提供商" width="100" />
        </el-table>
      </el-card>
      <template #footer>
        <el-button @click="showEmailDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { serviceAPI } from '@/utils/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

export default {
  name: 'ServiceListPage',
  components: {
    Plus
  },
  setup() {
    const services = ref([])
    const emails = ref([]) // 用于存储邮箱列表
    const loading = ref(false)
    const emailLoading = ref(false) // 邮箱列表加载状态
    const showCreateDialog = ref(false)
    const showEmailDialog = ref(false) // 控制邮箱列表对话框的显示
    const isEdit = ref(false)
    const serviceFormRef = ref(null)
    
    // 添加分页相关的响应式数据
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)

    const searchForm = reactive({
      name: '',
      category: ''
    })

    const serviceForm = reactive({
      id: null,
      name: '',
      website: '',
      category: '',
      description: '',
      logo_url: ''
    })

    const serviceRules = {
      name: [
        { required: true, message: '请输入服务名称', trigger: 'blur' }
      ],
      category: [
        { required: true, message: '请选择分类', trigger: 'change' }
      ]
    }

    // 修改加载服务方法，支持分页
    const loadServices = async () => {
      loading.value = true
      try {
        // 构建查询参数，包含分页信息
        const params = {
          ...searchForm,
          page: currentPage.value,
          pageSize: pageSize.value
        }
        
        const response = await serviceAPI.getServices(params)
        
        // 处理返回的分页数据
        if (response.data) {
          services.value = response.data
          total.value = response.total || 0
        } else {
          // 兼容直接返回数组的情况
          services.value = response
          total.value = response.length
        }
      } catch (error) {
        ElMessage.error('加载服务列表失败')
        services.value = []
        total.value = 0
      } finally {
        loading.value = false
      }
    }
    
    // 添加分页处理方法
    const handleSizeChange = (newPageSize) => {
      pageSize.value = newPageSize
      currentPage.value = 1 // 重置到第一页
      loadServices()
    }

    const handleCurrentChange = (newPage) => {
      currentPage.value = newPage
      loadServices()
    }

    const resetSearch = () => {
      searchForm.name = ''
      searchForm.category = ''
      currentPage.value = 1 // 重置页码
      loadServices()
    }

    const resetForm = () => {
      Object.keys(serviceForm).forEach(key => {
        serviceForm[key] = key === 'id' ? null : ''
      })
    }

    const editService = (row) => {
      isEdit.value = true
      Object.keys(serviceForm).forEach(key => {
        serviceForm[key] = row[key] || ''
      })
      showCreateDialog.value = true
    }

    const saveService = async () => {
      if (!serviceFormRef.value) return
      
      await serviceFormRef.value.validate(async (valid) => {
        if (valid) {
          try {
            if (isEdit.value) {
              await serviceAPI.updateService(serviceForm.id, serviceForm)
              ElMessage.success('更新服务成功')
            } else {
              await serviceAPI.createService(serviceForm)
              ElMessage.success('创建服务成功')
            }
            showCreateDialog.value = false
            resetForm()
            loadServices()
          } catch (error) {
            ElMessage.error(isEdit.value ? '更新服务失败' : '创建服务失败')
          }
        }
      })
    }

    const deleteService = async (row) => {
      try {
        await ElMessageBox.confirm('此操作将删除该服务，是否继续？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        await serviceAPI.deleteService(row.id)
        ElMessage.success('删除成功')
        loadServices()
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('删除失败')
        }
      }
    }

    const viewEmails = async (row) => {
      if (row.email_count === 0) {
        ElMessage.info(`服务 ${row.name} 没有绑定的邮箱`)
        return
      }
      
      emailLoading.value = true
      showEmailDialog.value = true
      
      try {
        // 假设 serviceAPI 有一个 getServiceEmails 方法
        const response = await serviceAPI.getServiceEmails(row.id)
        // API 拦截器会处理 response.data.data，所以 response 直接就是数据
        // 尝试更灵活地处理数据结构，例如 { items: [...] } 或直接是 [...]
        if (Array.isArray(response)) {
          emails.value = response
        } else if (response && Array.isArray(response.items)) {
          emails.value = response.items
        } else if (response && Array.isArray(response.data)) { // 再次检查 .data，以防万一
          emails.value = response.data
        }
        else {
          emails.value = [] // 默认空数组
        }
      } catch (error) {
        ElMessage.error('获取邮箱列表失败: ' + (error.message || '未知错误'))
        emails.value = []
      } finally {
        emailLoading.value = false
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
      loadServices()
    })
  
    return {
      services,
      emails, // 导出 emails
      loading,
      emailLoading, // 导出 emailLoading
      showCreateDialog,
      showEmailDialog, // 导出 showEmailDialog
      isEdit,
      serviceFormRef,
      searchForm,
      serviceForm,
      serviceRules,
      // 添加分页相关的返回值
      total,
      currentPage,
      pageSize,
      loadServices,
      resetSearch,
      editService,
      saveService,
      deleteService,
      viewEmails,
      formatDate,
      formatLink, // 导出 formatLink 方法
      // 添加分页处理方法
      handleSizeChange,
      handleCurrentChange
    }
  }
  }
</script>

<style scoped>
.service-list {
  padding: 20px;
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
</style>
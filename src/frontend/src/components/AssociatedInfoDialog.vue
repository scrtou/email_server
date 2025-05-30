<template>
  <el-dialog
    :model-value="visible"
    :title="title"
    width="60%"
    @update:model-value="$emit('update:visible', $event)"
    :close-on-click-modal="false"
    append-to-body
  >
    <div v-loading="loading" class="dialog-content-wrapper">
      <el-table :data="items" style="width: 100%" empty-text="没有可显示的信息。" class="dialog-table">
        <template v-for="col in itemLayout" :key="col.prop">
          <el-table-column
            :prop="col.prop"
            :label="col.label"
            :width="col.width"
            :min-width="col.minWidth || '100px'"
            show-overflow-tooltip
          >
            <template #default="scope">
              <span v-if="col.type === 'link'">
                <a :href="scope.row[col.prop]" target="_blank" rel="noopener noreferrer" class="table-link">
                  {{ scope.row[col.prop] || '-' }}
                </a>
              </span>
              <span v-else>
                {{ scope.row[col.prop] || '-' }}
              </span>
            </template>
          </el-table-column>
        </template>
      </el-table>

      <el-pagination
        v-if="pagination && pagination.totalItems > 0 && pagination.totalItems > pagination.pageSize"
        class="dialog-pagination"
        background
        layout="total, prev, pager, next"
        :total="pagination.totalItems"
        :current-page="pagination.currentPage"
        :page-size="pagination.pageSize"
        @current-change="handleCurrentPageChange"
      />
      <!-- Note: pageSize change is not implemented in this dialog as it's usually fixed for dialogs -->
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="$emit('update:visible', false)">关闭</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ElDialog, ElTable, ElTableColumn, ElPagination, ElButton } from 'element-plus'; // Explicitly import if needed, though auto-import usually works

// eslint-disable-next-line no-undef
const props = defineProps({
  visible: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    default: '关联信息',
  },
  items: {
    type: Array,
    default: () => [],
  },
  itemLayout: { // Describes table columns
    type: Array,
    required: true,
    // Example: [{ label: 'Platform Name', prop: 'platform_name', minWidth: '150px' }, { label: 'URL', prop: 'platform_website_url', type: 'link', minWidth: '200px' }]
  },
  pagination: { // Pagination object: { currentPage, pageSize, totalItems }
    type: Object,
    default: null, // null or undefined means no pagination
  },
  loading: {
    type: Boolean,
    default: false,
  }
});

// eslint-disable-next-line no-undef
const emit = defineEmits(['update:visible', 'page-change']);

const handleCurrentPageChange = (newPage) => {
  if (props.pagination) {
    emit('page-change', { currentPage: newPage, pageSize: props.pagination.pageSize });
  }
};

</script>

<style scoped>
.dialog-content-wrapper {
  min-height: 200px; /* Ensure a minimum height even when loading or empty */
}
.dialog-table {
  margin-bottom: 20px;
}
.table-link {
  color: #409EFF;
  text-decoration: none;
}
.table-link:hover {
  text-decoration: underline;
}
.dialog-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 15px;
}
.dialog-content { /* This class was used for the old card layout, can be removed or repurposed if needed */
  max-height: 60vh; /* Keep max height for scrollability if content overflows */
  overflow-y: auto;
}
</style>
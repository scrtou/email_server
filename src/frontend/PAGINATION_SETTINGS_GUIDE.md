# 分页设置系统说明

## 概述

本系统已经从全局统一分页设置改为页面独立分页设置，每个页面可以单独配置自己的分页大小和选项。

## 页面独立分页配置

### 当前配置的页面

1. **邮箱账户管理** (`emailAccounts`)
   - 默认分页大小: 10
   - 可选项: [5, 10, 15, 20, 30, 50]

2. **平台管理** (`platforms`)
   - 默认分页大小: 15
   - 可选项: [10, 15, 20, 30, 50]

3. **平台注册管理** (`platformRegistrations`)
   - 默认分页大小: 8
   - 可选项: [5, 8, 10, 15, 20, 30]

4. **服务订阅管理** (`serviceSubscriptions`)
   - 默认分页大小: 12
   - 可选项: [8, 12, 15, 20, 30, 50]

## 设置存储

- 每个页面的分页设置独立保存在 localStorage 中
- 存储键格式: `pageSize_${pageName}`
- 例如: `pageSize_emailAccounts`, `pageSize_platforms`

## 如何使用

### 在页面组件中

```javascript
// 获取页面专用的分页大小
const pageSize = settingsStore.getPageSize('emailAccounts');

// 获取页面专用的分页选项
const pageSizeOptions = settingsStore.getPageSizeOptions('emailAccounts');

// 设置页面专用的分页大小
settingsStore.setPageSize('emailAccounts', 15);
```

### 在模板中

```vue
<el-pagination
  :page-sizes="settingsStore.getPageSizeOptions('emailAccounts')"
  @size-change="handleSizeChange"
/>
```

## 添加新页面

如果需要为新页面添加独立分页设置：

1. 在 `src/frontend/src/stores/settings.js` 的 `pageSettings` 中添加新页面配置
2. 在页面组件中使用对应的页面名称调用设置方法
3. 在分页器的 `@size-change` 事件中保存设置

## 兼容性

- 保留了 `getDefaultPageSize` getter 用于向后兼容
- 旧的全局设置仍然可以作为后备选项使用

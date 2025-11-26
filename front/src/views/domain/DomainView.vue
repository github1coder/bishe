<template>
  <div class="domain-view">
    <t-typography-title level="h5">
      <span style="color: #0052d9">数据域管理</span>
    </t-typography-title>

    <!-- 操作选项卡 -->
    <t-tabs v-model="activeTab" :list="tabList" @change="onTabChange" />

    <!-- 创建数据域 -->
    <t-card v-show="activeTab === 'create'" class="operation-card">
      <template #header>
        <t-typography-title level="h6">创建数据域</t-typography-title>
      </template>
      <t-form
        :rules="createRules"
        :data="createForm"
        ref="createFormRef"
        label-width="120px"
        @submit="handleCreateDomain"
        colon
      >
        <t-form-item label="数据域名称" name="domainName">
          <t-input
            v-model="createForm.domainName"
            placeholder="请输入数据域名称"
            clearable
          />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" type="submit" :loading="createLoading">
            创建数据域
          </t-button>
          <t-button theme="default" @click="resetCreateForm" style="margin-left: 12px">
            重置
          </t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 更新数据域元数据 -->
    <t-card v-show="activeTab === 'update'" class="operation-card">
      <template #header>
        <t-typography-title level="h6">更新数据域元数据</t-typography-title>
      </template>
      <t-form
        :rules="updateRules"
        :data="updateForm"
        ref="updateFormRef"
        label-width="120px"
        @submit="handleUpdateDomain"
        colon
      >
        <t-form-item label="数据域名称" name="domainName">
          <t-input
            v-model="updateForm.domainName"
            placeholder="请输入数据域名称"
            clearable
          />
        </t-form-item>
        <t-form-item label="数据域成员" name="domainMembers">
          <t-textarea
            v-model="updateForm.domainMembers"
            placeholder="请输入数据域成员，多个成员用英文逗号或中文逗号分隔，例如：hospitalA,hospitalB 或 hospitalA，hospitalB"
            :autosize="{ minRows: 3, maxRows: 5 }"
            clearable
          />
        </t-form-item>
        <t-form-item label="数据域策略" name="domainPolicy">
          <t-textarea
            v-model="updateForm.domainPolicy"
            placeholder="请输入数据域策略"
            :autosize="{ minRows: 3, maxRows: 5 }"
            clearable
          />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" type="submit" :loading="updateLoading">
            更新元数据
          </t-button>
          <t-button theme="default" @click="resetUpdateForm" style="margin-left: 12px">
            重置
          </t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 查询我参与的数据域 -->
    <t-card v-show="activeTab === 'myDomains'" class="operation-card">
      <template #header>
        <div class="card-header">
          <t-typography-title level="h6">查询我参与的数据域</t-typography-title>
          <t-button theme="primary" @click="handleQueryMyDomains" :loading="queryMyDomainsLoading">
            查询
          </t-button>
        </div>
      </template>
      <div v-if="myDomainsResult && myDomainsResult.length > 0">
        <t-table
          :data="myDomainsResult"
          :columns="domainTableColumns"
          row-key="domainName"
          :hover="true"
          stripe
        />
      </div>
      <t-empty v-else description="暂无数据，请点击查询按钮" />
    </t-card>

    <!-- 查询我管理的数据域 -->
    <t-card v-show="activeTab === 'managedDomains'" class="operation-card">
      <template #header>
        <div class="card-header">
          <t-typography-title level="h6">查询我管理的数据域</t-typography-title>
          <t-button theme="primary" @click="handleQueryManagedDomains" :loading="queryManagedDomainsLoading">
            查询
          </t-button>
        </div>
      </template>
      <div v-if="managedDomainsResult && managedDomainsResult.length > 0">
        <t-table
          :data="managedDomainsResult"
          :columns="domainTableColumns"
          row-key="domainName"
          :hover="true"
          stripe
        />
      </div>
      <t-empty v-else description="暂无数据，请点击查询按钮" />
    </t-card>

    <!-- 查询数据域详细信息 -->
    <t-card v-show="activeTab === 'domainInfo'" class="operation-card">
      <template #header>
        <t-typography-title level="h6">查询数据域详细信息</t-typography-title>
      </template>
      <t-form
        :rules="queryInfoRules"
        :data="queryInfoForm"
        ref="queryInfoFormRef"
        label-width="120px"
        @submit="handleQueryDomainInfo"
        colon
      >
        <t-form-item label="数据域名称" name="domainName">
          <t-input
            v-model="queryInfoForm.domainName"
            placeholder="请输入数据域名称"
            clearable
          />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" type="submit" :loading="queryInfoLoading">
            查询
          </t-button>
          <t-button theme="default" @click="resetQueryInfoForm" style="margin-left: 12px">
            重置
          </t-button>
        </t-form-item>
      </t-form>
      <div v-if="domainInfoResult" class="info-result">
        <t-card>
          <div class="domain-info-display">
            <div class="info-item">
              <span class="info-label">数据域名称：</span>
              <span class="info-value">{{ domainInfoResult.domainName || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">数据域成员：</span>
              <span class="info-value">{{ domainInfoResult.domainMembers || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">数据域策略：</span>
            </div>
            <div v-if="domainInfoResult.domainPolicy" class="policy-section">
              <pre class="policy-content">{{ domainInfoResult.domainPolicy }}</pre>
            </div>
          </div>
        </t-card>
      </div>
      <t-empty v-else-if="!queryInfoLoading" description="暂无数据，请点击查询按钮" />
    </t-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  createDomain,
  updateDomainMetadata,
  queryMyDomains,
  queryMyManagedDomains,
  queryDomainInfo,
} from '@/api/domain'
import { useUserStore } from '@/stores/user'

// 选项卡
const activeTab = ref('create')
const tabList = [
  { value: 'create', label: '创建数据域' },
  { value: 'update', label: '更新元数据' },
  { value: 'myDomains', label: '我参与的数据域' },
  { value: 'managedDomains', label: '我管理的数据域' },
  { value: 'domainInfo', label: '查询详细信息' },
]

const onTabChange = (value) => {
  activeTab.value = value
}

// 创建数据域表单
const createFormRef = ref(null)
const createForm = reactive({
  domainName: '',
})
const createRules = {
  domainName: [{ required: true, message: '请输入数据域名称', type: 'error' }],
}
const createLoading = ref(false)

const handleCreateDomain = async ({ validateResult }) => {
  if (validateResult !== true) {
    return
  }
  createLoading.value = true
  try {
    const res = await createDomain(createForm.domainName)
    if (res.code === 0) {
      MessagePlugin.success(res.msg || '创建数据域成功')
      resetCreateForm()
    } else {
      MessagePlugin.error(res.msg || '创建数据域失败')
    }
  } catch (error) {
    MessagePlugin.error('创建数据域失败：' + (error.message || '未知错误'))
  } finally {
    createLoading.value = false
  }
}

const resetCreateForm = () => {
  createFormRef.value?.reset()
  createForm.domainName = ''
}

// 更新数据域元数据表单
const updateFormRef = ref(null)
const updateForm = reactive({
  domainName: '',
  domainMembers: '',
  domainPolicy: '',
})
const updateRules = {
  domainName: [{ required: true, message: '请输入数据域名称', type: 'error' }],
  domainMembers: [{ required: false }],
  domainPolicy: [{ required: false }],
}
const updateLoading = ref(false)

const handleUpdateDomain = async ({ validateResult }) => {
  if (validateResult !== true) {
    return
  }
  updateLoading.value = true
  try {
    // 处理数据域成员：将逗号分隔的字符串转换为JSON数组字符串
    let domainMembersJson = ''
    if (updateForm.domainMembers && updateForm.domainMembers.trim()) {
      // 支持英文逗号和中文逗号分隔
      const membersArray = updateForm.domainMembers
        .split(/[,，]/) // 使用正则表达式同时匹配英文逗号和中文逗号
        .map(member => member.trim()) // 去除每个成员的前后空格
        .filter(member => member.length > 0) // 过滤掉空字符串
      
      // 将数组转换为JSON字符串
      domainMembersJson = JSON.stringify(membersArray)
    }
    
    const res = await updateDomainMetadata(
      updateForm.domainName,
      domainMembersJson,
      updateForm.domainPolicy
    )
    if (res.code === 0) {
      MessagePlugin.success(res.msg || '更新数据域元数据成功')
      resetUpdateForm()
    } else {
      MessagePlugin.error(res.msg || '更新数据域元数据失败')
    }
  } catch (error) {
    MessagePlugin.error('更新数据域元数据失败：' + (error.message || '未知错误'))
  } finally {
    updateLoading.value = false
  }
}

const resetUpdateForm = () => {
  updateFormRef.value?.reset()
  updateForm.domainName = ''
  updateForm.domainMembers = ''
  updateForm.domainPolicy = ''
}

// 查询我参与的数据域
const queryMyDomainsLoading = ref(false)
const myDomainsResult = ref(null)

const handleQueryMyDomains = async () => {
  queryMyDomainsLoading.value = true
  try {
    const res = await queryMyDomains()
    if (res.code === 0) {
      // 解析返回的数据
      let domainData = res.data
      
      // 如果返回的是字符串（JSON字符串），需要先解析
      if (typeof domainData === 'string') {
        try {
          domainData = JSON.parse(domainData)
        } catch (e) {
          console.error('解析JSON失败:', e)
          MessagePlugin.error('解析返回数据失败')
          myDomainsResult.value = null
          return
        }
      }
      
      // 如果返回的是字符串数组（只有数据域名称），转换为对象数组
      if (Array.isArray(domainData)) {
        if (domainData.length > 0 && typeof domainData[0] === 'string') {
          // 是字符串数组，需要为每个名称查询详细信息
          const domainList = []
          const userStore = useUserStore()
          
          // 批量查询每个数据域的详细信息
          for (const domainName of domainData) {
            try {
              const infoRes = await queryDomainInfo(domainName)
              if (infoRes.code === 0 && infoRes.data) {
                let infoData = infoRes.data
                // 如果返回的是字符串，需要解析
                if (typeof infoData === 'string') {
                  try {
                    infoData = JSON.parse(infoData)
                  } catch (e) {
                    console.error('解析数据域信息失败:', e)
                  }
                }
                domainList.push({
                  domainName: infoData.domainName || domainName,
                  domainMembers: infoData.domainMembers || '-',
                  domainPolicy: infoData.domainPolicy || '-',
                  orgId: infoData.orgId || userStore.orgId || '-',
                  role: infoData.role || userStore.role || '-',
                })
              } else {
                // 如果查询详细信息失败，至少显示名称
                domainList.push({
                  domainName: domainName,
                  domainMembers: '-',
                  domainPolicy: '-',
                  orgId: userStore.orgId || '-',
                  role: userStore.role || '-',
                })
              }
            } catch (error) {
              console.error(`查询数据域 ${domainName} 详细信息失败:`, error)
              // 即使查询失败，也显示基本信息
              domainList.push({
                domainName: domainName,
                domainMembers: '-',
                domainPolicy: '-',
                orgId: userStore.orgId || '-',
                role: userStore.role || '-',
              })
            }
          }
          myDomainsResult.value = domainList
        } else {
          // 已经是对象数组，直接使用
          myDomainsResult.value = domainData
        }
      } else {
        myDomainsResult.value = []
      }
      MessagePlugin.success(res.msg || '查询成功')
    } else {
      MessagePlugin.error(res.msg || '查询失败')
      myDomainsResult.value = null
    }
  } catch (error) {
    MessagePlugin.error('查询失败：' + (error.message || '未知错误'))
    myDomainsResult.value = null
  } finally {
    queryMyDomainsLoading.value = false
  }
}

// 查询我管理的数据域
const queryManagedDomainsLoading = ref(false)
const managedDomainsResult = ref(null)

const handleQueryManagedDomains = async () => {
  queryManagedDomainsLoading.value = true
  try {
    const res = await queryMyManagedDomains()
    if (res.code === 0) {
      // 解析返回的数据
      let domainData = res.data
      
      // 如果返回的是字符串（JSON字符串），需要先解析
      if (typeof domainData === 'string') {
        try {
          domainData = JSON.parse(domainData)
        } catch (e) {
          console.error('解析JSON失败:', e)
          MessagePlugin.error('解析返回数据失败')
          managedDomainsResult.value = null
          return
        }
      }
      
      // 如果返回的是字符串数组（只有数据域名称），转换为对象数组
      if (Array.isArray(domainData)) {
        if (domainData.length > 0 && typeof domainData[0] === 'string') {
          // 是字符串数组，需要为每个名称查询详细信息
          const domainList = []
          const userStore = useUserStore()
          
          // 批量查询每个数据域的详细信息
          for (const domainName of domainData) {
            try {
              const infoRes = await queryDomainInfo(domainName)
              if (infoRes.code === 0 && infoRes.data) {
                let infoData = infoRes.data
                // 如果返回的是字符串，需要解析
                if (typeof infoData === 'string') {
                  try {
                    infoData = JSON.parse(infoData)
                  } catch (e) {
                    console.error('解析数据域信息失败:', e)
                  }
                }
                domainList.push({
                  domainName: infoData.domainName || domainName,
                  domainMembers: infoData.domainMembers || '-',
                  domainPolicy: infoData.domainPolicy || '-',
                  orgId: infoData.orgId || userStore.orgId || '-',
                  role: infoData.role || userStore.role || '-',
                })
              } else {
                // 如果查询详细信息失败，至少显示名称
                domainList.push({
                  domainName: domainName,
                  domainMembers: '-',
                  domainPolicy: '-',
                  orgId: userStore.orgId || '-',
                  role: userStore.role || '-',
                })
              }
            } catch (error) {
              console.error(`查询数据域 ${domainName} 详细信息失败:`, error)
              // 即使查询失败，也显示基本信息
              domainList.push({
                domainName: domainName,
                domainMembers: '-',
                domainPolicy: '-',
                orgId: userStore.orgId || '-',
                role: userStore.role || '-',
              })
            }
          }
          managedDomainsResult.value = domainList
        } else {
          // 已经是对象数组，直接使用
          managedDomainsResult.value = domainData
        }
      } else {
        managedDomainsResult.value = []
      }
      MessagePlugin.success(res.msg || '查询成功')
    } else {
      MessagePlugin.error(res.msg || '查询失败')
      managedDomainsResult.value = null
    }
  } catch (error) {
    MessagePlugin.error('查询失败：' + (error.message || '未知错误'))
    managedDomainsResult.value = null
  } finally {
    queryManagedDomainsLoading.value = false
  }
}

// 查询数据域详细信息
const queryInfoFormRef = ref(null)
const queryInfoForm = reactive({
  domainName: '',
})
const queryInfoRules = {
  domainName: [{ required: true, message: '请输入数据域名称', type: 'error' }],
}
const queryInfoLoading = ref(false)
const domainInfoResult = ref(null)

const domainInfoDescriptions = computed(() => {
  if (!domainInfoResult.value) {
    console.log('domainInfoResult.value 为空')
    return []
  }
  
  const result = domainInfoResult.value
  console.log('计算 domainInfoDescriptions, result:', result)
  
  // TDesign 的 t-descriptions 组件使用 label 和 content 字段
  const descriptions = [
    { label: '数据域名称', content: result.domainName || '-' },
    { label: '数据域成员', content: result.domainMembers || '-' },
    { label: '数据域策略', content: '见下方详情' }, // 策略在下方单独显示
  ]
  
  console.log('生成的 descriptions:', descriptions)
  return descriptions
})

const handleQueryDomainInfo = async ({ validateResult }) => {
  if (validateResult !== true) {
    return
  }
  queryInfoLoading.value = true
  domainInfoResult.value = null // 先清空之前的结果
  try {
    const res = await queryDomainInfo(queryInfoForm.domainName)
    console.log('查询返回的完整数据:', res) // 调试日志
    if (res.code === 0) {
      // 解析返回的数据
      let domainData = res.data
      console.log('原始 data:', domainData, '类型:', typeof domainData) // 调试日志
      
      // 如果返回的是字符串（JSON字符串），需要先解析
      if (typeof domainData === 'string') {
        try {
          domainData = JSON.parse(domainData)
          console.log('解析后的数据:', domainData) // 调试日志
        } catch (e) {
          console.error('解析JSON失败:', e, '原始数据:', domainData)
          MessagePlugin.error('解析返回数据失败')
          domainInfoResult.value = null
          return
        }
      }
      
      // 如果解析后是对象，直接使用
      if (domainData && typeof domainData === 'object' && !Array.isArray(domainData)) {
        // 处理members数组，转换为逗号分隔的字符串
        let domainMembers = '-'
        if (domainData.members) {
          if (Array.isArray(domainData.members)) {
            domainMembers = domainData.members.join(', ')
          } else if (typeof domainData.members === 'string') {
            domainMembers = domainData.members
          }
        }
        
        // 处理accessPolicy对象，格式化为JSON字符串
        let domainPolicy = '-'
        if (domainData.accessPolicy) {
          try {
            // 格式化JSON，使用2个空格缩进
            domainPolicy = JSON.stringify(domainData.accessPolicy, null, 2)
          } catch (e) {
            console.error('序列化策略失败:', e)
            domainPolicy = String(domainData.accessPolicy)
          }
        }
        
        // 设置结果数据
        domainInfoResult.value = {
          domainName: domainData.name || domainData.domainName || queryInfoForm.domainName || '-',
          domainMembers: domainMembers,
          domainPolicy: domainPolicy,
        }
        console.log('最终设置的数据:', domainInfoResult.value) // 调试日志
        MessagePlugin.success(res.msg || '查询成功')
      } else {
        // 如果解析后不是对象
        console.warn('解析后的数据不是对象:', domainData)
        MessagePlugin.error('返回数据格式不正确')
        domainInfoResult.value = null
      }
    } else {
      MessagePlugin.error(res.msg || '查询失败')
      domainInfoResult.value = null
    }
  } catch (error) {
    console.error('查询异常:', error)
    MessagePlugin.error('查询失败：' + (error.message || '未知错误'))
    domainInfoResult.value = null
  } finally {
    queryInfoLoading.value = false
  }
}

const resetQueryInfoForm = () => {
  queryInfoFormRef.value?.reset()
  queryInfoForm.domainName = ''
  domainInfoResult.value = null
}

// 表格列配置
const domainTableColumns = [
  { colKey: 'domainName', title: '数据域名称', width: 200 },
  { colKey: 'domainMembers', title: '数据域成员', width: 200 },
  { colKey: 'domainPolicy', title: '数据域策略', width: 200 },
]
</script>

<style scoped>
.domain-view {
  margin: 20px auto;
  width: 80%;
  max-width: 1200px;
}

.operation-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-result {
  margin-top: 24px;
}

.domain-info-display {
  padding: 8px 0;
}

.info-item {
  margin-bottom: 16px;
  display: flex;
  align-items: flex-start;
}

.info-label {
  font-weight: 500;
  color: #333;
  min-width: 120px;
  margin-right: 12px;
}

.info-value {
  color: #666;
  flex: 1;
  word-break: break-word;
}

.policy-section {
  margin-top: 12px;
}

.policy-content {
  background-color: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 400px;
  overflow-y: auto;
  margin: 0;
  border: 1px solid #e0e0e0;
}

:deep(.t-form-item) {
  margin-bottom: 20px;
}
</style>


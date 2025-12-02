<template>
  <t-form
    :rules="rules"
    :data="form"
    ref="formRef"
    label-width="120px"
    @submit="handleSubmit"
    colon
    labelAlign="right"
  >
    <t-form-item label="数据域名称" name="domainName" requiredMark>
      <t-input
        v-model="form.domainName"
        placeholder="请输入数据域名称（必填）"
        clearable
      />
    </t-form-item>

    <t-form-item label="姓名" name="name">
      <t-input
        v-model="form.name"
        placeholder="请输入姓名（选填）"
        clearable
      />
    </t-form-item>

    <t-form-item label="年龄" name="ageStart">
      <div style="display: flex; align-items: center; gap: 10px">
        <t-input-number
          v-model="form.ageStart"
          placeholder="起始年龄（选填）"
          :min="0"
          :max="150"
          clearable
          style="flex: 1"
        />
        <span>至</span>
        <t-input-number
          v-model="form.ageEnd"
          placeholder="结束年龄（选填）"
          :min="0"
          :max="150"
          clearable
          style="flex: 1"
        />
      </div>
      <div style="margin-top: 8px; color: #999; font-size: 12px">
        提示：只填写起始年龄表示精确匹配该年龄；填写范围表示查询该年龄范围内的数据
      </div>
    </t-form-item>

    <t-form-item label="性别" name="gender">
      <t-input
        v-model="form.gender"
        placeholder="请输入性别（选填）"
        clearable
      />
    </t-form-item>

    <t-form-item label="医院" name="hospital">
      <t-input
        v-model="form.hospital"
        placeholder="请输入医院（选填）"
        clearable
      />
    </t-form-item>

    <t-form-item label="科室" name="department">
      <t-input
        v-model="form.department"
        placeholder="请输入科室（选填）"
        clearable
      />
    </t-form-item>

    <t-form-item label="疾病代码" name="diseaseCode">
      <t-input
        v-model="form.diseaseCode"
        placeholder="请输入疾病代码（选填）"
        clearable
      />
    </t-form-item>

    <t-form-item>
      <t-button theme="primary" type="submit">查询</t-button>
      <t-button
        theme="default"
        variant="base"
        type="reset"
        style="margin-left: 10px"
        >重置</t-button
      >
    </t-form-item>
  </t-form>
</template>

<script setup>
import { reactive, ref, defineEmits } from "vue";
import { storeToRefs } from "pinia";
import { MessagePlugin } from "tdesign-vue-next";
import { useApiStore } from "@/stores/api.js";
import { useUserStore } from "@/stores/user.js";

const apiStore = useApiStore();
const userStore = useUserStore();
const { userId, orgId, role } = storeToRefs(userStore);

const formRef = ref(null);
const form = reactive({
  domainName: "",
  name: "",
  ageStart: undefined,
  ageEnd: undefined,
  gender: "",
  hospital: "",
  department: "",
  diseaseCode: "",
});

const rules = {
  domainName: [
    { required: true, message: "数据域名称为必填项", type: "error" },
  ],
};

const emit = defineEmits(["submitQuery"]);

const handleSubmit = ({ validateResult, firstError, e }) => {
  e.preventDefault(); // 阻止表单默认提交行为
  if (validateResult === true) {
    // 构建查询参数，包含必填的基础信息
    const queryParams = {
      uId: userId.value,
      apiUrl: {
        ipfsServiceUrl: apiStore.ipfsServiceUrl,
        chainServiceUrl: apiStore.chainServiceUrl,
        contractName: apiStore.contractName,
      },
      domainName: form.domainName.trim(),
      orgId: orgId.value,
      role: role.value,
    };

    // 只添加非空字段
    if (form.name && form.name.trim()) {
      queryParams.name = form.name.trim();
    }
    // 处理年龄范围
    if (form.ageStart !== undefined && form.ageStart !== null && form.ageStart > 0) {
      queryParams.ageStart = form.ageStart;
      // 如果只填写了起始年龄，则作为精确匹配
      if (form.ageEnd !== undefined && form.ageEnd !== null && form.ageEnd > 0) {
        queryParams.ageEnd = form.ageEnd;
      } else {
        // 只填写起始年龄，作为精确匹配
        queryParams.ageEnd = form.ageStart;
      }
    } else if (form.ageEnd !== undefined && form.ageEnd !== null && form.ageEnd > 0) {
      // 只填写了结束年龄，使用0作为起始年龄
      queryParams.ageStart = 0;
      queryParams.ageEnd = form.ageEnd;
    }
    if (form.gender && form.gender.trim()) {
      queryParams.gender = form.gender.trim();
    }
    if (form.hospital && form.hospital.trim()) {
      queryParams.hospital = form.hospital.trim();
    }
    if (form.department && form.department.trim()) {
      queryParams.department = form.department.trim();
    }
    if (form.diseaseCode && form.diseaseCode.trim()) {
      queryParams.diseaseCode = form.diseaseCode.trim();
    }

    emit("submitQuery", queryParams);
  } else {
    MessagePlugin.error("请检查输入项");
  }
};
</script>

<style scoped>
</style>


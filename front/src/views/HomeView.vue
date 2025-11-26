<template>
  <div class="home">
    <!-- <t-typography-title level="h5">
      <span style="color: #0052d9">#1 设置用户ID</span>
    </t-typography-title>
    <div class="input-container input-user-id">
      <t-input v-model="uId" placeholder="请输入用户Id" />
      <t-button @click="saveUserId">设置ID</t-button>
    </div> -->

    <div class="org-role-panel">
      <div class="org-role-inputs">
        <div class="input-block">
          <span class="input-label">组织ID</span>
          <t-input v-model="orgIdInput" placeholder="请输入组织ID" />
        </div>
        <div class="input-block">
          <span class="input-label">角色</span>
          <t-input v-model="roleInput" placeholder="请输入角色" />
        </div>
      </div>
      <t-button class="confirm-btn" theme="primary" @click="saveOrgRole">
        确定
      </t-button>
    </div>

    <t-typography-title level="h5" style="color: #0052d9">
      <span style="color: #0052d9">请选择操作</span>
    </t-typography-title>
    <div class="choose">
      <t-card
        class="option-card"
        hover-shadow
        @click="routeTo('upload')"
        subtitle="上传源文件以供查询使用"
      >
        <template #title>上传信息</template>
      </t-card>
      <t-card
        class="option-card"
        hover-shadow
        @click="routeTo('query')"
        subtitle="对他人上传的数据源进行查询"
      >
        <template #title>查询信息</template>
      </t-card>
    </div>
    <div class="choose">
      <t-card
        class="option-card domain-card"
        hover-shadow
        @click="routeTo('domain')"
        subtitle="数据域相关操作，包括新建、查询、管理"
      >
        <template #title>数据域</template>
      </t-card>
      <t-card
        class="option-card"
        hover-shadow
        @click="routeTo('log')"
        subtitle="查看查询操作的审计日志"
      >
        <template #title>审计日志</template>
      </t-card>
    </div>
    <!-- <t-guide v-model="currentGuide" :steps="GuideSteps" :show-overlay="false" /> -->
    <UserIdCheckAndSetDialog ref="UserIdCheckAndSetDialogRef" />
  </div>
</template>


<script setup>
import { onMounted, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useUserStore } from "../stores/user";
import router from "../router";
import { MessagePlugin } from "tdesign-vue-next";
import UserIdCheckAndSetDialog from "../components/UserIdCheckAndSetDialog.vue";

const userStore = useUserStore();
const { orgId, role } = storeToRefs(userStore);
const uId = ref("");
const orgIdInput = ref(orgId.value);
const roleInput = ref(role.value);

const saveUserId = () => {
  userStore.setUserId(uId.value);
};

const saveOrgRole = () => {
  userStore.setOrgRole(orgIdInput.value.trim(), roleInput.value.trim());
  MessagePlugin.success("组织ID与角色已更新");
};

const routeTo = (name_) => {
  // 判断是否设置了用户ID
  if (name_ != "download" && name_ != "log" && !userStore.userId) {
    MessagePlugin.error("此操作需先设置用户ID");
    // currentGuide.value = 0;
    return;
  }
  router.push({ name: name_ });
};

// // guide
// const currentGuide = ref(-1);
// const GuideSteps = [
//   {
//     element: ".input-user-id",
//     title: "设置用户ID",
//     body: "查询操作和上传操作均需先设置用户ID",
//     placement: "bottom-right",
//   },
// ];

const UserIdCheckAndSetDialogRef = ref(null);
onMounted(() => {
  UserIdCheckAndSetDialogRef.value.checkAndShowUserIdSetDialog();
});

watch(
  [orgId, role],
  ([newOrgId, newRole]) => {
    orgIdInput.value = newOrgId;
    roleInput.value = newRole;
  },
  { immediate: true }
);
</script>

<style scoped>
.home {
  margin: 20px auto;
  width: 30%;
}

.org-role-panel {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
  margin-bottom: 20px;
}

.org-role-inputs {
  display: flex;
  gap: 16px;
  width: 100%;
}

.input-block {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.input-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 6px;
}

.confirm-btn {
  align-self: flex-start;
}

.input-container {
  display: flex;
  justify-content: center;
  margin: 20px auto;
  width: 100%;
}

.choose {
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  text-align: left;
  margin-top: 20px;
  gap: 20px;
}

.option-card {
  width: 200px;
  cursor: pointer;
  text-align: center;
}

.domain-card {
  display: flex;
}

.domain-card :deep(.t-card__wrapper) {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
</style>
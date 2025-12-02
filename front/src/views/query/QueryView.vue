<template>
  <div class="query-view">
    <t-typography-title level="h5">
      <span style="color: #0052d9">查询数据</span>
    </t-typography-title>
    <SimpleQueryForm @submitQuery="submitQuery" v-show="status === 'query'" />
    <SimpleQueryResult ref="queryResultRef" v-show="status === 'result'" />
    <UserIdCheckAndSetDialog ref="UserIdCheckAndSetDialogRef" />
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import SimpleQueryForm from "./components/SimpleQueryForm.vue";
import SimpleQueryResult from "./components/SimpleQueryResult.vue";
import UserIdCheckAndSetDialog from "@/components/UserIdCheckAndSetDialog.vue";

const status = ref("query"); //query or result

const queryResultRef = ref();
const submitQuery = (queryParams) => {
  status.value = "result";
  console.log("查询参数:", queryParams);
  queryResultRef.value.queryAndShowResult(queryParams);
};

const UserIdCheckAndSetDialogRef = ref(null);
onMounted(() => {
  UserIdCheckAndSetDialogRef.value.checkAndShowUserIdSetDialog();
});
</script>

<style scoped>
.query-view {
  margin: 20px auto;
  width: 60%;
}
</style>
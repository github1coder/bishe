<template>
  <div class="simple-query-result">
    <div
      style="margin: 20px auto"
      v-if="dataStatus === 'loading'"
      class="load-card"
    >
      <t-loading
        style="margin-left: 50%"
        size="large"
        text="正在查询中..."
      ></t-loading>
    </div>

    <!-- 错误 -->
    <div v-else-if="dataStatus === 'error'" class="error-card">
      <t-alert
        theme="warning"
        title="查询出错了"
        :message="
          errMsg +
          ' 。若提示500，可能是未上传数字信封。请再次确认ipfs地址正确。'
        "
      >
        <template #operation>
          <span @click="handleRetry">点击重试</span>
        </template>
      </t-alert>
    </div>

    <!-- 成功 -->
    <div v-else-if="dataStatus === 'success'">
      <t-alert
        theme="success"
        title="查询成功"
        :message="successData.counts + '条数据'"
      >
        <template #operation>
          <span @click="handleRetry">重新查询</span>
        </template>
      </t-alert>

      <QueryResultTable ref="queryResultTableRef"></QueryResultTable>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, defineExpose, nextTick } from "vue";
import { queryByFields } from "@/api/query";
import { MessagePlugin } from "tdesign-vue-next";
import QueryResultTable from "./QueryResultTable.vue";

const dataStatus = ref(""); // loading, error, success
const errMsg = ref("");
const resultJson = ref("");
const successData = reactive({
  counts: 0,
  data: [],
});

const queryResultTableRef = ref(null);
const currentQueryParams = ref(null);

const queryAndShowResult = async (queryParams) => {
  dataStatus.value = "loading";
  errMsg.value = "";
  currentQueryParams.value = queryParams;

  try {
    const resp = await queryByFields(queryParams).catch((err) => {
      dataStatus.value = "error";
      console.log(err);
      errMsg.value = "查询失败！原因 —— " + err?.response?.data?.msg;
      MessagePlugin.error("查询失败！");
    });

    if (!resp) {
      return;
    }

    resultJson.value = resp;
    if (resp.code === 0) {
      if (resp.msg.includes("出现错误")) {
        dataStatus.value = "error";
        MessagePlugin.warning(resp.data);
        errMsg.value = resp.data;
        return;
      } else {
        dataStatus.value = "success";
        // 处理查询结果，将结果存储到successData中
        // resp.data 是一个 JSON 字符串，需要解析
        const parsedData = typeof resp.data === 'string' ? JSON.parse(resp.data) : resp.data;
        successData.counts = parsedData.counts || 0; // 查询结果总数
        successData.data = parsedData.data || []; // 查询结果数据
        showQueryResultTable(successData.counts, successData.data);
        MessagePlugin.success(parsedData.message || resp.msg);
      }
    } else {
      dataStatus.value = "error";
      errMsg.value = "查询失败！原因未知";
      MessagePlugin.error("查询失败！");
    }
  } catch (error) {
    errMsg.value = error.message || "查询失败";
    dataStatus.value = "error";
    MessagePlugin.error(errMsg.value);
  }
};

const showQueryResultTable = (counts, data) => {
  nextTick(() => {
    queryResultTableRef.value.ShowQueryResultTable(counts, data);
  });
};

const handleRetry = () => {
  // 重试查询
  if (currentQueryParams.value) {
    queryAndShowResult(currentQueryParams.value);
  } else {
    dataStatus.value = "";
  }
};

defineExpose({
  queryAndShowResult,
});
</script>

<style scoped>
.simple-query-result {
  margin-top: 20px;
}
</style>


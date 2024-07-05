<script setup>
import { ref } from "vue";
import { usePipelinesStore } from "@/stores";

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
async function loadData() {
    const store = usePipelinesStore();
    var res = await store.list("all", pageSize.value, currentPage.value);
    dataList.value = res.list
    total.value = res.total
}
loadData();
function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

function formatObject(row, column, cellValue) {
    if (typeof cellValue === 'object') {
        return JSON.stringify(cellValue, null, 4);
    }
    return cellValue;
}

</script>

<template>
    <div class="container">
        <el-table :data="dataList" border size="default">
            <el-table-column prop="metadata.namespace" label="Namespace" />
            <el-table-column prop="metadata.name" label="Name" />
            <el-table-column prop="spec.desc" label="Desc" />
            <el-table-column prop="spec.variables" label="Variables" :formatter="formatObject" />
            <el-table-column prop="spec.tasks" label="Tasks" :formatter="formatObject" />
        </el-table>
    </div>
</template>

<style scoped>
.contaner {
    margin-left: 7em;
}

.form-item {
    display: flex;
    align-items: center;
    margin-bottom: 20px;
}

.label {
    width: 30%;
    text-align: right;
}

.input {
    flex: 1;
    margin-left: 10px;
}
</style>

<script setup>
import { ref } from 'vue';
import { usePipelineRunsStore } from '@/stores';
import { formatObject } from '@/utils/common';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var dialogVisible = ref(false);
var selectedItem = ref(null);
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.pipelineRef', 'spec.variables', 'status.runStatus', 'status.startTime']);

loadData();

async function loadData() {
    const store = usePipelineRunsStore();
    var res = await store.list("all", pageSize.value, currentPage.value);
    dataList.value = res.list
    total.value = res.total
}

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

function view(item) {
    dialogVisible.value = true;
    selectedItem.value = item;
}
</script>

<template>
    <div class="container">
        <el-dialog v-model="dialogVisible" title="TaskRun Details">
            <div class="card-body">
                <div class="form-group">
                    <pre>{{ selectedItem?.status }}</pre>
                </div>
            </div>
        </el-dialog>
        <el-table :data="dataList" border size="default">
            <el-table-column v-for="field in selectedFields" :key="field" :prop="field"
                :label="field.split('.').pop().charAt(0).toUpperCase() + field.split('.').pop().slice(1)">
                <template #default="{ row }">
                    <span v-html="formatObject(row, field)"></span>
                </template>
            </el-table-column>
            <el-table-column label="Actions">
                <template #default="scope">
                    <el-button type="primary" @click="view(scope.row)">View</el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-pagination @current-change="onPaginationChange" @size-change="onPageSizeChange"
            v-model:currentPage="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 30]"
            layout="total, sizes, prev, pager, next" :total="total">
        </el-pagination>
    </div>
</template>

<style scoped>
.search-input {
    margin-bottom: 1em;
    width: 300px;
}

.column-select {
    margin-bottom: 1em;
}

.formatted-variables {
    white-space: normal;
}

.kv-container {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 4px;
}

.kv-pair {
    display: block;
    padding: 4px 8px;
    background-color: #e3f2fd;
    color: #1565c0;
    border-radius: 4px;
    font-family: monospace;
    font-size: 13px;
    border: 1px solid #1565c0;
    width: fit-content;
}

:deep(.el-table .cell) {
    white-space: normal;
}
</style>
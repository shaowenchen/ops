<script setup>
import { useTaskRunsStore } from '@/stores';
import { ref } from 'vue';
import { formatObject } from '@/utils/common';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.taskRef', 'spec.variables', 'status.runStatus', 'status.startTime']);

async function loadData() {
    const store = useTaskRunsStore();
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

var dialogVisible = ref(false);
var selectedItem = ref(null);

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
</style>
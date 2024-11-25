<script setup>
import { ref, computed } from 'vue';
import { useClustersStore } from '@/stores';
import { router } from '../router';
import { formatObject } from '@/utils/common';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
const allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.server', label: 'Server' },
    { value: 'status.version', label: 'Version' },
    { value: 'status.node', label: 'Node' },
    { value: 'status.runningPod', label: 'RunningPod' },
    { value: 'status.pod', label: 'Pod' },
    { value: 'status.certNotAfterDays', label: 'CertNotAfterDays' },
    { value: 'status.heartTime', label: 'HeartTime' },
    { value: 'status.heartStatus', label: 'HeartStatus' }
];

var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'status.version', 'status.node', 'status.certNotAfterDays', 'status.heartStatus', 'status.heartTime']);

async function loadData() {
    const store = useClustersStore();
    var res = await store.list("all", pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list;
    total.value = res.total;
}

loadData();

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

function run(item) {
    router.push({ name: 'cluster-details', params: { cluster: item.metadata.name } });
}
</script>

<template>
    <div class="container">
        <div class="form-control">
            <el-input v-model="searchQuery" placeholder="Search..." @input="loadData" />
            <el-select v-model="selectedFields" multiple placeholder="Select fields to display">
                <el-option v-for="field in allFields" :key="field.value" :label="field.label" :value="field.value" />
            </el-select>
        </div>
        <el-table :data="dataList" border size="default">
            <el-table-column v-for="field in selectedFields" :key="field" :prop="field"
                :label="field.split('.').pop().charAt(0).toUpperCase() + field.split('.').pop().slice(1)">
                <template #default="{ row }">
                    <span v-html="formatObject(row, field)"></span>
                </template>
            </el-table-column>
            <el-table-column label="Nodes">
                <template #default="scope">
                    <el-button type="primary" @click="run(scope.row)">View</el-button>
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
.form-control {
    display: inline-block;
    width: 90%;
    margin-right: 10px;
}
</style>

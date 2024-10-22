<script setup>
import { ref, watch, computed } from 'vue';
import { useClustersStore } from '@/stores';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'status.version', 'status.node', 'status.certNotAfterDays', 'status.heartTime']);

async function loadData() {
    const store = useClustersStore();
    var res = await store.list("all", pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list;
    total.value = res.total;
}

watch(searchQuery, () => {
    currentPage.value = 1;
    loadData();
});

loadData();

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

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
const displayedColumns = computed(() => {
    return allFields.filter(field => selectedFields.value.includes(field.value));
});
</script>

<template>
    <div class="container">
        <div class="form-control">
            <el-input v-model="searchQuery" placeholder="Search..." @input="onSearch" />
            <el-select v-model="selectedFields" multiple placeholder="Select fields to display">
                <el-option v-for="field in allFields" :key="field.value" :label="field.label" :value="field.value" />
            </el-select>
        </div>
        <el-table :data="dataList" border size="default">
            <el-table-column v-for="column in displayedColumns" :key="column.value" :prop="column.value" :label="column.label" />
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

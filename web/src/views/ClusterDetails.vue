<script setup>
import { ref, watch, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useClustersStore } from '@/stores';
import { formatMemory } from '@/utils/cluster';

const { cluster } = useRoute().params;
var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
const allFields = [
    { value: 'metadata.name', label: 'Name' },
    { value: 'status.addresses', label: 'IP' },
    { value: 'status.capacity.cpu', label: 'CPU' },
    { value: 'status.capacity.memory', label: 'Memory' },
    { value: 'status.capacity.pods', label: 'Pods' },
];
var selectedFields = ref(['Name', 'IP', 'CPU', 'Memory', 'Pods']);

loadData();

async function loadData() {
    const store = useClustersStore();
    var res = await store.listNodes("ops-system", cluster, pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list;
    total.value = res.total;
}


function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

function ipFormatter(row) {
    if (row.status && row.status.addresses) {
        const internalIP = row.status.addresses.find(addr => addr.type === 'InternalIP');
        return internalIP ? internalIP.address : 'N/A';
    }
    return 'N/A';
}

function capacityFormatter(row, type) {
    if (row.status && row.status.capacity) {
        const value = row.status.capacity[type];
        if (type === 'memory') {
            return formatMemory(value);
        }
        return value || 'N/A';
    }
    return 'N/A';
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
            <el-table-column v-if="selectedFields.includes('Name')" prop="metadata.name" label="Name" />

            <el-table-column v-if="selectedFields.includes('IP')" label="IP" :formatter="ipFormatter" />

            <el-table-column v-if="selectedFields.includes('CPU')" label="CPU"
                :formatter="(row) => capacityFormatter(row, 'cpu')" />

            <el-table-column v-if="selectedFields.includes('Memory')" label="Memory"
                :formatter="(row) => capacityFormatter(row, 'memory')" />
            <el-table-column v-if="selectedFields.includes('Pods')" label="Pods"
                :formatter="(row) => capacityFormatter(row, 'pods')" />
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

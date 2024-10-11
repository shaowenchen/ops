<script setup>
import { ref, computed } from "vue";
import { usePipelinesStore } from "@/stores";

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'spec.variables', 'spec.tasks']); 
async function loadData() {
    const store = usePipelinesStore();
    var res = await store.list("all", pageSize.value, currentPage.value);
    dataList.value = res.list;
    total.value = res.total;
}

const filteredDataList = computed(() => {
    if (!searchQuery.value) {
        return dataList.value;
    }
    return dataList.value.filter(item => {
        return Object.values(item.metadata).some(value =>
            value.toString().toLowerCase().includes(searchQuery.value.toLowerCase())
        ) || Object.values(item.spec).some(value =>
            value.toString().toLowerCase().includes(searchQuery.value.toLowerCase())
        );
    });
});

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

const allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.variables', label: 'Variables' },
    { value: 'spec.tasks', label: 'Tasks' },
];

</script>

<template>
    <div class="container">
        <el-input v-model="searchQuery" placeholder="Search..." class="search-input" clearable />

        <el-select v-model="selectedFields" multiple placeholder="Select columns to display" class="column-select">
            <el-option v-for="field in allFields" :key="field.value" :label="field.label" :value="field.value" />
        </el-select>

        <el-table :data="filteredDataList" border size="default">
            <el-table-column v-for="field in selectedFields" :key="field" :prop="field"
                :label="field.split('.').pop().charAt(0).toUpperCase() + field.split('.').pop().slice(1)"
                :formatter="field.includes('spec') ? formatObject : null" />
        </el-table>

        <el-pagination @current-change="onPaginationChange" @size-change="onPageSizeChange"
            v-model:currentPage="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 30]"
            layout="total, sizes, prev, pager, next" :total="total">
        </el-pagination>
    </div>
</template>

<style scoped>
.container {
    margin-left: 7em;
}

.search-input {
    margin-bottom: 1em;
    width: 300px;
}

.column-select {
    margin-bottom: 1em;
}
</style>

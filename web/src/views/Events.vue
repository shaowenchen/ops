<script setup>
import { ref, computed, watch } from 'vue';
import { useEventsStore } from '@/stores';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref('ops.*');
var selectedFields = ref(['subject', 'data']);

async function loadData() {
    const store = useEventsStore();
    var res = await store.list(searchQuery.value, pageSize.value, currentPage.value);

    dataList.value = res.list;
    total.value = res.total;

    if (len(dataList.value) > 0) {
        const firstItem = dataList.value[0];
        selectedFields.value = Object.keys(firstItem);
    }
}

watch(searchQuery, (newQuery) => {
    currentPage.value = 1;
    loadData();
});

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

const displayedColumns = computed(() => {
    if (len(dataList.value) === 0) return [];
    const firstItem = dataList.value[0];
    return Object.keys(firstItem).map(key => ({ value: key, label: key }));
});

loadData();
</script>

<template>
    <div class="container">
        <div class="form-control">
            <el-input v-model="searchQuery" placeholder="Search..." />
            <el-select v-model="selectedFields" multiple placeholder="Select fields to display">
                <el-option v-for="field in displayedColumns" :key="field.value" :label="field.label"
                    :value="field.value" />
            </el-select>
        </div>

        <el-table :data="dataList" border size="default">
            <el-table-column v-for="column in selectedFields" :key="column" :prop="column" :label="column" />
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

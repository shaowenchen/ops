<script setup>
import { useEventsStore } from '@/stores';
import { computed, ref, watch } from 'vue';
import { formatObject } from '@/utils/common';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var dialogVisible = ref(false);
var selectedItem = ref(null);
var searchQuery = ref('ops.>');

const allFields = [
    { value: 'event.id', label: 'ID' },
    { value: 'subject', label: 'Subject' },
    { value: 'event.type', label: 'Type' },
    { value: 'event.datacontenttype', label: 'DataContentType' },
    { value: 'event.data', label: 'Data' },
    { value: 'event.time', label: 'Time' }
];

const displayedColumns = computed(() => {
    return allFields.filter(field => selectedFields.value.includes(field.value));
});

var selectedFields = ref(['event.id', 'subject', 'event.type', 'event.datacontenttype', 'event.data', 'event.time']);

function view(item) {
    dialogVisible.value = true;
    selectedItem.value = item;
}

async function loadData() {
    const store = useEventsStore();
    var res = await store.list(searchQuery.value, pageSize.value, currentPage.value);

    dataList.value = res.list;
    total.value = res.total;
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

loadData();
</script>

<template>
    <div class="container">
        <div class="form-control">
            <el-input v-model="searchQuery" placeholder="Search..." />
            <el-select v-model="selectedFields" multiple placeholder="Select fields to display">
                <el-option v-for="field in allFields" :key="field.value" :label="field.label" :value="field.value" />
            </el-select>
        </div>
        <el-dialog v-model="dialogVisible" title="Event Details">
            <div class="card-body">
                <div class="form-group">
                    <pre>{{ selectedItem?.event?.data }}</pre>
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
                    <el-button type="primary" @click="view(scope.row)">Detais</el-button>
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

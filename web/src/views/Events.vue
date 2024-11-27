<script setup>
import { useEventsStore } from '@/stores';
import { onMounted, ref, watch } from 'vue';
import { formatObject } from '@/utils/common';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var dialogVisible = ref(false);
var selectedItem = ref(null);
var searchQuery = ref('ops.>');
var selectTime = ref(Date);
selectTime.value = new Date(new Date() - 5 * 60 * 1000);

const allFields = [
    { value: 'event.id', label: 'ID' },
    { value: 'subject', label: 'Subject' },
    { value: 'event.type', label: 'Type' },
    { value: 'event.datacontenttype', label: 'DataContentType' },
    { value: 'event.time', label: 'Time' }
];

var selectedFields = ref(['event.id', 'subject', 'event.type', 'event.time']);

const filteredSubjects = ref([]);

async function fetchSubjects(query) {
    if (!query) {
        query = "";
    }
    const store = useEventsStore();
    const res = await store.list(query, "9999", "1");
    filteredSubjects.value = [];
    if (!res.list) {
        return;
    }
    for (let i = 0; i < res.list.length; i++) {
        filteredSubjects.value.push({ value: res.list[i], label: res.list[i] });
    }
}

watch(searchQuery, (newQuery) => {
    fetchSubjects(newQuery);
});

function handleSelectChange(selectedValue) {
    loadData();
}

function view(item) {
    dialogVisible.value = true;
    selectedItem.value = item;
}

async function loadData() {
    const store = useEventsStore();
    const utcTime = new Date(Date.UTC(selectTime.value.getUTCFullYear(), selectTime.value.getUTCMonth(), selectTime.value.getUTCDate(), selectTime.value.getUTCHours(), selectTime.value.getUTCMinutes(), selectTime.value.getUTCSeconds()));
    const utcTimestamp = utcTime.getTime();
    var res = await store.get(searchQuery.value, utcTimestamp, pageSize.value, currentPage.value);
    dataList.value = res.list;
    total.value = res.total;
}

fetchSubjects("");
loadData();

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

function fetchSuggestions(query, callback) {
    const results = filteredSubjects.value.filter(item =>
        item.label.toLowerCase().includes(query.toLowerCase())
    );
    callback(results);
}

</script>

<template>
    <div class="container">
        <div class="form-control enhanced-form">
            <el-row :gutter="20" align="middle">
                <el-col :span="10">
                    <el-autocomplete v-model="searchQuery" :fetch-suggestions="fetchSuggestions" placeholder="Search"
                        @select="handleSelectChange" trigger-on-focus class="search-bar"></el-autocomplete>
                </el-col>
                <el-col :span="8">
                    <el-date-picker v-model="selectTime" type="datetime" placeholder="Select Start Time"
                        format="YYYY-MM-DD HH:mm:ss" class="date-picker"></el-date-picker>
                </el-col>
                <el-col :span="6">
                    <el-button type="primary" icon="el-icon-search" @click="loadData" class="search-button">
                        Search
                    </el-button>
                </el-col>
            </el-row>
        </div>

        <div class="form-control" hidden>
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
.container {
    padding: 20px;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
}

.form-control {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 20px;
    width: 100%;
}

.enhanced-form {
    background-color: #f9f9f9;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    max-width: 600px;
    margin: 0 auto;
}

.search-bar {
    width: 100%;
}

.date-picker {
    width: 100%;
}

.search-button {
    width: 100%;
    font-size: 16px;
    font-weight: bold;
}

.el-row {
    display: flex;
    align-items: center;
}


.search-input {
    width: 80%;
}

.search-button {
    margin-left: 10px;
}

.field-select {
    width: 250px;
}

.datetime-picker {
    width: 250px;
}

@media (max-width: 768px) {
    .form-control {
        flex-direction: column;
        align-items: stretch;
    }

    .search-input,
    .search-button,
    .datetime-picker {
        width: 100%;
        margin-left: 0;
    }

    .field-select {
        width: 100%;
    }
}
</style>

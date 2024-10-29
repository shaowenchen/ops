<script setup>
import { ref } from "vue";
import { useHostsStore, useTaskRunsStore, useTasksStore } from "@/stores";
import { proxyVariablesToJsonObject, formatObject } from '@/utils/common';

const taskStore = useTasksStore();
const taskrunStore = useTaskRunsStore();
const hostStore = useHostsStore();

var dataList = ref([]);
var hosts = ref([]);
var host = ref(null);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var dialogVisble = ref(false);
var selectedItem = ref(null);
const allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.host', label: 'Host' },
    { value: 'spec.variables', label: 'Variables' },
];
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'spec.host', 'spec.variables']);

loadData();
getHosts();

async function getHosts() {
    var resp = await hostStore.list("all", 999, 1);
    hosts = resp.list;
}

async function loadData() {
    var res = await taskStore.list("all", pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list;
    total.value = res.total;
}

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

async function confirm() {
    var vars = proxyVariablesToJsonObject(selectedItem.value.spec.variables);
    await taskrunStore.create(selectedItem.value.metadata.namespace, selectedItem.value.metadata.name, vars);
    dialogVisble.value = false;
}

function close() {
    dialogVisble.value = false;
}

async function open() {
    host.value = null;
}

function run(item) {
    selectedItem.value = item;
    dialogVisble.value = true;
}
</script>
<template>
    <div class="container">
        <el-input v-model="searchQuery" placeholder="Search..." class="search-input" clearable @input="loadData"/>

        <el-select v-model="selectedFields" multiple placeholder="Select columns to display">
            <el-option v-for="field in allFields" :key="field.value" :label="field.label" :value="field.value" />
        </el-select>

        <el-dialog title="Create TaskRun" v-model="dialogVisble" width="30%" :before-close="close" @open="open">
            <div class="card-body">
                <div class="form-group">
                    <label>Namespace</label>
                    <input name="name" type="text" disabled :value="selectedItem?.metadata?.namespace"
                        class="form-control" />
                </div>
                <div class="form-group">
                    <label>Name</label>
                    <input name="name" type="text" disabled :value="selectedItem?.metadata?.name"
                        class="form-control" />
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input name="desc" type="text" disabled :value="selectedItem?.spec?.desc" class="form-control" />
                </div>
                <div class="form-group" v-if="selectedItem.spec.host != 'anymaster'">
                    <label>Host</label>
                    <el-select v-model="host">
                        <el-option v-for="item in hosts" :key="item.metadata.name" :label="item.metadata.name"
                            :value="item.metadata.name" />
                    </el-select>
                </div>
                <div class="form-group" v-if="selectedItem.spec.variables">
                    <label>Variables</label>
                    <div class="form-item" v-for="(value, key) in selectedItem.spec.variables" :key="key">
                        <label>{{ key }}:</label>
                        <input type="text" v-model="value.default" class="form-control" />
                    </div>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="close">Cancel</el-button>
                    <el-button type="primary" @click="confirm">Run</el-button>
                </span>
            </template>
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
                    <el-button type="primary" @click="run(scope.row)">Run</el-button>
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
.form-item {
    display: flex;
    align-items: center;
    margin-bottom: 20px;
}

.form-item label {
    width: 40%;
    text-align: right;
    margin-right: 10px;
}

.label {
    width: 30%;
    text-align: right;
}

.input {
    flex: 1;
    margin-left: 10px;
}

.search-input {
    margin-bottom: 1em;
    width: 300px;
}
</style>

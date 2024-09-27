<script setup>
import { useClustersStore, useHostsStore, useTaskRunsStore, useTasksStore } from "@/stores";
import { ref } from "vue";

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
async function loadData() {
    const store = useTasksStore();
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

var dialogVisble = ref(false);
var selectedItem = ref(null);
async function confirm() {
    const store = useTaskRunsStore();
    const vars = {};
    vars['resType'] = selectedItem.spec.resType;
    vars['resName'] = selectedItem.spec.resName;
    await store.create(selectedItem.metadata.namespace, selectedItem.metadata.name, vars);
    dialogVisble.value = false;
}

function close() {
    selectedItem = null;
    dialogVisble.value = false;
}

var hosts = ref([]);
var clusters = ref([]);

function getResNameList() {
    if (selectedItem.spec.resType === 'host') {
        return hosts.value.list
    } else if (selectedItem.spec.resType === 'cluster') {
        return clusters.value.list
    }
    return []
}

async function open() {
    const hostStore = useHostsStore();
    hosts.value = await hostStore.list("all", 999, 1);
    const clusterStore = useClustersStore();
    clusters.value = await clusterStore.list("all", 999, 1);
}

function run(item) {
    selectedItem = item;
    dialogVisble.value = true;
}
</script>

<template>
    <div class="container">
        <el-dialog title="Create TaskRun" v-model="dialogVisble" width="30%" :before-close="close" @open="open">
            <div class="card-body">
                <div class="form-group">
                    <label>Namespace</label>
                    <input name="name" type="text" disabled :value="selectedItem?.metadata?.namespace"
                        class="form-control" />
                </div>
                <div class="form-group">
                    <label>Name</label>
                    <input name="name" type="text" disabled :value="selectedItem?.metadata?.name" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input name="desc" type="text" disabled :value="selectedItem?.spec?.desc" class="form-control" />
                </div>
                <div class="form-group">
                    <label>ResType</label>
                    <el-select v-model="selectedItem.spec.resType" class="w-100" placeholder="Select">
                        <el-option label="Host" value="host" />
                        <el-option label="Cluster" value="cluster" />
                    </el-select>

                </div>
                <div class="form-group">
                    <label>ResName</label>
                    <el-select v-model="selectedItem.spec.resName" v-if="selectedItem.spec.resType">
                        <el-option v-for="item in getResNameList()" :key="item.metadata.name" :value="item.metadata.name" />
                    </el-select>
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
            <el-table-column prop="metadata.namespace" label="Namespace" />
            <el-table-column prop="metadata.name" label="Name" />
            <el-table-column prop="spec.crontab" label="Crontab" />
            <el-table-column prop="spec.resType" label="ResType" />
            <el-table-column prop="spec.resName" label="ResName" />
            <el-table-column prop="spec.nodeName" label="NodeName" />
            <el-table-column prop="status.startTime" label="Start Time" />
            <el-table-column prop="status.runStatus" label="Run Status" />
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
.contaner {
    margin-left: 7em;
}

.form-item {
    display: flex;
    align-items: center;
    margin-bottom: 20px;
}

.label {
    width: 30%;
    text-align: right;
}

.input {
    flex: 1;
    margin-left: 10px;
}
</style>

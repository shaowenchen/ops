<script setup>
import { useClustersStore, useHostsStore, useTaskRunsStore, useTasksStore } from "@/stores";
import { ref, computed } from "vue";

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'spec.cluster', 'spec.host']);

async function loadData() {
    const store = useTasksStore();
    var res = await store.list("all", pageSize.value, currentPage.value, searchQuery.value);
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

var dialogVisble = ref(false);
var selectedItem = ref(null);
async function confirm() {
    const store = useTaskRunsStore();
    const vars = {};
    vars['cluster'] = selectedItem.spec.cluster;
    vars['host'] = selectedItem.spec.host;
    await store.create(selectedItem.metadata.namespace, selectedItem.metadata.name, vars);
    dialogVisble.value = false;
}

function close() {
    selectedItem.value = null;
    dialogVisble.value = false;
}

var hosts = ref([]);
var clusters = ref([]);

function getHostList() {
    if (selectedItem.spec.cluster === 'host') {
        return hosts.value.list;
    } else if (selectedItem.spec.cluster === 'cluster') {
        return clusters.value.list;
    }
    return [];
}

async function open() {
    const hostStore = useHostsStore();
    hosts.value = await hostStore.list("all", 999, 1);
    const clusterStore = useClustersStore();
    clusters.value = await clusterStore.list("all", 999, 1);
}

function run(item) {
    selectedItem.value = item;
    dialogVisble.value = true;
}

const allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.cluster', label: 'Cluster' },
    { value: 'spec.host', label: 'Host' },
];

</script>

<template>
    <div class="container">
        <el-input
            v-model="searchQuery"
            placeholder="Search..."
            class="search-input"
            clearable
        />

        <el-select v-model="selectedFields" multiple placeholder="Select columns to display">
            <el-option
                v-for="field in allFields"
                :key="field.value"
                :label="field.label"
                :value="field.value"
            />
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
                    <input name="name" type="text" disabled :value="selectedItem?.metadata?.name" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input name="desc" type="text" disabled :value="selectedItem?.spec?.desc" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Cluster</label>
                    <el-select v-model="selectedItem.spec.cluster" class="w-100" placeholder="Select">
                        <el-option label="Cluster" value="cluster" />
                        <el-option label="Host" value="host" />
                    </el-select>
                </div>
                <div class="form-group">
                    <label>Host</label>
                    <el-select v-model="selectedItem.spec.host" v-if="selectedItem.spec.cluster">
                        <el-option v-for="item in getHostList()" :key="item.metadata.name" :value="item.metadata.name" />
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

        <el-table :data="filteredDataList" border size="default">
            <el-table-column v-for="field in selectedFields" :key="field" :prop="field" :label="field.split('.').pop().charAt(0).toUpperCase() + field.split('.').pop().slice(1)" />
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
.container {
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

.search-input {
    margin-bottom: 1em;
    width: 300px;
}
</style>

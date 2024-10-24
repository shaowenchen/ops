<script setup>
import { ref, computed } from "vue";
import { usePipelinesStore, useClustersStore, usePipelineRunsStore } from "@/stores";
import { proxyVariablesToJsonObject, formatObject } from "@/utils/common";

const clusterStore = useClustersStore();
var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.variables', label: 'Variables' },
    { value: 'spec.tasks', label: 'Tasks' },
];
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'spec.variables', 'spec.tasks']);
var dialogVisble = ref(false);
var selectedItem = ref(null);
var clusters = ref([]);
var hosts = ref([]);
var cluster = ref(null);
var host = ref(null);

loadData();

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

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

async function open() {
    var resp = await clusterStore.list("all", 999, 1);
    clusters.value = resp.list;
}

function run(item) {
    selectedItem.value = item;
    dialogVisble.value = true;
}

async function onClusterChange() {
    var resp = await clusterStore.listNodes("ops-system", cluster.value, 999, 1);
    hosts.value = resp.list;
    host.value = null;
}

function close() {
    dialogVisble.value = false;
}

async function confirm() {
    const store = usePipelineRunsStore();
    var vars = proxyVariablesToJsonObject(selectedItem.value.spec.variables);
    vars['cluster'] = cluster.value;
    vars['host'] = host.value;
    await store.create(selectedItem.value.metadata.namespace, selectedItem.value.metadata.name, vars);
    dialogVisble.value = false;
}
</script>

<template>
    <div class="container">
        <el-input v-model="searchQuery" placeholder="Search..." class="search-input" clearable />

        <el-select v-model="selectedFields" multiple placeholder="Select columns to display" class="column-select">
            <el-option v-for="field in allFields" :key="field.value" :label="field.label" :value="field.value" />
        </el-select>

        <el-dialog title="Create PipelineRun" v-model="dialogVisble" width="30%" :before-close="close" @open="open">
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
                <div class="form-group" v-if="'cluster' in selectedItem.spec.variables">
                    <label>Cluster</label>
                    <el-select v-model="cluster" @change="onClusterChange">
                        <el-option v-for="item in clusters" :key="item.metadata.name" :label="item.metadata.name"
                            :value="item.metadata.name" />
                    </el-select>
                </div>
                <div class="form-group" v-if="'host' in selectedItem.spec.variables">
                    <label>Host</label>
                    <el-select v-model="host">
                        <el-option v-for="item in hosts" :key="item.metadata.name" :label="item.metadata.name"
                            :value="item.metadata.name" />
                    </el-select>
                </div>
                <div class="form-group" v-if="selectedItem.spec.variables">
                    <label>Variables</label>
                    <div
                        v-if="Object.entries(selectedItem.spec.variables).filter(([k, v]) => k !== 'host' && k !== 'cluster').length > 0">
                        <div class="form-item"
                            v-for="([key, value]) in Object.entries(selectedItem.spec.variables).filter(([k, v]) => k !== 'host' && k !== 'cluster')"
                            :key="key">
                            <label>{{ key }}:</label>
                            <input type="text" v-model="value.default" class="form-control" />
                        </div>
                    </div>
                    <div v-else>
                        none
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

        <el-table :data="filteredDataList" border size="default">
            <el-table-column v-for="field in selectedFields" :key="field" :prop="field"
                :label="field.split('.').pop().charAt(0).toUpperCase() + field.split('.').pop().slice(1)" >
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
.search-input {
    margin-bottom: 1em;
    width: 300px;
}

.column-select {
    margin-bottom: 1em;
}
</style>

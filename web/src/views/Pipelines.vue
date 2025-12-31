<script setup>
import { ref, watch, onMounted } from "vue";
import { usePipelinesStore, useClustersStore, usePipelineRunsStore, useNamespacesStore } from "@/stores";
import { proxyVariablesToJsonObject, formatObject } from "@/utils/common";

const clusterStore = useClustersStore();
const pipelineStore = usePipelinesStore();
const namespacesStore = useNamespacesStore();
var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var namespaces = ref([]);
var allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.variables', label: 'Variables' },
    { value: 'spec.tasks', label: 'Tasks' },
];
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'spec.variables', 'spec.tasks']);
var dialogVisble = ref(false);
var viewDialogVisible = ref(false);
var editDialogVisible = ref(false);
var selectedItem = ref(null);
var viewItem = ref(null);
var editItem = ref(null);
var editItemJson = ref("");
var createDialogVisible = ref(false);
var createItem = ref(null);
var createItemJson = ref("");
var clusters = ref([]);
var hosts = ref([]);
var cluster = ref(null);
var host = ref(null);

async function loadNamespaces() {
    try {
        await namespacesStore.list();
        namespaces.value = namespacesStore.namespaces;
    } catch (error) {
        console.error("Error loading namespaces:", error);
        namespaces.value = ['ops-system'];
    }
}

async function loadData() {
    var res = await pipelineStore.list(namespacesStore.selectedNamespace, pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list;
    total.value = res.total;
}

function onNamespaceChange() {
    currentPage.value = 1;
    loadData();
}

onMounted(() => {
    loadNamespaces();
    loadData();
});

watch(() => namespacesStore.selectedNamespace, () => {
    onNamespaceChange();
});

watch(editItem, (newVal) => {
    if (newVal) {
        editItemJson.value = JSON.stringify(newVal, null, 2);
    }
}, { deep: true });

watch(createItem, (newVal) => {
    if (newVal) {
        createItemJson.value = JSON.stringify(newVal, null, 2);
    }
}, { deep: true });

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

async function open() {
    var resp = await clusterStore.list(namespacesStore.selectedNamespace, 999, 1);
    clusters.value = resp.list;
}

function run(item) {
    selectedItem.value = item;
    dialogVisble.value = true;
}

async function onClusterChange() {
    var resp = await clusterStore.listNodes("ops-system", cluster.value, 999, 1);
    hosts.value = resp.list;
}

function close() {
    dialogVisble.value = false;
}

function closeView() {
    viewDialogVisible.value = false;
    viewItem.value = null;
}

function closeEdit() {
    editDialogVisible.value = false;
    editItem.value = null;
}

async function confirm() {
    const store = usePipelineRunsStore();
    var vars = proxyVariablesToJsonObject(selectedItem.value.spec.variables);
    vars['cluster'] = cluster.value;
    vars['host'] = host.value;
    await store.create(selectedItem.value.metadata.namespace, selectedItem.value.metadata.name, vars);
    dialogVisble.value = false;
    loadData();
}

async function view(item) {
    try {
        const data = await pipelineStore.get(item.metadata.namespace, item.metadata.name);
        viewItem.value = data;
        viewDialogVisible.value = true;
    } catch (error) {
        console.error("Error loading pipeline:", error);
        alert("Failed to load pipeline: " + (error.message || error));
    }
}

async function edit(item) {
    try {
        const data = await pipelineStore.get(item.metadata.namespace, item.metadata.name);
        editItem.value = JSON.parse(JSON.stringify(data));
        editDialogVisible.value = true;
    } catch (error) {
        console.error("Error loading pipeline:", error);
        alert("Failed to load pipeline: " + (error.message || error));
    }
}

async function save() {
    try {
        const updatedPipeline = JSON.parse(editItemJson.value);
        await pipelineStore.update(updatedPipeline.metadata.namespace, updatedPipeline.metadata.name, updatedPipeline);
        closeEdit();
        loadData();
    } catch (error) {
        console.error("Error updating pipeline:", error);
        alert("Failed to update pipeline: " + (error.message || error));
    }
}

async function remove(item) {
    if (!confirm(`Are you sure you want to delete ${item.metadata.name}?`)) {
        return;
    }
    try {
        await pipelineStore.delete(item.metadata.namespace, item.metadata.name);
        loadData();
    } catch (error) {
        console.error("Error deleting pipeline:", error);
        alert("Failed to delete pipeline: " + (error.message || error));
    }
}

function create() {
    createItem.value = {
        apiVersion: "ops.shaowenchen.io/v1",
        kind: "Pipeline",
        metadata: {
            namespace: "ops-system",
            name: "",
        },
        spec: {
            desc: "",
            variables: {},
            tasks: [],
        },
    };
    createItemJson.value = JSON.stringify(createItem.value, null, 2);
    createDialogVisible.value = true;
}

function closeCreate() {
    createDialogVisible.value = false;
    createItem.value = null;
    createItemJson.value = "";
}

async function saveCreate() {
    try {
        const newPipeline = JSON.parse(createItemJson.value);
        if (!newPipeline.metadata.name) {
            alert("Name is required");
            return;
        }
        await pipelineStore.create(namespacesStore.selectedNamespace, newPipeline);
        closeCreate();
        loadData();
    } catch (error) {
        console.error("Error creating pipeline:", error);
        const errorMessage = error?.message || error?.toString() || String(error);
        alert("Failed to create pipeline: " + errorMessage);
    }
}
</script>

<template>
    <div class="container">
        <div class="namespace-filter">
            <label>Namespace:</label>
            <el-select :model-value="namespacesStore.selectedNamespace" class="namespace-select" @change="namespacesStore.setSelectedNamespace">
                <el-option
                    v-for="ns in namespaces"
                    :key="ns"
                    :label="ns"
                    :value="ns"
                />
            </el-select>
        </div>
        <div class="form-control enhanced-form">
            <el-row :gutter="20" align="middle">
                <el-col :span="18">
                    <el-input
                        v-model="searchQuery"
                        placeholder="Search..."
                        class="search-bar"
                        clearable
                        @input="loadData"
                    />
                </el-col>
                <el-col :span="6">
                    <el-button type="primary" @click="create" class="search-button">
                        Create
                    </el-button>
                </el-col>
            </el-row>
        </div>

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
                <div class="form-group">
                    <label>Cluster</label>
                    <el-select v-model="cluster" @change="onClusterChange">
                        <el-option v-for="item in clusters" :key="item.metadata.name" :label="item.metadata.name"
                            :value="item.metadata.name" />
                    </el-select>
                </div>
                <div class="form-group">
                    <label>Host</label>
                    <el-select v-model="host">
                        <el-option v-for="item in hosts" :key="item.metadata.name" :label="item.metadata.name"
                            :value="item.metadata?.name" />
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

        <el-dialog title="View Pipeline" v-model="viewDialogVisible" width="60%" :before-close="closeView">
            <div class="card-body" v-if="viewItem">
                <div class="form-group">
                    <label>Namespace</label>
                    <input type="text" disabled :value="viewItem.metadata?.namespace" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Name</label>
                    <input type="text" disabled :value="viewItem.metadata?.name" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input type="text" disabled :value="viewItem.spec?.desc" class="form-control" />
                </div>
                <div class="form-group" v-if="viewItem.spec?.variables">
                    <label>Variables</label>
                    <pre>{{ JSON.stringify(viewItem.spec.variables, null, 2) }}</pre>
                </div>
                <div class="form-group" v-if="viewItem.spec?.tasks">
                    <label>Tasks</label>
                    <pre>{{ JSON.stringify(viewItem.spec.tasks, null, 2) }}</pre>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="closeView">Close</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog title="Create Pipeline" v-model="createDialogVisible" width="60%" :before-close="closeCreate">
            <div class="card-body" v-if="createItem">
                <div class="form-group">
                    <label>Namespace</label>
                    <input type="text" v-model="createItem.metadata.namespace" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Name *</label>
                    <input type="text" v-model="createItem.metadata.name" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Pipeline JSON</label>
                    <textarea v-model="createItemJson" class="form-control" rows="20"></textarea>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="closeCreate">Cancel</el-button>
                    <el-button type="primary" @click="saveCreate">Create</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog title="Edit Pipeline" v-model="editDialogVisible" width="60%" :before-close="closeEdit">
            <div class="card-body" v-if="editItem">
                <div class="form-group">
                    <label>Namespace</label>
                    <input type="text" disabled v-model="editItem.metadata.namespace" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Name</label>
                    <input type="text" disabled v-model="editItem.metadata.name" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input type="text" v-model="editItem.spec.desc" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Pipeline JSON</label>
                    <textarea v-model="editItemJson" class="form-control" rows="20"></textarea>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="closeEdit">Cancel</el-button>
                    <el-button type="primary" @click="save">Save</el-button>
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
            <el-table-column label="Actions" width="400" class-name="actions-column">
                <template #default="scope">
                    <div class="actions-container">
                        <el-button type="primary" size="small" @click="view(scope.row)">View</el-button>
                        <el-button type="warning" size="small" @click="edit(scope.row)">Edit</el-button>
                        <el-button type="danger" size="small" @click="remove(scope.row)">Delete</el-button>
                        <el-button type="success" size="small" @click="run(scope.row)">Run</el-button>
                    </div>
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
    margin-bottom: 20px;
    width: 100%;
}

.enhanced-form {
    margin-bottom: 20px;
}

.search-bar {
    width: 100%;
}

.search-button {
    width: 100%;
}

.column-select {
    margin-bottom: 1em;
}

.actions-container {
    display: flex;
    gap: 5px;
    flex-wrap: nowrap;
    white-space: nowrap;
}

.actions-container .el-button {
    margin: 0;
}

.container {
    padding: 20px;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
}

.namespace-filter {
    margin-bottom: 20px;
    display: flex;
    align-items: center;
    gap: 10px;
}

.namespace-filter label {
    font-weight: bold;
    min-width: 80px;
}

.namespace-select {
    width: 200px;
}

.form-control {
    padding: 8px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
}

.form-group {
    margin-bottom: 20px;
}

.form-group label {
    display: block;
    margin-bottom: 5px;
    font-weight: bold;
}

.card-body {
    padding: 20px;
}

.dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
}
</style>

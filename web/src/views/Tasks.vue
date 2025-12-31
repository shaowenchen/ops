<script setup>
import { ref, watch, onMounted } from "vue";
import { useHostsStore, useTaskRunsStore, useTasksStore, useNamespacesStore } from "@/stores";
import { proxyVariablesToJsonObject, formatObject } from '@/utils/common';

const taskStore = useTasksStore();
const taskrunStore = useTaskRunsStore();
const hostStore = useHostsStore();
const namespacesStore = useNamespacesStore();

var dataList = ref([]);
var hosts = ref([]);
var host = ref(null);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
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
var namespaces = ref([]);

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

const allFields = [
    { value: 'metadata.namespace', label: 'Namespace' },
    { value: 'metadata.name', label: 'Name' },
    { value: 'spec.desc', label: 'Desc' },
    { value: 'spec.host', label: 'Host' },
    { value: 'spec.variables', label: 'Variables' },
];
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.desc', 'spec.host', 'spec.variables']);

async function loadNamespaces() {
    try {
        await namespacesStore.list();
        namespaces.value = namespacesStore.namespaces;
    } catch (error) {
        console.error("Error loading namespaces:", error);
        namespaces.value = ['ops-system'];
    }
}

async function getHosts() {
    var resp = await hostStore.list(namespacesStore.selectedNamespace, 999, 1);
    hosts = resp.list;
}

async function loadData() {
    var res = await taskStore.list(namespacesStore.selectedNamespace, pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list;
    total.value = res.total;
}

function onNamespaceChange() {
    currentPage.value = 1;
    loadData();
    getHosts();
}

onMounted(() => {
    loadNamespaces();
    loadData();
    getHosts();
});

watch(() => namespacesStore.selectedNamespace, () => {
    onNamespaceChange();
});

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
    loadData();
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

async function open() {
    host.value = null;
}

function run(item) {
    selectedItem.value = item;
    dialogVisble.value = true;
}

async function view(item) {
    try {
        const data = await taskStore.get(item.metadata.namespace, item.metadata.name);
        viewItem.value = data;
        viewDialogVisible.value = true;
    } catch (error) {
        console.error("Error loading task:", error);
        alert("Failed to load task: " + (error.message || error));
    }
}

async function edit(item) {
    try {
        const data = await taskStore.get(item.metadata.namespace, item.metadata.name);
        editItem.value = JSON.parse(JSON.stringify(data));
        editDialogVisible.value = true;
    } catch (error) {
        console.error("Error loading task:", error);
        alert("Failed to load task: " + (error.message || error));
    }
}

async function save() {
    try {
        const updatedTask = JSON.parse(editItemJson.value);
        await taskStore.update(updatedTask.metadata.namespace, updatedTask.metadata.name, updatedTask);
        closeEdit();
        loadData();
    } catch (error) {
        console.error("Error updating task:", error);
        alert("Failed to update task: " + (error.message || error));
    }
}

async function remove(item) {
    if (!confirm(`Are you sure you want to delete ${item.metadata.name}?`)) {
        return;
    }
    try {
        await taskStore.delete(item.metadata.namespace, item.metadata.name);
        loadData();
    } catch (error) {
        console.error("Error deleting task:", error);
        alert("Failed to delete task: " + (error.message || error));
    }
}

function create() {
    createItem.value = {
        apiVersion: "ops.shaowenchen.io/v1",
        kind: "Task",
        metadata: {
            namespace: "ops-system",
            name: "",
        },
        spec: {
            desc: "",
            host: "",
            variables: {},
            steps: [],
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
        const newTask = JSON.parse(createItemJson.value);
        if (!newTask.metadata.name) {
            alert("Name is required");
            return;
        }
        await taskStore.create(namespacesStore.selectedNamespace, newTask);
        closeCreate();
        loadData();
    } catch (error) {
        console.error("Error creating task:", error);
        const errorMessage = error?.message || error?.toString() || String(error);
        alert("Failed to create task: " + errorMessage);
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
                <div class="form-group" v-if="selectedItem?.spec?.host != 'anymaster'">
                    <label>Host</label>
                    <el-select v-model="host">
                        <el-option v-for="item in hosts" :key="item.metadata.name" :label="item.metadata.name"
                            :value="item.metadata.name" />
                    </el-select>
                </div>
                <div class="form-group" v-if="selectedItem?.spec?.variables">
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

        <el-dialog title="View Task" v-model="viewDialogVisible" width="60%" :before-close="closeView">
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
                <div class="form-group">
                    <label>Host</label>
                    <input type="text" disabled :value="viewItem.spec?.host" class="form-control" />
                </div>
                <div class="form-group" v-if="viewItem.spec?.variables">
                    <label>Variables</label>
                    <pre>{{ JSON.stringify(viewItem.spec.variables, null, 2) }}</pre>
                </div>
                <div class="form-group" v-if="viewItem.spec?.steps">
                    <label>Steps</label>
                    <pre>{{ JSON.stringify(viewItem.spec.steps, null, 2) }}</pre>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="closeView">Close</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog title="Create Task" v-model="createDialogVisible" width="60%" :before-close="closeCreate">
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
                    <label>Task JSON</label>
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

        <el-dialog title="Edit Task" v-model="editDialogVisible" width="60%" :before-close="closeEdit">
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
                    <label>Host</label>
                    <input type="text" v-model="editItem.spec.host" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Task JSON</label>
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
    height: 100%;
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

.actions-container {
    display: flex;
    gap: 5px;
    flex-wrap: nowrap;
    white-space: nowrap;
}

.actions-container .el-button {
    margin: 0;
}
</style>

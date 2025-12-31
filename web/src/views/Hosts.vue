<script setup>
import { ref, watch, onMounted } from 'vue';
import { useHostsStore, useNamespacesStore } from '@/stores';
import { formatObject } from '@/utils/common';

const hostStore = useHostsStore();
const namespacesStore = useNamespacesStore();

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref('');
var namespaces = ref([]);
var viewDialogVisible = ref(false);
var editDialogVisible = ref(false);
var createDialogVisible = ref(false);
var viewItem = ref(null);
var editItem = ref(null);
var createItem = ref(null);
var editItemJson = ref("");
var createItemJson = ref("");

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
    { value: 'status.hostname', label: 'Hostname' },
    { value: 'spec.address', label: 'Address' },
    { value: 'status.distribution', label: 'Distribution' },
    { value: 'status.arch', label: 'Arch' },
    { value: 'status.cpuTotal', label: 'CPU' },
    { value: 'status.memTotal', label: 'Mem' },
    { value: 'status.diskTotal', label: 'Disk' },
    { value: 'status.acceleratorVendor', label: 'Vendor' },
    { value: 'status.acceleratorModel', label: 'Model' },
    { value: 'status.acceleratorCount', label: 'Count' },
    { value: 'status.heartTime', label: 'HeartTime' },
    { value: 'status.heartStatus', label: 'HeartStatus' }
];
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'status.hostname', 'status.acceleratorVendor', 'status.heartStatus', 'status.heartTime']);

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
    var res = await hostStore.list(namespacesStore.selectedNamespace, pageSize.value, currentPage.value, searchQuery.value);
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

function onPaginationChange() {
    loadData();
}

function onPageSizeChange() {
    loadData();
}

async function view(item) {
    try {
        const data = await hostStore.get(item.metadata.namespace, item.metadata.name);
        viewItem.value = data;
        viewDialogVisible.value = true;
    } catch (error) {
        console.error("Error loading host:", error);
        alert("Failed to load host: " + (error.message || error));
    }
}

function closeView() {
    viewDialogVisible.value = false;
    viewItem.value = null;
}

async function edit(item) {
    try {
        const data = await hostStore.get(item.metadata.namespace, item.metadata.name);
        editItem.value = JSON.parse(JSON.stringify(data));
        editDialogVisible.value = true;
    } catch (error) {
        console.error("Error loading host:", error);
        alert("Failed to load host: " + (error.message || error));
    }
}

function closeEdit() {
    editDialogVisible.value = false;
    editItem.value = null;
}

function create() {
    createItem.value = JSON.parse(JSON.stringify({
        metadata: {
            namespace: "ops-system",
            name: "",
        },
        spec: {
            desc: "",
            address: "",
            port: 22,
            username: "",
            password: "",
            privateKey: "",
            privateKeyPath: "",
            timeoutSeconds: 30,
            secretRef: "",
        },
    }));
    createDialogVisible.value = true;
}

function closeCreate() {
    createDialogVisible.value = false;
    createItem.value = null;
}

async function save() {
    try {
        const updatedHost = JSON.parse(editItemJson.value);
        await hostStore.update(updatedHost.metadata.namespace, updatedHost.metadata.name, updatedHost);
        closeEdit();
        loadData();
    } catch (error) {
        console.error("Error updating host:", error);
        alert("Failed to update host: " + (error.message || error));
    }
}

async function saveCreate() {
    try {
        const newHost = JSON.parse(createItemJson.value);
        if (!newHost.metadata.name) {
            alert("Name is required");
            return;
        }
        await hostStore.create(namespacesStore.selectedNamespace, newHost);
        closeCreate();
        loadData();
    } catch (error) {
        console.error("Error creating host:", error);
        alert("Failed to create host: " + (error.message || error));
    }
}

async function remove(item) {
    if (!confirm(`Are you sure you want to delete ${item.metadata.name}?`)) {
        return;
    }
    try {
        await hostStore.delete(item.metadata.namespace, item.metadata.name);
        loadData();
    } catch (error) {
        console.error("Error deleting host:", error);
        alert("Failed to delete host: " + (error.message || error));
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

        <el-dialog title="View Host" v-model="viewDialogVisible" width="60%" :before-close="closeView">
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
                    <label>Hostname</label>
                    <input type="text" disabled :value="viewItem.status?.hostname" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Address</label>
                    <input type="text" disabled :value="viewItem.spec?.address" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Full Details</label>
                    <pre>{{ JSON.stringify(viewItem, null, 2) }}</pre>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="closeView">Close</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog title="Edit Host" v-model="editDialogVisible" width="60%" :before-close="closeEdit">
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
                    <label>Host JSON</label>
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

        <el-dialog title="Create Host" v-model="createDialogVisible" width="60%" :before-close="closeCreate">
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
                    <label>Host JSON</label>
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

        <el-table :data="dataList" border size="default">
            <el-table-column v-for="field in selectedFields" :key="field" :prop="field"
                :label="field.split('.').pop().charAt(0).toUpperCase() + field.split('.').pop().slice(1)">
                <template #default="{ row }">
                    <span v-html="formatObject(row, field)"></span>
                </template>
            </el-table-column>
            <el-table-column label="Actions" width="300" class-name="actions-column">
                <template #default="scope">
                    <div class="actions-container">
                        <el-button type="primary" size="small" @click="view(scope.row)">View</el-button>
                        <el-button type="warning" size="small" @click="edit(scope.row)">Edit</el-button>
                        <el-button type="danger" size="small" @click="remove(scope.row)">Delete</el-button>
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

.form-control {
    height: 100%;
    width: 100%;
    padding: 8px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
}

pre {
    background-color: #f5f5f5;
    padding: 10px;
    border-radius: 4px;
    overflow-x: auto;
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

<script setup>
import { ref, watch, onMounted } from 'vue';
import { usePipelineRunsStore, usePipelinesStore, useNamespacesStore } from '@/stores';
import { formatObject } from '@/utils/common';

const pipelineRunsStore = usePipelineRunsStore();
const pipelinesStore = usePipelinesStore();
const namespacesStore = useNamespacesStore();

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var searchQuery = ref("");
var namespaces = ref([]);
var dialogVisible = ref(false);
var createDialogVisible = ref(false);
var selectedItem = ref(null);
var createItem = ref(null);
var createItemJson = ref("");
var pipelines = ref([]);
var selectedFields = ref(['metadata.namespace', 'metadata.name', 'spec.pipelineRef', 'spec.variables', 'status.runStatus', 'status.startTime']);

watch(createItem, (newVal) => {
    if (newVal) {
        createItemJson.value = JSON.stringify(newVal, null, 2);
    }
}, { deep: true });

async function loadNamespaces() {
    try {
        await namespacesStore.list();
        namespaces.value = namespacesStore.namespaces;
    } catch (error) {
        console.error("Error loading namespaces:", error);
        namespaces.value = ['ops-system'];
    }
}

async function loadPipelines() {
    try {
        var res = await pipelinesStore.list(namespacesStore.selectedNamespace, 999, 1, "");
        pipelines.value = res.list || [];
    } catch (error) {
        console.error("Error loading pipelines:", error);
        pipelines.value = [];
    }
}

async function loadData() {
    var res = await pipelineRunsStore.list(namespacesStore.selectedNamespace, pageSize.value, currentPage.value, searchQuery.value);
    dataList.value = res.list
    total.value = res.total
}

function onNamespaceChange() {
    currentPage.value = 1;
    loadData();
    loadPipelines();
}

onMounted(() => {
    loadNamespaces();
    loadData();
    loadPipelines();
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

function view(item) {
    dialogVisible.value = true;
    selectedItem.value = item;
}

async function remove(item) {
    if (!confirm(`Are you sure you want to delete ${item.metadata.name}?`)) {
        return;
    }
    try {
        await pipelineRunsStore.delete(item.metadata.namespace, item.metadata.name);
        loadData();
    } catch (error) {
        console.error("Error deleting pipelinerun:", error);
        alert("Failed to delete pipelinerun: " + (error.message || error));
    }
}

function create() {
    createItem.value = JSON.parse(JSON.stringify({
        apiVersion: "ops.shaowenchen.io/v1",
        kind: "PipelineRun",
        metadata: {
            namespace: "ops-system",
            name: "",
        },
        spec: {
            desc: "",
            crontab: "",
            pipelineRef: "",
            variables: {},
        },
    }));
    createDialogVisible.value = true;
}

function closeCreate() {
    createDialogVisible.value = false;
    createItem.value = null;
    createItemJson.value = "";
}

async function saveCreate() {
    try {
        const newPipelineRun = JSON.parse(createItemJson.value);
        if (!newPipelineRun.spec.pipelineRef) {
            alert("PipelineRef is required");
            return;
        }
        await pipelineRunsStore.create(
            namespacesStore.selectedNamespace,
            newPipelineRun.spec.pipelineRef,
            newPipelineRun.spec.variables || {}
        );
        closeCreate();
        loadData();
    } catch (error) {
        console.error("Error creating pipelinerun:", error);
        const errorMessage = error?.message || error?.toString() || String(error);
        alert("Failed to create pipelinerun: " + errorMessage);
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

        <el-dialog v-model="dialogVisible" title="PipelineRun Details">
            <div class="card-body">
                <div class="form-group">
                    <pre>{{ JSON.stringify(selectedItem, null, 2) }}</pre>
                </div>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dialogVisible = false">Close</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog title="Create PipelineRun" v-model="createDialogVisible" width="60%" :before-close="closeCreate">
            <div class="card-body" v-if="createItem">
                <div class="form-group">
                    <label>Namespace</label>
                    <input type="text" v-model="createItem.metadata.namespace" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Name (optional, auto-generated if empty)</label>
                    <input type="text" v-model="createItem.metadata.name" class="form-control" />
                </div>
                <div class="form-group">
                    <label>PipelineRef *</label>
                    <el-select v-model="createItem.spec.pipelineRef" class="form-control" filterable>
                        <el-option
                            v-for="pipeline in pipelines"
                            :key="pipeline.metadata.name"
                            :label="`${pipeline.metadata.namespace}/${pipeline.metadata.name}`"
                            :value="pipeline.metadata.name"
                        />
                    </el-select>
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input type="text" v-model="createItem.spec.desc" class="form-control" />
                </div>
                <div class="form-group">
                    <label>Crontab</label>
                    <input type="text" v-model="createItem.spec.crontab" class="form-control" placeholder="e.g., 0 0 * * *" />
                </div>
                <div class="form-group">
                    <label>PipelineRun JSON</label>
                    <textarea v-model="createItemJson" class="form-control" rows="15"></textarea>
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
            <el-table-column label="Actions" width="200" class-name="actions-column">
                <template #default="scope">
                    <div class="actions-container">
                        <el-button type="primary" size="small" @click="view(scope.row)">View</el-button>
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
    margin-bottom: 20px;
    width: 100%;
    padding: 8px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
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

.formatted-variables {
    white-space: normal;
}

.kv-container {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 4px;
}

.kv-pair {
    display: block;
    padding: 4px 8px;
    background-color: #e3f2fd;
    color: #1565c0;
    border-radius: 4px;
    font-family: monospace;
    font-size: 13px;
    border: 1px solid #1565c0;
    width: fit-content;
}

:deep(.el-table .cell) {
    white-space: normal;
}
</style>
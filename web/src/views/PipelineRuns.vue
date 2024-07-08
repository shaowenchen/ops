<script setup>
import { ref } from 'vue';
import { usePipelineRunsStore } from '@/stores';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total    = ref(0);
async function loadData() {
    const store = usePipelineRunsStore();
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

var dialogVisible = ref(false);
var selectedItem = ref(null);
function view(item) {
    dialogVisible.value = true;
    selectedItem.value = item;
}

</script>

<template>
    <div class="container">
        <el-dialog v-model="dialogVisible" title="TaskRun Details">
            <div class="card-body">
                <div class="form-group">
                    <pre>{{ selectedItem?.status }}</pre>
                </div>
            </div>
        </el-dialog>
        <el-table :data="dataList" border size="default">
            <el-table-column prop="metadata.namespace" label="Namespace" />
            <el-table-column prop="metadata.name" label="Name"/>
            <el-table-column prop="spec.pipelineRef" label="PipelineRef" />
            <el-table-column prop="spec.nameRef" label="NameRef" />
            <el-table-column prop="status.runStatus" label="Run Status" />
            <el-table-column prop="status.startTime" label="Start Time" />
            <el-table-column label="Actions">
            <template #default="scope">
                <el-button type="primary" @click="view(scope.row)">View</el-button>  
            </template>
            </el-table-column>
        </el-table>
        <el-pagination
        @current-change="onPaginationChange"
        @size-change="onPageSizeChange"
        v-model:currentPage="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 30]"
        layout="total, sizes, prev, pager, next"
        :total="total">
        </el-pagination>
    </div>
</template>

<style scoped>
.contaner {
    margin-left: 7em;
}
.form-control {
    display: inline-block;
    margin-right: 10px;
}
</style>

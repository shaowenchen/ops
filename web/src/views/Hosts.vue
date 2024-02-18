<script setup>
import { ref } from 'vue';
import { useHostsStore } from '@/stores';

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
async function loadData() {
    const store = useHostsStore();
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
</script>

<template>
    <div class="container">
        <el-table :data="dataList" border size="default">
            <el-table-column prop="metadata.namespace" label="Namespace" />
            <el-table-column prop="metadata.name" label="Name" />
            <el-table-column prop="status.hostname" label="Hostname" />
            <el-table-column prop="spec.address" label="Address" />
            <el-table-column prop="status.distribution" label="Distribution" />
            <el-table-column prop="status.arch" label="Arch" />
            <el-table-column prop="status.cputotal" label="CPU" />
            <el-table-column prop="status.memtotal" label="Mem" />
            <el-table-column prop="status.disktotal" label="Disk" />
            <el-table-column prop="status.heartTime" label="HeartTime" />
            <el-table-column prop="status.heartStatus" label="HeartStatus" />
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

.form-control {
    display: inline-block;
    width: 80%;
    margin-right: 10px;
}
</style>

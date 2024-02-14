<script setup>
import { ref } from 'vue';
import { useTaskRunsStore } from '@/stores';

var dataList = ref([]);
async function fresh() {
    const store = useTaskRunsStore();
    dataList.value = await store.list("all");
}
fresh();

var dialogVisible = ref(false);
var selectedItem = ref(null);
function view(item) {
    dialogVisible.value = true;
    selectedItem.value = item.item;
}

</script>

<template>
    <el-dialog v-model="dialogVisible" title="TaskRun Details">
        <div class="card-body">
            <div class="form-group">
                <pre>{{ selectedItem?.status }}</pre>
            </div>
        </div>
    </el-dialog>
    <div class="card m-3">
        <table class="table table-bordered">
            <thead>
                <tr>
                    <th>Namespace</th>
                    <th>Name</th>
                    <th>TaskRef</th>
                    <th>NameRef</th>
                    <th>NodeName</th>
                    <th>All</th>
                    <th>Create Time</th>
                    <th>Run Status</th>
                    <th>Log</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="item in dataList">
                    <td>{{ item.metadata.namespace }}</td>
                    <td>{{ item.metadata.name }}</td>
                    <td>{{ item.spec.taskRef }}</td>
                    <td>{{ item.spec.nameRef }}</td>
                    <td>{{ item.spec.nodeName }}</td>
                    <td>{{ item.spec.all }}</td>
                    <td>{{ item.metadata.creationTimestamp }}</td>
                    <td>{{ item.status.runStatus }}</td>
                    <td>
                        <el-button type="primary" @click="view({ item })">View</el-button>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<style scoped>
.form-control {
    display: inline-block;
    width: 80%;
    margin-right: 10px;
}
</style>

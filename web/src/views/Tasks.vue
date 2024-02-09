<script setup>
import { ref } from 'vue';
import { useTasksStore, useTaskRunsStore } from '@/stores';

var dataList = ref([]);
async function fresh() {
    const store = useTasksStore();
    dataList.value = await store.list("all");
}
fresh();

var dialogVisble = ref(false)
var selectedItem = ref(null)
async function confirm() {
    const store = useTaskRunsStore();
    dataList.value = await store.create();
    dialogVisble.value = false;
}

function close() {
    dialogVisble.value = false;
}

function run(item) {
    selectedItem.value = item.item;
    dialogVisble.value = true;
}
</script>

<template>
    <el-dialog title="Create TaskRun" v-model="dialogVisble" width="30%" :before-close="close">
        <div class="card-body" v-if="selectedItem">
            <div class="form-group">
                <label>Namespace</label>
                <input name="namespace" type="text" :value="selectedItem?.value?.metadata?.namespace"
                    class="form-control" />
            </div>
            <div class="form-group">
                <label>Name</label>
                <input name="name" type="text" :value="selectedItem?.value?.metadata?.name" class="form-control" />
            </div>
            <div class="form-group">
                <label>Description</label>
                <input name="desc" type="text" :value="selectedItem?.value?.spec?.desc" class="form-control" />
            </div>
            <div class="form-group">
                <label>Steps</label>
                <input name="steps" type="text" :value="selectedItem?.value?.spec?.steps" class="form-control" />
            </div>
            <div class="form-group">
                <label>TypeRef</label>
                <input name="typeRef" type="text" :value="selectedItem?.value?.spec?.typeRef" class="form-control" />
            </div>
            <div class="form-group">
                <label>NameRef</label>
                <input name="nameRef" type="text" :value="selectedItem?.value?.spec?.nameRef" class="form-control" />
            </div>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="close">Cancel</el-button>
                <el-button type="primary" @click="confirm">Run</el-button>
            </span>
        </template>
    </el-dialog>
    <div class="card m-3">
        <table class="table table-bordered">
            <thead>
                <tr>
                    <th>Namespace</th>
                    <th>Name</th>
                    <th>Crontab</th>
                    <th>TypeRef</th>
                    <th>NameRef</th>
                    <th>Start Time</th>
                    <th>Run Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="item in dataList">
                    <td>{{ item.metadata.namespace }}</td>
                    <td>{{ item.metadata.name }}</td>
                    <td>{{ item.spec.crontab }}</td>
                    <td>{{ item.spec.typeRef }}</td>
                    <td>{{ item.spec.nameRef }}</td>
                    <td>{{ item.status.startTime }}</td>
                    <td>{{ item.status.runStatus }}</td>
                    <td><el-button type="primary" @click="run({ item })">run</el-button></td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<style scoped>
.card-body {
    width: 500px;
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

<script setup>
import { ref } from 'vue';
import { useTaskRunsStore } from '@/stores';

var dataList = ref([]);
async function fresh() {
    const store = useTaskRunsStore();
    dataList.value = await store.list("all");
}
fresh();
</script>

<template>
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
                    <th>Start Time</th>
                    <th>Run Status</th>
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
                    <td>{{ item.status.startTime }}</td>
                    <td>{{ item.status.runStatus }}</td>
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

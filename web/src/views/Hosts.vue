<script setup>
import { ref } from 'vue';
import { useHostsStore } from '@/stores';

var dataList = ref([]);
async function fresh() {
    const store = useHostsStore();
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
                    <th>Hostname</th>
                    <th>Address</th>
                    <th>Distribution</th>
                    <th>Arch</th>
                    <th>CPU</th>
                    <th>Mem</th>
                    <th>Disk</th>
                    <th>HeartTime</th>
                    <th>HeartStatus</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="item in dataList">
                    <td>{{ item.metadata.namespace }}</td>
                    <td>{{ item.metadata.name }}</td>
                    <td>{{ item.status.hostname }}</td>
                    <td>{{ item.spec.address }}</td>
                    <td>{{ item.status.distribution }}</td>
                    <td>{{ item.status.arch }}</td>
                    <td>{{ item.status.cputotal }}</td>
                    <td>{{ item.status.memtotal }}</td>
                    <td>{{ item.status.disktotal }}</td>
                    <td>{{ item.status.heartTime }}</td>
                    <td>{{ item.status.heartStatus }}</td>
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

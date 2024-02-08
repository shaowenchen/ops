<script setup>
import { ref } from 'vue';
import { useClustersStore } from '@/stores';

var dataList = ref([]);
async function fresh() {
    const store = useClustersStore();
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
                    <th>Server</th>
                    <th>Version</th>
                    <th>Node</th>
                    <th>Running</th>
                    <th>TotalPod</th>
                    <th>CertDays</th>
                    <th>HeartTime</th>
                    <th>HeartStatus</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="item in dataList">
                    <td>{{ item.metadata.namespace }}</td>
                    <td>{{ item.metadata.name }}</td>
                    <td>{{ item.spec.server }}</td>
                    <td>{{ item.status.version }}</td>
                    <td>{{ item.status.node }}</td>
                    <td>{{ item.status.runningPod }}</td>
                    <td>{{ item.status.pod }}</td>
                    <td>{{ item.status.certNotAfterDays }}</td>
                    <td>{{ item.status.heartTime }}</td>
                    <td>{{ item.status.heartstatus }}</td>
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

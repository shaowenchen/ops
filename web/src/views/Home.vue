<script setup>
import { ref } from 'vue';
import { useLoginStore, useSummaryStore } from "@/stores";

const loginStore = useLoginStore();
loginStore.check();

const summaryStore = useSummaryStore();
const summaryData = ref({});

async function loadData() {
    summaryData.value = await summaryStore.get();
}

loadData();
</script>

<template>
    <div class="container" v-if="summaryData">
        <p>
        <div class="card">
            <router-link to="/copilot" class="card-link">
                <div class="card-title">Copilot</div>
            </router-link>
        </div>

        <div class="card">
            <router-link to="/clusters" class="card-link">
                <div class="card-title">Clusters</div>
                <div class="card-data">{{ summaryData.clusters }}</div>
            </router-link>
        </div>

        <div class="card">
            <router-link to="/hosts" class="card-link">
                <div class="card-title">Hosts</div>
                <div class="card-data">{{ summaryData.hosts }}</div>
            </router-link>
        </div>
        </p>

        <p>
        <div class="card">
            <router-link to="/pipelines" class="card-link">
                <div class="card-title">Pipelines</div>
                <div class="card-data">{{ summaryData.pipelines }}</div>
            </router-link>
        </div>

        <div class="card">
            <router-link to="/pipelineruns" class="card-link">
                <div class="card-title">Pipelineruns</div>
                <div class="card-data">{{ summaryData.pipelineruns }}</div>
            </router-link>
        </div>

        <div class="card">
            <router-link to="/tasks" class="card-link">
                <div class="card-title">Tasks</div>
                <div class="card-data">{{ summaryData.tasks }}</div>
            </router-link>
        </div>

        <div class="card">
            <router-link to="/taskruns" class="card-link">
                <div class="card-title">Taskruns</div>
                <div class="card-data">{{ summaryData.taskruns }}</div>
            </router-link>
        </div>
        </p>

    </div>
</template>

<style scoped>
.container {
    margin: 0 auto;
    padding: 2em;
    min-height: 100vh;
    display: flex;
    justify-content: center;
}

.card {
    width: 200px; 
    height: 150px; 
    margin: 10px; 
    border-radius: 4px;
    overflow: hidden;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1); 
    display: flex; 
    justify-content: center; 
    align-items: center; 
}

.card-title {
    padding: 10px;
    background-color: #f7f7f7;
    font-weight: bold;
    text-align: center; 
}

.card-data {
    padding: 10px;
    font-size: 1.2em;
    text-align: center; 
}
</style>

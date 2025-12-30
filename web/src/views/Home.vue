<script setup>
import { ref, onMounted } from "vue";
import { useLoginStore, useSummaryStore } from "@/stores";

var loginStore = useLoginStore();
loginStore.check();

const summaryStore = useSummaryStore();
var summaryData = ref({
  clusters: 0,
  hosts: 0,
  pipelines: 0,
  pipelineruns: 0,
  tasks: 0,
  taskruns: 0,
  eventhooks: 0
});

async function loadData() {
  try {
    const data = await summaryStore.get();
    summaryData.value = data;
  } catch (error) {
    console.error("Error loading summary data:", error);
    summaryData.value = {
      clusters: 0,
      hosts: 0,
      pipelines: 0,
      pipelineruns: 0,
      tasks: 0,
      taskruns: 0,
      eventhooks: 0
    };
  }
}

onMounted(() => {
  loadData();
});
</script>
<template>

    <el-row :gutter="12" class="container" v-if="summaryData">
        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/clusters" class="card-link">
                    <div class="card-title">Clusters</div>
                    <div class="card-data">{{ summaryData.clusters }}</div>
                </router-link>
            </el-card>
        </el-col>

        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/hosts" class="card-link">
                    <div class="card-title">Hosts</div>
                    <div class="card-data">{{ summaryData.hosts }}</div>
                </router-link>
            </el-card>
        </el-col>

        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/pipelines" class="card-link">
                    <div class="card-title">Pipelines</div>
                    <div class="card-data">{{ summaryData.pipelines }}</div>
                </router-link>
            </el-card>
        </el-col>

        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/pipelineruns" class="card-link">
                    <div class="card-title">Pipelineruns</div>
                    <div class="card-data">{{ summaryData.pipelineruns }}</div>
                </router-link>
            </el-card>
        </el-col>

        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/tasks" class="card-link">
                    <div class="card-title">Tasks</div>
                    <div class="card-data">{{ summaryData.tasks }}</div>
                </router-link>
            </el-card>
        </el-col>

        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/taskruns" class="card-link">
                    <div class="card-title">Taskruns</div>
                    <div class="card-data">{{ summaryData.taskruns }}</div>
                </router-link>
            </el-card>
        </el-col>
        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/events" class="card-link">
                    <div class="card-title">Events</div>
                </router-link>
            </el-card>
        </el-col>

        <el-col :span="8">
            <el-card shadow="hover" class="card">
                <router-link to="/eventhooks" class="card-link">
                    <div class="card-title">EventHooks</div>
                    <div class="card-data">{{ summaryData.eventhooks }}</div>
                </router-link>
            </el-card>
        </el-col>
    </el-row>
</template>

<style scoped>
.container {
    margin: 0 auto;
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
}

.el-col {
    margin: 0;
}

.card {
    width: 100%;
    margin: 1em;
    border-radius: 4px;
    overflow: hidden;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    display: flex;
    justify-content: center;
    align-items: center;
    text-align: center;
}

.card-link {
    text-decoration: none;
    color: inherit;
    display: block;
    width: 100%;
}

.card-title {
    padding: 1em;
    background-color: #f7f7f7;
    font-weight: bold;
    font-size: 1.2em;
}

.card-data {
    padding: 1em;
    font-size: 1.5em;
    color: #409EFF;
}
</style>

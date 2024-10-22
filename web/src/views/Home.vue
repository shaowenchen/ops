<script setup>
import { useLoginStore } from "@/stores";

var loginStore = useLoginStore();
loginStore.check();
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
                <router-link to="/copilot" class="card-link">
                    <div class="card-title">Copilot</div>
                </router-link>
            </el-card>
        </el-col>

    </el-row>
</template>

<script>
export default {
    data() {
        return {
            summaryData: {
                clusters: 10,
                hosts: 20,
                pipelines: 5,
                pipelineruns: 8,
                tasks: 12,
                taskruns: 15
            }
        };
    }
};
</script>

<style scoped>
.container {
    margin: 0 auto;
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
}

e-col {
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

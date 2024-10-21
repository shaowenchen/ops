<script setup>
import { useLoginStore } from "@/stores";

var loginStore = useLoginStore();
loginStore.check();
</script>
<template>
    <div class="contaner">
        <div class="content">
            The home page is under development,, you can see <router-link to="/tasks">Tasks</router-link>
        </div>
    </div>
</template>

<style scoped>
.contaner {
    margin-left: 7em;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
}

.form-control {
    display: inline-block;
    width: 80%;
    margin-right: 10px;
}

.content {
    text-align: center;
}
</style>

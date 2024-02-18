<script setup>
import { ref } from 'vue';
import { useLoginStore } from '@/stores';
import { router } from '@/router'

var token = ref([]);

function save() {
    const store = useLoginStore();
    store.save(token);
    router.push("/");
}

</script>

<template>
    <div class="container">
        <el-input class="token" v-model="token" placeholder="Please input token"></el-input>
        <el-button type="primary" @click="save">Save</el-button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    margin-left: 7em;
}

.container .token {
    width: 20em;
    margin-bottom: 1em;
}
</style>

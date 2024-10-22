<script setup>
import { ref } from 'vue';
import { useLoginStore } from '@/stores';
import { useAlertStore } from '@/stores/alert.store';
import { router } from '@/router'

var token = ref([]);
var isDisabled = ref(false);

async function save() {
    isDisabled.value = true;
    const store = useLoginStore();
    store.save(token.value);
    if (store.check()) {
        useAlertStore().success("Login success");
        setTimeout(() => {
            useAlertStore().clear();
            router.push("/");
        }, 1500);

    } else {
        store.clear();
    }
    isDisabled.value = false;
}

</script>

<template>
    <div class="container" :hidden="isDisabled">
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
}

.container .token {
    width: 20em;
    margin-bottom: 1em;
}
</style>

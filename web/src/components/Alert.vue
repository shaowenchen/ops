<script setup>
import { storeToRefs } from 'pinia';
import { useAlertStore } from '@/stores';
import { ref, watch } from 'vue';

const alertStore = useAlertStore();
const { alert } = storeToRefs(alertStore);

const isVisible = ref(false);

watch(alert, (newValue) => {
    if (newValue) {
        isVisible.value = true;

        setTimeout(() => {
            alertStore.clear();
            isVisible.value = false;
        }, 4000);
    }
});
</script>

<template>
    <div v-if="isVisible" class="alert-container">
        <div class="m-3">
            <div class="alert alert-dismissable" :class="alert.type">
                <button @click="alertStore.clear()" class="btn btn-link close">&times;</button>
                {{ alert.message }}
            </div>
        </div>
    </div>
</template>

<style scoped>
.alert-container {
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 1000;
    max-width: 300px;
    width: 100%;
    text-align: center;
}
</style>

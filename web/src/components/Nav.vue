<template>
    <div>
        <el-col :span="4">
            <el-menu :default-active="$route.path" router mode="vertical" class="custom-menu" background-color="#333"
                text-color="#fff" active-text-color="#ffd04b" :default-openeds="['/']" @open="handleOpen"
                @close="handleClose">

                <el-menu-item v-for="(item, i) in navList" :key="i" :index="item.path">
                    <router-link :to="item.path" class="nav-link" exact-active-class="active">
                        <i :class="item.icon"></i>
                        <span>{{ item.navItem }}</span>
                    </router-link>
                </el-menu-item>
                <el-menu-item>
                    <a href="https://github.com/shaowenchen/ops/issues/new" target="_blank" class="nav-link">
                        <i class="el-icon-question"></i>
                        <span>Help</span>
                    </a>
                </el-menu-item>
            </el-menu>
        </el-col>
    </div>
</template>


<script>
export default {
    data() {
        return {
            navList: [
                { navItem: 'Home', path: '/', icon: 'el-icon-house' },
                { navItem: 'Hosts', path: '/hosts', icon: 'el-icon-s-platform' },
                { navItem: 'Clusters', path: '/clusters', icon: 'el-icon-s-help' },
                { navItem: 'Tasks', path: '/tasks', icon: 'el-icon-s-check' },
                { navItem: 'TaskRuns', path: '/taskruns', icon: 'el-icon-s-operation' },
                { navItem: 'Pipelines', path: '/pipelines', icon: 'el-icon-s-promotion' },
                { navItem: 'PipelineRuns', path: '/pipelineruns', icon: 'el-icon-s-tools' },
                { navItem: 'Events', path: '/events', icon: 'el-icon-s-release' },
                { navItem: 'Logout', path: '/logout', icon: 'el-icon-s-release' }
            ]
        };
    },
    methods: {
        handleOpen() {
            // Handle open action
        },
        handleClose() {
            // Handle close action
        }
    }
};
</script>

<style scoped>
.custom-menu {
    min-height: 100vh;
    min-width: 10em;
    border-right: 1px solid #ddd;
    background-color: #333;
}

.nav-link {
    color: #fff;
    display: flex;
    align-items: center;
    padding: 0;
}

.nav-link:hover {
    color: #ffd04b;
}

.el-menu-item:hover {
    background-color: #444;
}

/* 添加激活状态下的样式 */
.active {
    color: #ffd04b;
    font-weight: bold;
}
</style>

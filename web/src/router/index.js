import { createRouter, createWebHistory } from "vue-router";

import {
  ClusterDetails,
  Clusters,
  Events,
  Home,
  Hosts,
  Login,
  Logout,
  PipelineRuns,
  Pipelines,
  TaskRuns,
  Tasks
} from "@/views";

export const router = createRouter({
  history: createWebHistory(),
  linkActiveClass: "active",
  routes: [
    { path: "/", component: Home, name: "home" },
    { path: "/hosts", component: Hosts, name: "hosts" },
    { path: "/clusters", component: Clusters, name: "clusters" },
    { path: "/clusters/:cluster/details", component: ClusterDetails, name: "cluster-details" },
    { path: "/tasks", component: Tasks, name: "tasks" },
    { path: "/taskruns", component: TaskRuns, name: "taskruns" },
    { path: "/pipelines", component: Pipelines, name: "pipelines" },
    { path: "/pipelineruns", component: PipelineRuns, name: "pipelineruns" },
    { path: "/login", component: Login, name: "login" },
    { path: "/logout", component: Logout, name: "logout" },
    { path: "/events", component: Events, name: "events" },
  ],
});

router.beforeEach(async (to) => {});

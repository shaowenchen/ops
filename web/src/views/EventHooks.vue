<script setup>
import { ref } from "vue";
import { useEventHooksStore, useLoginStore } from "@/stores";
import { formatObject } from "@/utils/common";

var loginStore = useLoginStore();
loginStore.check();

const eventHooksStore = useEventHooksStore();

var dataList = ref([]);
var currentPage = ref(1);
var pageSize = ref(10);
var total = ref(0);
var dialogVisible = ref(false);
var dialogMode = ref("view"); // 'view', 'create', 'edit'
var selectedItem = ref(null);
var formData = ref({
  metadata: {
    namespace: "all",
    name: "",
  },
  spec: {
    type: "",
    subject: "",
    url: "",
    options: {},
    keywords: {
      include: [],
      exclude: [],
      matchMode: "ANY",
      matchType: "CONTAINS",
    },
    additional: "",
  },
});

const allFields = [
  { value: "metadata.namespace", label: "Namespace" },
  { value: "metadata.name", label: "Name" },
  { value: "spec.type", label: "Type" },
  { value: "spec.subject", label: "Subject" },
  { value: "spec.url", label: "URL" },
  { value: "spec.keywords", label: "Keywords" },
];

var selectedFields = ref([
  "metadata.namespace",
  "metadata.name",
  "spec.type",
  "spec.subject",
  "spec.url",
]);

loadData();

async function loadData() {
  var res = await eventHooksStore.list("all", pageSize.value, currentPage.value);
  dataList.value = res.list;
  total.value = res.total;
}

function onPaginationChange() {
  loadData();
}

function onPageSizeChange() {
  loadData();
}

function view(item) {
  selectedItem.value = item;
  dialogMode.value = "view";
  dialogVisible.value = true;
}

function create() {
  formData.value = JSON.parse(JSON.stringify({
    metadata: {
      namespace: "all",
      name: "",
    },
    spec: {
      type: "",
      subject: "",
      url: "",
      options: {},
      keywords: {
        include: [],
        exclude: [],
        matchMode: "ANY",
        matchType: "CONTAINS",
      },
      additional: "",
    },
  }));
  dialogMode.value = "create";
  dialogVisible.value = true;
}

function edit(item) {
  selectedItem.value = item;
  // Deep clone to avoid mutating the original object
  formData.value = JSON.parse(JSON.stringify({
    metadata: {
      namespace: item.metadata.namespace || "all",
      name: item.metadata.name || "",
    },
    spec: {
      type: item.spec.type || "",
      subject: item.spec.subject || "",
      url: item.spec.url || "",
      options: item.spec.options ? { ...item.spec.options } : {},
      keywords: item.spec.keywords ? {
        include: [...(item.spec.keywords.include || [])],
        exclude: [...(item.spec.keywords.exclude || [])],
        matchMode: item.spec.keywords.matchMode || "ANY",
        matchType: item.spec.keywords.matchType || "CONTAINS",
      } : {
        include: [],
        exclude: [],
        matchMode: "ANY",
        matchType: "CONTAINS",
      },
      additional: item.spec.additional || "",
    },
  }));
  dialogMode.value = "edit";
  dialogVisible.value = true;
}

function close() {
  dialogVisible.value = false;
  selectedItem.value = null;
}

async function save() {
  try {
    if (dialogMode.value === "create") {
      await eventHooksStore.create("all", formData.value);
    } else if (dialogMode.value === "edit") {
      await eventHooksStore.update(
        selectedItem.value.metadata.namespace,
        selectedItem.value.metadata.name,
        formData.value
      );
    }
    close();
    loadData();
  } catch (error) {
    console.error("Error saving eventhook:", error);
    alert("Failed to save eventhook: " + (error.message || error));
  }
}

async function remove(item) {
  if (!confirm(`Are you sure you want to delete ${item.metadata.name}?`)) {
    return;
  }
  try {
    await eventHooksStore.delete(item.metadata.namespace, item.metadata.name);
    loadData();
  } catch (error) {
    console.error("Error deleting eventhook:", error);
    alert("Failed to delete eventhook: " + (error.message || error));
  }
}

function addKeyword(type) {
  if (!formData.value.spec.keywords[type]) {
    formData.value.spec.keywords[type] = [];
  }
  formData.value.spec.keywords[type].push("");
}

function removeKeyword(type, index) {
  formData.value.spec.keywords[type].splice(index, 1);
}

function addOption() {
  if (!formData.value.spec.options) {
    formData.value.spec.options = {};
  }
  // Add a new empty option with a temporary key
  const tempKey = `new_option_${Date.now()}`;
  formData.value.spec.options[tempKey] = "";
}

function removeOption(key) {
  if (formData.value.spec.options) {
    delete formData.value.spec.options[key];
  }
}

function updateOptionKey(oldKey, newKey) {
  if (oldKey === newKey || !newKey) return;
  if (formData.value.spec.options && formData.value.spec.options[oldKey] !== undefined) {
    formData.value.spec.options[newKey] = formData.value.spec.options[oldKey];
    delete formData.value.spec.options[oldKey];
  }
}
</script>

<template>
  <div class="container">
    <div class="form-control">
      <el-button type="primary" @click="create">Create EventHook</el-button>
    </div>

    <el-select
      v-model="selectedFields"
      multiple
      placeholder="Select columns to display"
      class="column-select"
    >
      <el-option
        v-for="field in allFields"
        :key="field.value"
        :label="field.label"
        :value="field.value"
      />
    </el-select>

    <el-dialog
      :title="
        dialogMode === 'view'
          ? 'View EventHook'
          : dialogMode === 'create'
          ? 'Create EventHook'
          : 'Edit EventHook'
      "
      v-model="dialogVisible"
      width="60%"
      :before-close="close"
    >
      <div class="card-body" v-if="dialogMode === 'view'">
        <div class="form-group">
          <label>Namespace</label>
          <input
            type="text"
            disabled
            :value="selectedItem?.metadata?.namespace"
            class="form-control"
          />
        </div>
        <div class="form-group">
          <label>Name</label>
          <input
            type="text"
            disabled
            :value="selectedItem?.metadata?.name"
            class="form-control"
          />
        </div>
        <div class="form-group">
          <label>Type</label>
          <input
            type="text"
            disabled
            :value="selectedItem?.spec?.type"
            class="form-control"
          />
        </div>
        <div class="form-group">
          <label>Subject</label>
          <input
            type="text"
            disabled
            :value="selectedItem?.spec?.subject"
            class="form-control"
          />
        </div>
        <div class="form-group">
          <label>URL</label>
          <input
            type="text"
            disabled
            :value="selectedItem?.spec?.url"
            class="form-control"
          />
        </div>
        <div class="form-group" v-if="selectedItem?.spec?.keywords">
          <label>Keywords</label>
          <div v-if="selectedItem.spec.keywords.include?.length > 0">
            <strong>Include:</strong>
            <ul>
              <li v-for="(kw, idx) in selectedItem.spec.keywords.include" :key="idx">
                {{ kw }}
              </li>
            </ul>
          </div>
          <div v-if="selectedItem.spec.keywords.exclude?.length > 0">
            <strong>Exclude:</strong>
            <ul>
              <li v-for="(kw, idx) in selectedItem.spec.keywords.exclude" :key="idx">
                {{ kw }}
              </li>
            </ul>
          </div>
          <div v-if="selectedItem.spec.keywords.matchMode">
            <strong>Match Mode:</strong> {{ selectedItem.spec.keywords.matchMode }}
          </div>
          <div v-if="selectedItem.spec.keywords.matchType">
            <strong>Match Type:</strong> {{ selectedItem.spec.keywords.matchType }}
          </div>
        </div>
        <div class="form-group" v-if="selectedItem?.spec?.options">
          <label>Options</label>
          <pre>{{ JSON.stringify(selectedItem.spec.options, null, 2) }}</pre>
        </div>
        <div class="form-group" v-if="selectedItem?.spec?.additional">
          <label>Additional</label>
          <textarea
            disabled
            :value="selectedItem.spec.additional"
            class="form-control"
            rows="4"
          ></textarea>
        </div>
      </div>

      <div class="card-body" v-else>
        <div class="form-group">
          <label>Namespace</label>
          <input
            type="text"
            v-model="formData.metadata.namespace"
            class="form-control"
            :disabled="dialogMode === 'edit'"
          />
        </div>
        <div class="form-group">
          <label>Name *</label>
          <input
            type="text"
            v-model="formData.metadata.name"
            class="form-control"
            :disabled="dialogMode === 'edit'"
          />
        </div>
        <div class="form-group">
          <label>Type</label>
          <input type="text" v-model="formData.spec.type" class="form-control" />
        </div>
        <div class="form-group">
          <label>Subject</label>
          <input
            type="text"
            v-model="formData.spec.subject"
            class="form-control"
            placeholder="e.g., ops.>"
          />
        </div>
        <div class="form-group">
          <label>URL</label>
          <input type="text" v-model="formData.spec.url" class="form-control" />
        </div>
        <div class="form-group">
          <label>Keywords - Include</label>
          <div
            v-for="(kw, index) in formData.spec.keywords.include"
            :key="'include-' + index"
            class="form-item"
          >
            <input
              type="text"
              v-model="formData.spec.keywords.include[index]"
              class="form-control"
            />
            <el-button
              type="danger"
              size="small"
              @click="removeKeyword('include', index)"
            >
              Remove
            </el-button>
          </div>
          <el-button type="primary" size="small" @click="addKeyword('include')">
            Add Include Keyword
          </el-button>
        </div>
        <div class="form-group">
          <label>Keywords - Exclude</label>
          <div
            v-for="(kw, index) in formData.spec.keywords.exclude"
            :key="'exclude-' + index"
            class="form-item"
          >
            <input
              type="text"
              v-model="formData.spec.keywords.exclude[index]"
              class="form-control"
            />
            <el-button
              type="danger"
              size="small"
              @click="removeKeyword('exclude', index)"
            >
              Remove
            </el-button>
          </div>
          <el-button type="primary" size="small" @click="addKeyword('exclude')">
            Add Exclude Keyword
          </el-button>
        </div>
        <div class="form-group">
          <label>Match Mode</label>
          <el-select v-model="formData.spec.keywords.matchMode">
            <el-option label="ANY" value="ANY" />
            <el-option label="ALL" value="ALL" />
          </el-select>
        </div>
        <div class="form-group">
          <label>Match Type</label>
          <el-select v-model="formData.spec.keywords.matchType">
            <el-option label="CONTAINS" value="CONTAINS" />
            <el-option label="EXACT" value="EXACT" />
            <el-option label="REGEX" value="REGEX" />
          </el-select>
        </div>
        <div class="form-group">
          <label>Options</label>
          <div
            v-for="(value, key) in formData.spec.options"
            :key="key"
            class="form-item"
          >
            <input
              type="text"
              :value="key"
              @blur="(e) => updateOptionKey(key, e.target.value)"
              class="form-control"
              placeholder="Key"
            />
            <input
              type="text"
              v-model="formData.spec.options[key]"
              class="form-control"
              placeholder="Value"
            />
            <el-button
              type="danger"
              size="small"
              @click="removeOption(key)"
            >
              Remove
            </el-button>
          </div>
          <el-button type="primary" size="small" @click="addOption">
            Add Option
          </el-button>
        </div>
        <div class="form-group">
          <label>Additional</label>
          <textarea
            v-model="formData.spec.additional"
            class="form-control"
            rows="4"
          ></textarea>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="close">Cancel</el-button>
          <el-button
            v-if="dialogMode !== 'view'"
            type="primary"
            @click="save"
          >
            Save
          </el-button>
        </span>
      </template>
    </el-dialog>

    <el-table :data="dataList" border size="default">
      <el-table-column
        v-for="field in selectedFields"
        :key="field"
        :prop="field"
        :label="
          field
            .split('.')
            .pop()
            .charAt(0)
            .toUpperCase() + field.split('.').pop().slice(1)
        "
      >
        <template #default="{ row }">
          <span v-html="formatObject(row, field)"></span>
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="250">
        <template #default="scope">
          <el-button type="primary" @click="view(scope.row)">View</el-button>
          <el-button type="warning" @click="edit(scope.row)">Edit</el-button>
          <el-button type="danger" @click="remove(scope.row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      @current-change="onPaginationChange"
      @size-change="onPageSizeChange"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="[10, 20, 30]"
      layout="total, sizes, prev, pager, next"
      :total="total"
    >
    </el-pagination>
  </div>
</template>

<style scoped>
.container {
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.form-control {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  width: 100%;
}

.column-select {
  margin-bottom: 1em;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-control {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.form-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.form-item input {
  flex: 1;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

pre {
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
}

ul {
  margin: 5px 0;
  padding-left: 20px;
}
</style>

